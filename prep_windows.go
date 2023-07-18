package process

import (
	"io"
	"os"
)

func (p *Process) prepareStdin(stdin io.Reader) error {
	if stdin != nil {
		if stdin == os.Stdin {
			p.InheritConsole(true)
		}
		p.Exec.Stdin = stdin
		return nil
	}

	var err error
	p.in, err = p.Exec.StdinPipe()
	return err
}

func (p *Process) prepareStdout(stdout io.Writer) error {
	p.captureOut.Reset()

	if stdout != nil {
		if stdout == os.Stdout {
			p.InheritConsole(true)
		}
		p.Exec.Stdout = stdout
		return nil
	}
	
	outPipe, err := p.Exec.StdoutPipe()
	if err != nil {
		return nil
	}
	go func() {
		io.Copy(p.captureOut, outPipe)
	}()
	return nil
}

func (p *Process) prepareStderr(stderr io.Writer) error {
	p.captureOut.Reset()

	if stderr != nil {
		if stderr == os.Stderr {
			p.InheritConsole(true)
		}
		p.Exec.Stderr = stderr
		return nil
	}
	
	errPipe, err := p.Exec.StderrPipe()
	if err != nil {
		return nil
	}
	go func() {
		io.Copy(p.captureErr, errPipe)
	}()
	return nil
}
