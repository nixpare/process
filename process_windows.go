package process

import (
	"fmt"
	"os/exec"
	"syscall"
)

// NewProcess creates a new detached process like with exec.Command function.
// By default the window is hidden and neither Stdin, nor Stdout not Stderr will
// be set
func createCommandFromProcess(p *Process) *exec.Cmd {
	exec := exec.Command(p.execName, p.args...)
	exec.SysProcAttr = &syscall.SysProcAttr{ CreationFlags: 16, HideWindow: true }
	return exec
}

// Reverts the default behaiour and will let the window show when the process starts
func (p *Process) ShowWindow() {
	p.Exec.SysProcAttr.HideWindow = false
}

// Generates a CTRL+C signal (os.Interrupt for managing it in the process) through the
// kill.exe program in the working directory. See the package documentation for more details.
// Stop does not wait the process to exit, use Wait() instead
func (p *Process) stop() error {
	if !p.running {
		return fmt.Errorf("process %s not started", p.name)
	}

	return StopProcess(p.Exec.Process.Pid)
}