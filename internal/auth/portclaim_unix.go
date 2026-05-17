//go:build !windows

package auth

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// Unix wiring for the cross-platform function-variables in portclaim.go.
// macOS and Linux both ship `lsof` and `ps` in the base install, and both
// support POSIX signals through syscall.Kill — so the same implementation
// covers both platforms.
//
// The Windows counterpart in portclaim_windows.go is a no-op set; callers
// should treat takeover as unsupported there.
func init() {
	findPortHolder = findPortHolderViaLsof
	commandForPID = commandForPIDViaPS
	signalProcess = signalProcessReal
	processExists = processExistsReal
}

// findPortHolderViaLsof returns the PID + command of the first process
// LISTENing on the given TCP port (loopback or otherwise). Returns (nil, nil)
// when no process holds the port — that's a normal pre-bind state, not an
// error. Returns an error only when lsof itself errors in an unexpected way
// (missing binary, malformed output).
func findPortHolderViaLsof(port int) (*PortHolder, error) {
	// `lsof -ti tcp:PORT -sTCP:LISTEN` prints just PIDs, one per line, exit
	// code 1 when nothing matches (treated as "no holder").
	cmd := exec.Command("lsof", "-ti", fmt.Sprintf("tcp:%d", port), "-sTCP:LISTEN") //nolint:gosec // fixed args, integer port — no shell interpolation
	out, err := cmd.Output()
	if err != nil {
		// Exit code 1 with empty stdout = "no holder" — normal.
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 && len(out) == 0 {
			return nil, nil
		}
		return nil, fmt.Errorf("lsof failed: %w", err)
	}
	first := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
	if first == "" {
		return nil, nil
	}
	pid, err := strconv.Atoi(strings.TrimSpace(first))
	if err != nil {
		return nil, fmt.Errorf("lsof returned non-numeric pid %q: %w", first, err)
	}
	command, _ := commandForPID(pid) // best-effort; empty string is fine
	return &PortHolder{PID: pid, Command: command}, nil
}

// commandForPIDViaPS returns the short command name for a PID via
// `ps -o comm= -p PID`. Empty string + nil error when the PID does not exist
// (caller treats empty as "unknown"); non-nil error only on unexpected ps
// failures.
func commandForPIDViaPS(pid int) (string, error) {
	cmd := exec.Command("ps", "-o", "comm=", "-p", strconv.Itoa(pid)) //nolint:gosec // fixed flags, integer pid
	out, err := cmd.Output()
	if err != nil {
		// ps returns non-zero when the PID doesn't exist; treat as "unknown".
		if _, ok := err.(*exec.ExitError); ok {
			return "", nil
		}
		return "", fmt.Errorf("ps failed: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// signalProcessReal sends sig to pid. Wraps syscall.Kill so tests can mock it.
func signalProcessReal(pid int, sig syscall.Signal) error {
	return syscall.Kill(pid, sig)
}

// processExistsReal returns true if a process with the given PID is currently
// alive. Implemented via `kill -0`, which is portable across macOS and Linux.
func processExistsReal(pid int) bool {
	return syscall.Kill(pid, syscall.Signal(0)) == nil
}
