//go:build windows

package lock

import "os"

// processAlive reports whether pid identifies a currently-running process
// on Windows. Unlike POSIX, os.FindProcess on Windows actually opens a
// handle to the process and fails if it does not exist, so a successful
// FindProcess (with no further signal needed) is sufficient evidence of
// liveness.
func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// os.FindProcess on Windows opens a real handle; Release it since we
	// only needed it to answer the liveness question, not to hold a
	// reference to the process.
	_ = proc.Release()
	return true
}
