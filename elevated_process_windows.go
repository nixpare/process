package process

import (
	"os"
)

// If the current context does not have elevate privileges, returns
// a process that will ask for permission, otherwise returns a usual
// process like you called process.NewProcess(). For this to work you
// must have installed a "sudo" program, one that support every type
// of redirection is "gsudo" from https://github.com/gerardog/gsudo.
func NewElevatedProcess(wd string, execPath string, args ...string) (*Process, error) {
	if AmAdmin() {
		return NewProcess(wd, execPath, args...)
	}
	
	return NewProcess(wd, "sudo", append([]string{execPath}, args...)...)
}

// AmAdmin tells whether the current context has elevated privileges
func AmAdmin() bool {
	_, err := os.Open(`\\.\PHYSICALDRIVE0`)
	return err == nil
}
