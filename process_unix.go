//go:build unix
package process

import (
	"os"
	"syscall"
)

func initSysProcAttr() *syscall.SysProcAttr {
	return new(syscall.SysProcAttr)
}

func inheritConsole(spa *syscall.SysProcAttr, flag bool) {
	spa.Setpgid = !flag
}

// stop sends a CTRL+C signal
func (p *Process) stop() error {
	return p.Exec.Process.Signal(os.Interrupt)
}
