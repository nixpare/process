package process

import (
	"os"
	"os/exec"
	"syscall"
)

// createCommand creates a *cmd.Exec suitable for the platform
func createCommand(execPath string, args ...string) *exec.Cmd {
	cmd := exec.Command(execPath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{ Noctty: true }
	return cmd
}

// stop sends a CTRL+C signal
func (p *Process) stop() error {
	return p.Exec.Process.Signal(os.Interrupt)
}
