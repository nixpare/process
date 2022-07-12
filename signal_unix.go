//go:build !windows
package process

import "os"

// Sends a os.Interrupt signal to the process
func StopProcess(p *os.Process) error {
	return p.Signal(os.Interrupt)
}