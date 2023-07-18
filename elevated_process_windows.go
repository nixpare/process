package process

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const sudo_command = "--nix-sudo"

func init() {
	if len(os.Args) < 4 || os.Args[1] != sudo_command || !AmAdmin() {
		return
	}

	os.Exit(elevatedRun())
}

func NewElevatedProcess(wd string, execPath string, args ...string) (*Process, error) {
	if AmAdmin() {
		return NewProcess(wd, execPath, args...)
	}
	
	elevatedCommand := fmt.Sprintf(
		`'Start-Process -FilePath "%s" -ArgumentList "%s %d %s" -Verb RunAs -WindowStyle hidden -Wait'`,
		os.Args[0],
		sudo_command,
		os.Getpid(),
		strings.Join(append([]string{execPath}, args...), " "),
	)
	return NewProcess(wd, "powershell.exe", ParseCommandArgs("-NoProfile", "-Command", elevatedCommand)...)
}

func AmAdmin() bool {
	_, err := os.Open(`\\.\PHYSICALDRIVE0`)
	return err == nil
}

func elevatedRun() int {
	ppid, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	os.Args = append([]string{os.Args[0]}, os.Args[3:]...)

	err = ChangeConsole(ppid)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	p, err := NewProcess("", os.Args[1], os.Args[2:]...)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	err = p.Start(os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	exitStatus := p.Wait()
	return exitStatus.ExitCode
}

type StdHandle int

const (
	STD_INPUT_HANDLE StdHandle = -10
	STD_OUTPUT_HANDLE StdHandle = -11
	STD_ERROR_HANDLE StdHandle = -12
)

func GetStdHandle(stdType StdHandle) (uintptr, error) {
	d, e := LoadKernel32()
	if e != nil {
		return 0, e
	}

	p, e := d.FindProc("GetStdHandle")
	if e != nil {
		return 0, fmt.Errorf("findProc: %v", e)
	}

	r, _, e := p.Call(uintptr(stdType))
	if r == 0 {
		return 0, fmt.Errorf("getStdHandle: %w", e)
	}

	return r, nil
}

func ChangeConsole(pid int) error {
	err := FreeConsole()
	if err != nil {
		return err
	}

	err = AttachConsole(pid)
	if err != nil {
		return err
	}

	inHandle, err := GetStdHandle(STD_INPUT_HANDLE)
	if err != nil {
		return err
	}

	outHandle, err := GetStdHandle(STD_OUTPUT_HANDLE)
	if err != nil {
		return err
	}

	errHandle, err := GetStdHandle(STD_ERROR_HANDLE)
	if err != nil {
		return err
	}

	*os.Stdin = *os.NewFile(inHandle, "stdin")
	*os.Stdout = *os.NewFile(outHandle, "stdout")
	*os.Stderr = *os.NewFile(errHandle, "stderr")

	return nil
}
