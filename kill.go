//go:build ignore
package process

import (
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		os.Exit(1)
	}

	pid, err := strconv.Atoi(os.Args[1])
	if err != nil {
		os.Exit(1)
	}

	p, err := os.FindProcess(pid)
	if err != nil {
		os.Exit(1)
	}

	err = StopProcess(p)
	if err != nil {
		os.Exit(1)
	}
}
