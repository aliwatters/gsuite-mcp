//go:build windows

package auth

import "syscall"

// Windows wiring for the cross-platform function-variables in portclaim.go.
// The takeover feature relies on lsof, ps, and POSIX signals — none of which
// exist on Windows in the same form. Rather than blocking the Windows build
// (which gsuite-mcp ships for windows_amd64 and windows_arm64), we wire stub
// implementations that report "no holder" + reject any kill request. The
// effective behaviour is: `gsuite-mcp auth` on Windows falls back to the
// original EADDRINUSE failure, which is no worse than before this fix.
//
// If Windows takeover is ever wanted, the place to add it is here — likely
// via golang.org/x/sys/windows.GetExtendedTcpTable to enumerate listeners and
// OpenProcess/TerminateProcess to manage them.
func init() {
	findPortHolder = func(port int) (*PortHolder, error) { return nil, nil }
	commandForPID = func(pid int) (string, error) { return "", nil }
	signalProcess = func(pid int, sig syscall.Signal) error {
		return syscall.EWINDOWS // "not supported by windows"
	}
	processExists = func(pid int) bool { return false }
}
