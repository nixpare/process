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

 > Windows

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