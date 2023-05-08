package process

import (
	"fmt"
	"os/exec"
	"syscall"
)

// See https://learn.microsoft.com/en-us/windows/win32/procthread/process-creation-flags
const (
	create_new_console       = 0x00000010
	create_new_process_group = 0x00000200
)

// createCommand creates a *cmd.Exec suitable for the
// platform.
//
// By default, on windows, the process is created in a new console
// (creation flag = 16) hidden, so the process can be easily stopped
// without interfearing with the parent process
func createCommand(execPath string, args ...string) *exec.Cmd {
	exec := exec.Command(execPath, args...)
	exec.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: create_new_console | create_new_process_group,
		HideWindow: true,
	}
	return exec
}

// stop generates a CTRL+C signal
func (p *Process) stop() error {
	if !p.running {
		return fmt.Errorf("process %s not started", p.ExecName)
	}

	return StopProcess(p.Exec.Process.Pid)
}

// ShowWindow reverts the default behaviour and will let the window show up when the process starts
func (p *Process) ShowWindow() {
	p.Exec.SysProcAttr.HideWindow = false
}
