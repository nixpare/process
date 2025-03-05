package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/nixpare/process"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "At least one argument needed")
		fmt.Println(help())
		os.Exit(1)
	}

	if os.Args[1] == "--help" {
		fmt.Println(help())
		os.Exit(0)
	}

	token, err := process.GetLowerPrivilegeToken()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to generate lower privilege token: %v\n", errors.Unwrap(err))
		os.Exit(1)
	}

	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	cmd.SysProcAttr.Token = token

	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "An error has occurred during program \"%s\" execution: %v\n", os.Args[1], err)
		os.Exit(1)
	}
}

func help() string {
	return `Usage: drop <exec_path> [ args ... ]`
}
