/*
Package Process provides a method to spawn child processes,
like os/exec package, but with one huge difference: on Windows
you have the capabilty of sending an Interrupt signal, that is
you can simulate a CTRL+C event to the child process, which
can then decide whether to handle it or not.

The underlying structure used to spawn and manage the process is
the exec.Cmd object; you have access to it after creating the executable,
but its recommended to access it only reading fields (like accessing the 
ProcessState infos after the process has exited) or to tweek the creation
options before the process starts: for managing redirects and controlling
the process state (Start, Wait, Stop and Kill) its highly recommended
to use the functions provided by the package

The package also gives for Windows users also some usefull functions
to recreate the behaviour via the various functions from the kernel32.dll

OS Compatibility

The package is obviously compatible with all operating systems,
but there are some differences:

# Windows

By default, in windows, the process is spawned in a different console (so has
different stdin, out and err from the parent) and is hidden, but this behaviour
can be reverted with the method Process.ShowWindow() by calling it before the start
When starting a process in windows, depending on what you provide, you can
experience different behaviours:
For the stdin argument:
 + if you provide nil, the child will use the standard input of the new console,
   you can't send any data to the child
 + if you provide process.DevNull(), the pipe will be created and usable, the child
   will not use the standard input of the console and will only receive data through
   the pipe by the parent
 + for any other value, the pipe will be created and usable, and any data inside the
   io.Reader provided will be passed through the pipe. This means that if you pass
   os.Stdin, you can send data from the parent via the methods of the pipe and via
   your console natively
For the stdout and stderr argument:
 + if you provide nil, the child will use the standard output / error of the new console,
   you can't capture any data
 + if you provide process.DevNull(), you will be able to access any data sent by the child,
   but only through the process methods (no real standard output / error)
 + for any other value, the pipe will be created and usable, and any data inside the
   io.Reader provided will be passed through the pipe. This means that if you pass
   os.Stdin, you can send data from the parent via the methods of the pipe and via
   your console natively

On windows, in order to make it work, you must have installed the
kill.exe executable that you can find in my repository at
https://github.com/nixpare/kill: this executable must be
installed in the working folder with the name "kill.exe".
The only one caviat is that the process is not spawned as a child, so
if the calling program crashes the spawned process will remain active
(unless it terminates by itself).
Working on a manager that checks if the main process is still active, keep
track of the spawned processes and kills them if the main process goes off

 > Non Windows OSes

Considering its natively supported to send signals between processes,
there is full support without any external program
*/
package process