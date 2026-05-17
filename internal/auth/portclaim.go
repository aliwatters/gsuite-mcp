package auth

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// PortHolder describes the process currently bound to a port.
type PortHolder struct {
	PID     int
	Command string // typically the short executable name (15-char truncated on Linux)
}

// Process-discovery and signalling are wired through package-level function
// variables so unit tests can override them. The default values are assigned
// from the platform-specific init() in portclaim_unix.go or
// portclaim_windows.go — `syscall.Kill` exists only on Unix, and `lsof`/`ps`
// are not present on Windows, so production wiring lives there.
var (
	findPortHolder func(port int) (*PortHolder, error)
	commandForPID  func(pid int) (string, error)
	signalProcess  func(pid int, sig syscall.Signal) error
	processExists  func(pid int) bool
)

// daemonExecName is the EXACT basename matched against a holder's command
// when classifying via process-name fallback. Substring matching was rejected
// here because a foreign binary whose name happens to contain "gsuite-mcp"
// (e.g. "gsuite-mcp-helper", "my-gsuite-mcp-wrapper") would otherwise be
// flagged as ours and signalled — a real risk per Copilot review of
// gsuite-mcp#159. The kernel TASK_COMM_LEN cap is 15 chars, comfortably
// above the 10-char "gsuite-mcp", so ps -o comm= does not truncate. Path
// prefixes that macOS' ps returns are stripped via filepath.Base before
// comparison.
const daemonExecName = "gsuite-mcp"

// takeoverWaitForExit is how long to wait between SIGTERM and SIGKILL.
const takeoverWaitForExit = 2 * time.Second

// takeoverPollInterval is the polling interval used while waiting for a
// signalled daemon to release the port.
const takeoverPollInterval = 50 * time.Millisecond

// IsOurDaemon reports whether the process with the given PID is a gsuite-mcp
// daemon we (or a prior session) started. Uses two signals:
//
//  1. PRIMARY: the pidfile at PidFilePath(). If readable AND its recorded PID
//     matches the candidate, it's ours. This is the reliable post-reboot
//     path (process names alone can collide with foreign binaries).
//
//  2. FALLBACK: an EXACT match on the candidate's command basename (via
//     `ps -o comm=`, then `filepath.Base`, lowercased). Substring matching
//     was rejected here as too permissive — see daemonExecName for the
//     reasoning. The fallback only fires when the pidfile is missing or
//     stale (e.g. older daemon predating this fix, or manual cleanup).
//
// The function never kills anything — it only classifies. The caller decides
// whether to act on the classification (see TakeOverPort).
func IsOurDaemon(pid int) bool {
	if pid <= 0 {
		return false
	}
	if pidFilePID, err := ReadPidFile(); err == nil && pidFilePID == pid {
		return true
	}
	cmdName, err := commandForPID(pid)
	if err != nil || cmdName == "" {
		return false
	}
	base := strings.ToLower(filepath.Base(strings.TrimSpace(cmdName)))
	return base == daemonExecName
}

// TakeOverPort is called when a bind to `port` fails with EADDRINUSE. It
// inspects the holder; if it's our own daemon, it sends SIGTERM, waits up to
// takeoverWaitForExit for graceful release, escalates to SIGKILL if needed,
// and returns nil once the port is free. If the holder is foreign (or
// unidentifiable) it returns a descriptive error that the caller surfaces to
// the user — we never signal an unfamiliar process.
//
// Returns nil with no action when no holder is found (caller should retry the
// bind regardless — the EADDRINUSE may have raced with a port release).
func TakeOverPort(port int) error {
	holder, err := findPortHolder(port)
	if err != nil {
		return fmt.Errorf("identifying holder of port %d: %w", port, err)
	}
	if holder == nil {
		return nil
	}
	if !IsOurDaemon(holder.PID) {
		cmd := holder.Command
		if cmd == "" {
			cmd = "<unknown>"
		}
		return fmt.Errorf("port %d is held by foreign process %q (pid %d); refusing to kill — "+
			"stop that process manually or change oauth_port", port, cmd, holder.PID)
	}
	fmt.Fprintf(os.Stderr, "auth: replacing stale gsuite-mcp daemon (pid %d) on port %d\n", holder.PID, port)
	if err := signalProcess(holder.PID, syscall.SIGTERM); err != nil {
		// ESRCH = process already gone, which is fine — port should be free.
		if err != syscall.ESRCH {
			return fmt.Errorf("SIGTERM to gsuite-mcp pid %d: %w", holder.PID, err)
		}
	}
	deadline := time.Now().Add(takeoverWaitForExit)
	for time.Now().Before(deadline) {
		if !processExists(holder.PID) {
			break
		}
		time.Sleep(takeoverPollInterval)
	}
	if processExists(holder.PID) {
		fmt.Fprintf(os.Stderr, "auth: gsuite-mcp pid %d did not exit within %s; sending SIGKILL\n", holder.PID, takeoverWaitForExit)
		_ = signalProcess(holder.PID, syscall.SIGKILL)
		time.Sleep(takeoverPollInterval * 4)
	}
	// One last sanity check: is the port actually free now?
	if h, _ := findPortHolder(port); h != nil {
		return fmt.Errorf("port %d still held after SIGKILL of pid %d (now held by pid %d)", port, holder.PID, h.PID)
	}
	return nil
}
