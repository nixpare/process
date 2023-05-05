package process

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

// createCommand creates a *cmd.Exec suitable for the
// platform.
//
// By default, on windows, the process is created in a new console
// (creation flag = 16) hidden, so the process can be easily stopped
// without interfearing with the parent process
func createCommand(execPath string, args ...string) *exec.Cmd {
	exec := exec.Command(execPath, args...)
	exec.SysProcAttr = &syscall.SysProcAttr{ CreationFlags: 16, HideWindow: true }
	return exec
}

// stop generates a CTRL+C signal
func (p *Process) stop() error {
	if !p.running {
		return fmt.Errorf("process %s not started", p.execName)
	}

	return StopProcess(p.Exec.Process.Pid)
}

// ShowWindow reverts the default behaviour and will let the window show up when the process starts
func (p *Process) ShowWindow() {
	p.Exec.SysProcAttr.HideWindow = false
}

func (p *Process) prepareStdin(stdin io.Reader) (err error) {
	// is stdin is nil, the child will use the stdin from
	// the new console spawned for him, no pipe supported
	if stdin == nil {
		p.Exec.Stdin = os.Stdin
		return
	}

	// the pipe is created and set
	p.in, err = p.Exec.StdinPipe()
	if err != nil {
		return
	}

	// is stdin == dev_null, the child will have interaction
	// with the parent only through the pipe
	if stdin == dev_null {
		return
	}

	// everything inside stdin will be sent to the child through
	// the pipe (even os.Stdin of the parent)
	go io.Copy(p.in, stdin)
	return
}

func (p *Process) prepareStdout(stdout io.Writer) error {
	// the capture buffer is cleared
	p.captureOut.Reset()

	// is stdout is nil, the child will use the stdout from
	// the new console spawned for him, no capture supported
	if stdout == nil {
		p.Exec.Stdout = os.Stdout
		return nil
	}
	
	// if stdout == dev_null, the child will write everything
	// only to the buffer
	if stdout == dev_null {
		p.Exec.Stdout = p.captureOut
		return nil
	}
	
	// we create the pipe, so that we can later write both
	// to the capture buffer and the stdout (even os.Stdout
	// of the parent)
	outPipe, err := p.Exec.StdoutPipe()
	if err != nil {
		return nil
	}
	go func() {
		sc := bufio.NewScanner(outPipe)
		sc.Split(bufio.ScanBytes)

		for sc.Scan() {
			b := sc.Bytes()

			p.captureOut.Write(b)
			stdout.Write(b)
		}
	}()
	return nil
}

func (p *Process) prepareStderr(stderr io.Writer) error {
	p.captureErr.Reset()
	
	if stderr == nil {
		p.Exec.Stderr = os.Stderr
		return nil
	}

	if stderr == dev_null {
		p.Exec.Stderr = p.captureErr
		return nil
	}
	
	errPipe, err := p.Exec.StderrPipe()
	if err != nil {
		return nil
	}
	go func() {
		sc := bufio.NewScanner(errPipe)
		sc.Split(bufio.ScanBytes)

		for sc.Scan() {
			b := sc.Bytes()

			p.captureErr.Write(b)
			stderr.Write(b)
		}
	}()
	return nil
}
