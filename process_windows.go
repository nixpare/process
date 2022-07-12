package process

import (
	"fmt"
	"os/exec"
	"syscall"
)

func NewProcess(name string, args ...string) *Process {
	p := new(Process)

	p.Cmd = exec.Command(name, args...)
	p.Cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: 16}

	return p
}

func (p *Process) HideWindow() {
	p.Cmd.SysProcAttr.HideWindow = true
}

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