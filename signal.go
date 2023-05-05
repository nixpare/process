package process

import (
	"os"
)

func StopProcess(PID int) error {
	p, err := os.FindProcess(PID)
	if err != nil {
		return err
	}

	return stopProcess(p)
}
