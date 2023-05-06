package process

import (
	"bufio"
	"io"
	"os"
)

func (p *Process) prepareStdin(stdin io.Reader) (err error) {
	if stdin == nil {
		p.Exec.Stdin = os.Stdin
		return
	}

	p.in, err = p.Exec.StdinPipe()
	if err != nil {
		return
	}

	if stdin == dev_null {
		p.Exec.Stdin = dev_null
		return
	}

	go io.Copy(p.in, stdin)
	return
}

func (p *Process) prepareStdout(stdout io.Writer) error {
	p.captureOut.Reset()

	if stdout == nil {
		p.Exec.Stdout = os.Stdout
		return nil
	}
	
	if stdout == dev_null {
		p.Exec.Stdout = p.captureOut
		return nil
	}
	
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
