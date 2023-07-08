package process

import "fmt"

// ExitStatus holds the status information of a Process
// after it has exited
type ExitStatus struct {
	PID       int
	ExitCode  int
	ExitError error
}

func (exitStatus ExitStatus) Error() error {
	if (exitStatus.ExitCode == 0 || exitStatus.ExitCode == interrupt_errno) && exitStatus.ExitError == nil {
		return nil
	}

	return fmt.Errorf("exit status (code %d): %v", exitStatus.ExitCode, exitStatus.ExitError)
}

func (exitStatus ExitStatus) Unwrap() error {
	return exitStatus.ExitError
}
