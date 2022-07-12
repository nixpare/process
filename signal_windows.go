package process

import (
	"fmt"
	"os"
	"syscall"
)

func StopProcess(p *os.Process) error {
	err := FreeConsole()
	if err != nil {
		return err
	}

	err = AttachConsole(p.Pid)
	if err != nil {
		return err
	}

	err = GenerateConsoleCtrlEvent()
	if err != nil {
		return err
	}

	return nil
}

func LoadKernel32() (*syscall.DLL, error) {
	d, e := syscall.LoadDLL("kernel32.dll")
	if e != nil {
		return nil, fmt.Errorf("loadDLL: %w", e)
	}

	return d, nil
}

func GenerateConsoleCtrlEvent() error {
	d, e := LoadKernel32()
	if e != nil {
		return e
	}

	p, e := d.FindProc("GenerateConsoleCtrlEvent")
	if e != nil {
		return fmt.Errorf("findProc: %w", e)
	}

	r, _, e := p.Call(syscall.CTRL_C_EVENT, 0)
	if r == 0 {
		return fmt.Errorf("generateConsoleCtrlEvent: %w", e)
	}

	return nil
}

func SetConsoleCtrlHandler(flag bool) error {
	d, e := LoadKernel32()
	if e != nil {
		return e
	}

	p, e := d.FindProc("SetConsoleCtrlHandler")
	if e != nil {
		return fmt.Errorf("findProc: %w", e)
	}

	a := 0
	if flag {
		a = 1
	}

	r, _, e := p.Call(0, uintptr(a))
	if r == 0 {
		return fmt.Errorf("setConsoleCtrlHandler: %w", e)
	}

	return nil
}

func RemoveConsoleCtrlHandler() error {
	return SetConsoleCtrlHandler(true)
}

func RestoreConsoleCtrlHandler() error {
	return SetConsoleCtrlHandler(false)
}

func FreeConsole() error {
	d, e := LoadKernel32()
	if e != nil {
		return e
	}

	p, e := d.FindProc("FreeConsole")
	if e != nil {
		return fmt.Errorf("findProc: %v", e)
	}

	r, _, e := p.Call(0)
	if r == 0 {
		return fmt.Errorf("freeConsole: %w", e)
	}

	return nil
}

func AllocConsole() error {
	d, e := LoadKernel32()
	if e != nil {
		return e
	}

	p, e := d.FindProc("AllocConsole")
	if e != nil {
		return fmt.Errorf("findProc: %v", e)
	}

	r, _, e := p.Call(0)
	if r == 0 {
		return fmt.Errorf("allocConsole: %w", e)
	}

	return nil
}

func AttachConsole(pid int) error {
	d, e := LoadKernel32()
	if e != nil {
		return e
	}

	p, e := d.FindProc("AttachConsole")
	if e != nil {
		return fmt.Errorf("findProc: %v", e)
	}

	r, _, e := p.Call(uintptr(pid))
	if r == 0 {
		return fmt.Errorf("attachConsole: %w", e)
	}

	return nil
}

func SendCtrlC() error {
	if e := RemoveConsoleCtrlHandler(); e != nil {
		return e
	}

	if e := GenerateConsoleCtrlEvent(); e != nil {
		return e
	}

	return nil
}
