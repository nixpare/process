package process

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const runElevateCommand = "--run-me-elevated"

const (
	no_error = iota
	pid_conv_error
	invalid_pid_error
	get_curr_tok_error
	dup_token_error
	open_proc_error
	get_tok_info_error
	set_tok_error
	create_proc_snap
	read_thread_info
	open_proc_thread
	resume_error
)

func init() {
	if len(os.Args) < 3 || os.Args[1] != runElevateCommand {
		return
	}

	os.Exit(tokenSwitch())
}

// See https://learn.microsoft.com/en-us/windows/win32/procthread/process-creation-flags
const (
	create_new_console       uint32 = 0x00000010
	create_new_process_group uint32 = 0x00000200
)

func initSysProcAttr() *syscall.SysProcAttr {
	spa := new(syscall.SysProcAttr)

	inheritConsole(spa, false)
	showWindow(spa, false)
	
	return spa
}

// stop generates a CTRL+C signal
func (p *Process) stop() error {
	if !p.running {
		return nil
	}

	return StopProcess(p.Exec.Process.Pid)
}

func showWindow(spa *syscall.SysProcAttr, flag bool) {
	spa.HideWindow = !flag
}

func (p *Process) ShowWindow(flag bool) {
	showWindow(p.SysProcAttr, flag)
}

func inheritConsole(spa *syscall.SysProcAttr, flag bool) {
	if flag {
		spa.CreationFlags = spa.CreationFlags & ^(create_new_console | create_new_process_group)
	} else {
		spa.CreationFlags |= create_new_console | create_new_process_group
	}
}

func AmAdmin() bool {
	return windows.GetCurrentProcessToken().IsElevated()
}

func (p *Process) StartElevated(stdin io.Reader, stdout, stderr io.Writer) error {
	if AmAdmin() {
		return p.Start(stdin, stdout, stderr)
	}

	p.SysProcAttr.CreationFlags |= windows.CREATE_SUSPENDED

	err := p.Start(stdin, stdout, stderr)
	if err != nil {
		return err
	}

	cwd, _ := windows.Getwd()
	err = windows.ShellExecute(
		0, windows.StringToUTF16Ptr("runas"),
		windows.StringToUTF16Ptr(os.Args[0]), windows.StringToUTF16Ptr(fmt.Sprintf("%s %d", runElevateCommand, p.PID())),
		windows.StringToUTF16Ptr(cwd), windows.SW_HIDE,
	)
	if err != nil {
		p.Kill()
		return fmt.Errorf("UAC error: %w", err)
	}

	return nil
}

func (p *Process) RunElevated(stdin io.Reader, stdout, stderr io.Writer) (exitStatus ExitStatus, err error) {
	if AmAdmin() {
		return p.Run(stdin, stdout, stderr)
	}

	err = p.StartElevated(stdin, stdout, stderr)
	if err != nil {
		return
	}

	exitStatus = p.Wait()
	err = exitStatus.Error()
	return
}

func tokenSwitch() (exitCode int) {
	pid, err := strconv.Atoi(os.Args[2])
	if err != nil {
		return pid_conv_error
	}
	if pid < 0 {
		return invalid_pid_error
	}

	var tok windows.Token
	err = windows.OpenProcessToken(windows.CurrentProcess(), windows.MAXIMUM_ALLOWED, &tok)
	if err != nil {
		return get_curr_tok_error
	}

	var newTok windows.Token
	err = windows.DuplicateTokenEx(
		tok, windows.MAXIMUM_ALLOWED,
		nil, windows.SecurityImpersonation,
		windows.TokenPrimary, &newTok,
	)
	if err != nil {
		return dup_token_error
	}
	defer newTok.Close()

	proc, err := windows.OpenProcess(
		windows.PROCESS_SET_INFORMATION,
		false,
		uint32(pid),
	)
	if err != nil {
		return open_proc_error
	}
	defer windows.CloseHandle(proc)

	var info struct {
		Token  windows.Handle
		Thread windows.Handle
	}

	info.Token = windows.Handle(newTok)
	info.Thread = windows.Handle(0)
	err = windows.NtSetInformationProcess(proc, windows.ProcessAccessToken, unsafe.Pointer(&info), uint32(unsafe.Sizeof(info)))
	if err != nil {
		return set_tok_error
	}

	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPTHREAD, 0)
	if err != nil {
		return create_proc_snap
	}
	defer windows.CloseHandle(snap)

	var threadEntry windows.ThreadEntry32
	threadEntry.Size = uint32(unsafe.Sizeof(threadEntry))

	for err = windows.Thread32First(snap, &threadEntry); err == nil; err = windows.Thread32Next(snap, &threadEntry) {
		if threadEntry.OwnerProcessID == uint32(pid) {
			break
		}
	}
	if err != nil {
		return read_thread_info
	}

	thread, err := windows.OpenThread(windows.THREAD_SUSPEND_RESUME, false, threadEntry.ThreadID)
	if err != nil {
		return open_proc_thread
	}
	defer windows.CloseHandle(thread)

	_, err = windows.ResumeThread(thread)
	if err != nil {
		return resume_error
	}

	return
}
