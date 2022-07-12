package process

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
)

type constError string

func (err constError) Error() string {
    return string(err)
}

const (
	ProcessAlreadyStartedErr = constError("process already started")
	ProcessNotStartedErr = constError("process not started")
	StopProcessErr = constError("stop process failed")
	ProcessNotRunningErr = constError("process not started")
)

// Process wraps the exec.Cmd structure and provides more control over the
// Stdin, Stdout, Stderr and gracefull termination by sending a CTRL+C signal
// even on Windows (see package documentation for more details)
type Process struct {
	Cmd *exec.Cmd
	InPipe  io.WriteCloser
	OutPipe io.ReadCloser
	ErrPipe io.ReadCloser
	started bool
	running bool
	startMutex sync.Mutex
	waitMutex sync.Mutex
	exitError error
}

// Pipes the Stdin to f. Even though you could use os.Stdin
// as the pipe destination, this will not work, use RedirectStdin
// method instead
func (p *Process) PipeStdin(f io.Reader) {
	p.Cmd.Stdin = f
}

// Pipes the Stdout to f. Even though you could use os.Stdout
// as the pipe destination, this will not work, use RedirectStdout
// method instead
func (p *Process) PipeStdout(f io.Writer) {
	p.Cmd.Stdout = f
}

// Pipes the Stderr to f. Even though you could use os.Stderr
// as the pipe destination, this will not work, use RedirectStderr
// method instead
func (p *Process) PipeStderr(f io.Writer) {
	p.Cmd.Stderr = f
}

// Redirects the os.Stdin to the child process. This must be called before
// the process starts; consider that this function is non-blocking, so you don't
// need to call it in a goroutine
func (p *Process) RedirectStdin() (err error) {
	p.InPipe, err = p.Cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}

	go io.Copy(p.InPipe, os.Stdin)
	return nil
}

// Redirects the process Stdout to os.Stdout. This must be called before
// the process starts; consider that this function is non-blocking, so you don't
// need to call it in a goroutine
func (p *Process) RedirectStout() (err error) {
	p.OutPipe, err = p.Cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

	go io.Copy(os.Stdout, p.OutPipe)
	return nil
}

// Redirects the process Stderr to os.Stderr. This must be called before
// the process starts; consider that this function is non-blocking, so you don't
// need to call it in a goroutine
func (p *Process) RedirectSterr() (err error) {
	p.ErrPipe, err = p.Cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	go io.Copy(os.Stderr, p.ErrPipe)
	return nil
}

// Redirects all Stdin, Stdout, Stderr from the child process to the calling process.
// This function is non-blocking, so you don't need to call it in a goroutine
func (p *Process) RedirectAll() (err error) {
	err = p.RedirectStdin()
	if err != nil {
		return
	}
	err = p.RedirectStout()
	if err != nil {
		return
	}
	err = p.RedirectSterr()
	if err != nil {
		return
	}

	return
}

// Starts the process like exec.Cmd.Start method
func (p *Process) Start() (err error) {
	if p.started {
		return ProcessAlreadyStartedErr
	}

	p.startMutex.Lock()
	p.started = true
	err = p.Cmd.Start()

	if err != nil {
		p.startMutex.Unlock()
		return err
	}

	go func() {
		p.running = true
		p.startMutex.Unlock()
		
		p.waitMutex.Lock()
		p.exitError = p.Cmd.Wait()
		p.waitMutex.Unlock()

		p.release()
	}()

	return
}

// Waits the process to exit, either after a Stop call (graceful)
// or after it's killed
func (p *Process) Wait() (err error) {
	if !p.started {
		return ProcessNotStartedErr
	}

	p.startMutex.Lock()
	if !p.running {
		p.startMutex.Unlock()
		return ProcessNotRunningErr
	}
	p.startMutex.Unlock()

	p.waitMutex.Lock()
	err = p.exitError
	p.waitMutex.Unlock()

	return
}

func (p *Process) release() {
	p.running = false

	if p.InPipe != nil {
		p.InPipe.Close()
	}
	if p.OutPipe != nil {
		p.OutPipe.Close()
	}
	if p.ErrPipe != nil {
		p.ErrPipe.Close()
	}
}

// Kills the process
func (p *Process) Kill() (err error) {
	if !p.started {
		return ProcessNotStartedErr
	}

	if !p.running {
		return ProcessNotRunningErr
	}

	return p.Cmd.Process.Kill()
}

// Tells whether the process is running
func (p *Process) IsRunning() bool {
	return p.running
}
