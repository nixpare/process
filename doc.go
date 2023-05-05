/*
Package Process provides a method to spawn child processes,
like os/exec package, but with one huge difference: on Windows
you have the capabilty of sending an Interrupt signal, that is
you can simulate a CTRL+C event to the child process, which
can then decide whether to handle it or not.

The underlying structure used to spawn and manage the process is
the exec.Cmd: for managing input, output and lifecycle of the process,
its highly recommended to use the functions provided by the package

The package also gives for Windows users some usefull functions
to recreate the behaviour via the various functions from the kernel32.dll

# OS Compatibility

The package is obviously compatible with all operating systems,
but there are some differences:

> Non Windows OSes

Considering its natively supported to send signals between processes,
there is full support without any external program

> Windows

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

This process might still fail, the reasons are still not known. For example, if the child
process you want to stop is created by the "go run" command, this will not work
*/
package process