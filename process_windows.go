package process

import (
	"os/exec"
	"syscall"
)

const killProcessName = "./kill.exe"

func NewProcess(name string, args ...string) *Process {
	p := new(Process)

	p.Cmd = exec.Command(name, args...)
	p.Cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 16}

	return p
}

func (p *Process) HideWindow() {
	p.Cmd.SysProcAttr.HideWindow = true
}