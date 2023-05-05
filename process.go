package process

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/nixpare/comms"
)

type ExitStatus struct {
	PID       int
	ExitCode  int
	ExitError error
}

// Process wraps the default *exec.Cmd structure and makes easier the
// access to redirect the standard output and check when it terminates.
// Another limitation is that graceful shutdown is not implemented yet
// due to Windows limitations, but will be. It's possible to wait for its
// termination on multiple goroutines by waiting for exitC closure. Both
// in and out can be nil
type Process struct {
	name             string
	wd               string
	execName         string
	args             []string
	exitComm         *comms.Broadcaster[ExitStatus]
	Exec             *exec.Cmd
	running          bool
	lastExitStatus   ExitStatus
	in               io.WriteCloser
	out              io.ReadCloser
	captureOut       *bytes.Buffer
	err              io.ReadCloser
	captureErr       *bytes.Buffer
}

// NewProcess creates a new program with the diven parameters
func NewProcess(name, wd string, execName string, args ...string) (*Process, error) {
	wd, err := filepath.Abs(wd)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(wd)
	if err != nil {
		return nil, fmt.Errorf("directory \"%s\" not found", wd)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("\"%s\" is not a directory", wd)
	}

	execName, err = filepath.Abs(execName)
	if err != nil {
		return nil, err
	}

	return &Process{
		name:     name,
		wd:       wd,
		execName: execName,
		args:     args,
		exitComm: comms.NewBroadcaster[ExitStatus](),
	}, nil
}

// Start starts the process and with a goroutine waits for its
// termination. It returns an error if there is a problem with
// the creation of the new process, but if something happens during
// the execution it will be reported to che channel returned
func (p *Process) Start(stdin io.Reader, stdout, stderr io.Writer) error {
	if p.IsRunning() {
		return fmt.Errorf("process \"%s\" is already running", p.name)
	}

	p.Exec = createCommandFromProcess(p)
	err := p.preparePipesAndFiles(stdin, stdout, stderr)
	if err != nil {
		return fmt.Errorf("process \"%s\" pipe error: %w", p.name, err)
	}

	err = p.Exec.Start()
	if err != nil {
		return fmt.Errorf("process \"%s\" startup error: %w", p.name, err)
	}

	p.running = true
	go p.afterStart()

	return nil
}

func (p *Process) preparePipesAndFiles(stdin io.Reader, stdout, stderr io.Writer) (err error) {
	// STDIN
	p.in, err = p.Exec.StdinPipe()
	if err != nil {
		return
	}
	if stdin != nil {
		go io.Copy(p.in, stdin)
	}

	// STDOUT
	p.out, err = p.Exec.StdoutPipe()
	if err != nil {
		return
	}
	p.captureOut = bytes.NewBuffer(nil)
	go func() {
		sc := bufio.NewScanner(p.out)
		sc.Split(bufio.ScanBytes)
		for sc.Scan() {
			b := sc.Bytes()
			p.captureOut.Write(b)
			if stdout != nil {
				stdout.Write(b)
			}
		}
	}()

	// STDERR
	p.err, err = p.Exec.StderrPipe()
	if err != nil {
		return
	}
	p.captureErr = bytes.NewBuffer(nil)
	go func() {
		sc := bufio.NewScanner(p.err)
		sc.Split(bufio.ScanBytes)
		for sc.Scan() {
			b := sc.Bytes()
			p.captureErr.Write(b)
			if stderr != nil {
				stderr.Write(b)
			}
		}
	}()
	
	return
}

// afterStart waits for the process with the already provided function by *os.Process,
// then closes the exitC channel to segnal its termination
func (p *Process) afterStart() {
	err := p.Exec.Wait()
	p.lastExitStatus = ExitStatus{
		PID: p.Exec.Process.Pid,
		ExitCode: p.Exec.ProcessState.ExitCode(),
		ExitError: err,
	}

	p.exitComm.Send(p.lastExitStatus)
	p.running = false
}

// wait waits for the process termination (if running) and returns the last process
// state known
func (p *Process) Wait() ExitStatus {
	if p.IsRunning() {
		return p.exitComm.Get()
	}
	return p.lastExitStatus
}

func (p *Process) Stop() error {
	return p.stop()
}

// Kill forcibly kills the process and waits for the cleanup
func (p *Process) Kill() error {
	if !p.IsRunning() {
		return fmt.Errorf("program \"%s\" is already stopped", p.name)
	}

	err := p.Exec.Process.Kill()
	if err != nil {
		return fmt.Errorf("program \"%s\" kill error: %w", p.name, err)
	}

	return nil
}

func (p *Process) SendInput(data []byte) error {
	if !p.IsRunning() {
		return fmt.Errorf("program \"%s\" is not running", p.name)
	}

	_, err := p.in.Write(data)
	return err
}

func (p *Process) SendText(text string) error {
	return p.SendInput(append([]byte(text), '\n'))
}

func (p *Process) CloseInput() error {
	return p.in.Close()
}

func (p *Process) StdOut() []byte {
	return p.captureOut.Bytes()
}

func (p *Process) LastOutputLines(n int) string {
	if n <= 0 {
		return ""
	}

	output := string(p.StdOut())
	outputSplit := strings.Split(output, "\n")
	
	return strings.Join(outputSplit[len(outputSplit) - n:], "\n")
}

func (p *Process) StdErr() []byte {
	return p.captureErr.Bytes()
}

// IsRunning reports whether the program is running
func (p *Process) IsRunning() bool {
	return p.running
}

func (p *Process) String() string {
	var state string
	if p.IsRunning() {
		state = fmt.Sprintf("Running - %d", p.Exec.Process.Pid)
	} else {
		state = "Stopped"
	}
	return fmt.Sprintf("%s (%s)", p.name, state)
}
