//go:build !windows
package process

import "os"

func StopProcess(p *os.Process) error {
	return p.Signal(os.Interrupt)
}