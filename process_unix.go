//go:build !windows
package process

import "os/exec"

const killProcessName = "./kill"

func NewProcess(name string, args ...string) *Process {
	p := new(Process)
	p.Cmd = exec.Command(name, args...)

	return p
}