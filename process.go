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

var err error

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

func (p *Process) PipeStdin(f io.Reader) {
	p.Cmd.Stdin = f
}

func (p *Process) PipeStdout(f io.Writer) {
	p.Cmd.Stdout = f
}

func (p *Process) PipeStderr(f io.Writer) {
	p.Cmd.Stderr = f
}

func (p *Process) RedirectStdin() error {
	p.InPipe, err = p.Cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}

	go io.Copy(p.InPipe, os.Stdin)
	return nil
}

func (p *Process) RedirectStout() error {
	p.OutPipe, err = p.Cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

	go io.Copy(os.Stdout, p.OutPipe)
	return nil
}

func (p *Process) RedirectSterr() error {
	p.ErrPipe, err = p.Cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	go io.Copy(os.Stderr, p.ErrPipe)
	return nil
}

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

func (p *Process) Kill() (err error) {
	if !p.started {
		return ProcessNotStartedErr
	}

	if !p.running {
		return ProcessNotRunningErr
	}

	return p.Cmd.Process.Kill()
}

func (p *Process) IsRunning() bool {
	return p.running
}
