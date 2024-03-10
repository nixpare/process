package process

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/nixpare/comms"
)

// Process wraps the default *exec.Cmd structure and makes easier to
// access and redirect the standard input, output and error. It also
// allows to gracefully stop a process both in Windows and UNIX-like
// OSes by generating a CTRL-C event without stopping the parent process.
//
// For more details, see the package documentation
type Process struct {
	ExecName       string
	execPath       string
	args           []string
	wd             string
	Env            []string
	SysProcAttr    syscall.SysProcAttr
	Exec           *exec.Cmd
	exitComm       *comms.Broadcaster[ExitStatus]
	running        bool
	lastExitStatus ExitStatus
	in             io.WriteCloser
	outC           chan []byte
	captureOut     [][]byte
	errC           chan []byte
	captureErr     [][]byte
}

// NewProcess creates a new Process with the given arguments.
//
// The Process, once started, will run on the given working
// directory, but the executable path can be a relative path,
// calculated from the parent working directory, not the child
// one
func NewProcess(wd string, execPath string, args ...string) (*Process, error) {
	execName := execPath

	if filepath.Base(execPath) == execPath {
		lp, err := exec.LookPath(execPath)
		if lp != "" {
			execPath = lp
		} else if err != nil {
			return nil, err
		}
	}

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

	p := &Process{
		ExecName:    execName,
		execPath:    execPath,
		args:        args,
		wd:          wd,
		SysProcAttr: initSysProcAttr(),
		exitComm:    comms.NewBroadcaster[ExitStatus](),
	}

	return p, nil
}

func (p *Process) Run(stdin io.Reader, stdout, stderr io.Writer) (exitStatus ExitStatus, err error) {
	err = p.Start(stdin, stdout, stderr)
	if err != nil {
		return
	}

	exitStatus = p.Wait()
	err = exitStatus.Error()
	return
}

// Start starts the Process. It returns an error if there is a problem with
// the creation of the new Process, but if something happens during
// the execution it will be reported in the ExitStatus provided by calling
// the Wait method
func (p *Process) Start(stdin io.Reader, stdout, stderr io.Writer) error {
	if p.IsRunning() {
		return fmt.Errorf("process \"%s\" is already running", p.ExecName)
	}

	p.initCommand()

	p.outC = make(chan []byte, 10)
	p.captureOut = nil

	p.errC = make(chan []byte, 10)
	p.captureErr = nil

	err := p.preparePipes(stdin, stdout, stderr)
	if err != nil {
		return fmt.Errorf("process \"%s\" pipe error: %w", p.ExecName, err)
	}

	err = p.Exec.Start()
	if err != nil {
		return fmt.Errorf("process \"%s\" startup error: %w", p.ExecName, err)
	}

	p.running = true
	go p.afterStart()

	return nil
}

func (p *Process) initCommand() {
	p.Exec = exec.Command(p.execPath, p.args...)
	p.Exec.Dir = p.wd
	p.Exec.Env = p.Env
	
	p.Exec.SysProcAttr = new(syscall.SysProcAttr)
	*p.Exec.SysProcAttr = p.SysProcAttr
}

func (p *Process) preparePipes(stdin io.Reader, stdout, stderr io.Writer) error {
	err := p.prepareStdin(stdin)
	if err != nil {
		return err
	}

	err = p.prepareStdout(stdout)
	if err != nil {
		return err
	}

	return p.prepareStderr(stderr)
}

// afterStart waits for the Process with the already provided function by *os.Process,
// then sends the ExitStatus via the broadcaster
func (p *Process) afterStart() {
	err := p.Exec.Wait()

	p.lastExitStatus = ExitStatus{
		PID:       p.Exec.Process.Pid,
		ExitCode:  p.Exec.ProcessState.ExitCode(),
		ExitError: err,
	}

	p.exitComm.Send(p.lastExitStatus)
	p.running = false
}

// Wait waits for the Process termination (if running) and returns the last Process
// state known
func (p *Process) Wait() ExitStatus {
	if p.IsRunning() {
		return p.exitComm.Get()
	}
	return p.lastExitStatus
}

// Stop sends a CTRL-C event to the Process to allow a graceful exit
func (p *Process) Stop() error {
	return p.stop()
}

// Kill forcibly kills the Process
func (p *Process) Kill() error {
	if !p.IsRunning() {
		return fmt.Errorf("program \"%s\" is already stopped", p.ExecName)
	}

	err := p.Exec.Process.Kill()
	if err != nil {
		return fmt.Errorf("program \"%s\" kill error: %w", p.ExecName, err)
	}

	return nil
}

// SendInput sends data to the Process via a pipe, if the Process is
// running and can pipe data. The Process might take any input until
// a newline or an EOF: for the first one you can use the SendText method,
// for the second case, you can close the pipe via the CloseInput method
//
// For more details, see the package documentation
func (p *Process) SendInput(data []byte) error {
	if !p.IsRunning() {
		return fmt.Errorf("program \"%s\" is not running", p.ExecName)
	}

	if p.in == nil {
		return errors.New("can't pipe input to the process, see package documentation for more details")
	}

	_, err := p.in.Write(data)
	return err
}

// Sends a text with a newline appended automatically, to
// simulate a real user behind a keyboard
func (p *Process) SendText(text string) error {
	return p.SendInput(append([]byte(text), '\n'))
}

// Closes the input pipe, simulating a CTRL-Z or an EOF
// (if the stdin comes from a file)
func (p *Process) CloseInput() error {
	return p.in.Close()
}

// Stdout returns all the standard output captured at
// the moment until the Process is started again
func (p *Process) Stdout() [][]byte {
	return p.captureOut
}

// Stderr returns all the standard error captured at
// the moment until the Process is started again
func (p *Process) Stderr() [][]byte {
	return p.captureErr
}

// StdoutChan returns a channel connected to the standard
// output of the program in which are sent lines. This
// channel is automatically closed when the program stops,
// so it's safe to range over it and must not be closed manually
func (p *Process) StdoutChan() <- chan []byte {
	return p.outC
}

// StderrChan returns a channel connected to the standard
// error of the program in which are sent lines. This
// channel is automatically closed when the program stops,
// so it's safe to range over it and must not be closed manually
func (p *Process) StderrChan() <- chan []byte {
	return p.errC
}

// IsRunning reports whether the Process is running
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
	return fmt.Sprintf("%s (%s)", p.ExecName, state)
}
