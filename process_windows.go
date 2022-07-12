package process

import (
	"fmt"
	"os/exec"
	"syscall"
)

// NewProcess creates a new detached process like with exec.Command function.
// By default the window is hidden and neither Stdin, nor Stdout not Stderr will
// be set
func NewProcess(name string, args ...string) *Process {
	p := new(Process)

	p.Cmd = exec.Command(name, args...)
	p.Cmd.SysProcAttr = &syscall.SysProcAttr{ CreationFlags: 16, HideWindow: true }

	return p
}

// Reverts the default behaiour and will let the window show when the process starts
func (p *Process) ShowWindow() {
	p.Cmd.SysProcAttr.HideWindow = false
}

// Generates a CTRL+C signal (os.Interrupt for managing it in the process) through the
// kill.exe program in the working directory. See the package documentation for more details.
// Stop does not wait the process to exit, use Wait() instead
func (p *Process) Stop() (err error) {
	if !p.started {
		return ProcessNotStartedErr
	}

	cmd := exec.Command("./kill.exe", fmt.Sprint(p.Cmd.Process.Pid))
	cmd.Run()

	if cmd.ProcessState.ExitCode() == -1 {
		return StopProcessErr
	}

	return
}