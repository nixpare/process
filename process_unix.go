//go:build !windows

package process

import (
	"os"
	"os/exec"
)

func NewProcess(name string, args ...string) *Process {
	p := new(Process)
	p.Cmd = exec.Command(name, args...)

	return p
}

func (p *Process) Stop() (err error) {
	return p.Cmd.Process.Signal(os.Interrupt)
}