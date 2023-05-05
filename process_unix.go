//go:build !windows
package process

import (
	"os"
	"os/exec"
)

// NewProcess creates a new child process like with exec.Command function.
// By default neither Stdin, nor Stdout not Stderr will
// be set
func createCommandFromProcess(p *Process) *exec.Cmd {
	return exec.Command(p.execName, p.args...)
}

// Sends a CTRL+C signal (os.Interrupt for managing it) to the process. Stop
// does not wait the process to exit, use Wait() instead
func (p *Process) stop() (err error) {
	return p.Cmd.Process.Signal(os.Interrupt)
}