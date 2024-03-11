package process

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/nixpare/broadcaster"
)

func (p *Process) prepareStdout(stdout io.Writer) error {
	outPipe, err := p.Exec.StdoutPipe()
	if err != nil {
		return nil
	}

	go pipeOutput(p.outBc, &p.captureOut, outPipe, stdout, "stdout")
	return nil
}

func (p *Process) prepareStderr(stderr io.Writer) error {
	errPipe, err := p.Exec.StderrPipe()
	if err != nil {
		return nil
	}

	go pipeOutput(p.errBc, &p.captureErr, errPipe, stderr, "stderr")
	return nil
}

func pipeOutput(bc *broadcaster.Broadcaster[[]byte], capture *[][]byte, r io.ReadCloser, w io.Writer, pipeID string) {
	defer bc.Close()

	pipeR, pipeW := io.Pipe()
	var buf [1024]byte

	br := bufio.NewReader(pipeR)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			line, err := br.ReadBytes('\n')
			if err != nil {
				if !errors.Is(err, io.EOF) {
					line = []byte(fmt.Sprintf("broken %s pipe: %v", pipeID, err))
				}
			}

			if len(line) == 0 {
				if err != nil {
					break
				} else {
					continue
				}
			}

			if len(line) > 0 && line[len(line)-1] == '\n' {
				line = line[:len(line)-1]
			}

			bc.Send(line)
			*capture = append(*capture, line)

			if err != nil {
				break
			}
		}
	}()

	for {
		n, err := r.Read(buf[:])
		b := buf[:n]
		if err != nil {
			if !errors.Is(err, io.EOF) {
				b = append(b, []byte(fmt.Sprintf("broken %s pipe: %v", pipeID, err))...)
			}
		}

		pipeW.Write(b)
		if w != nil && w != dev_null {
			w.Write(b)
		}

		if err != nil {
			pipeW.CloseWithError(err)
			break
		}
	}

	wg.Wait()
}
