package auth

import (
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"syscall"
	"testing"

	"github.com/aliwatters/gsuite-mcp/internal/config"
)

// withTempConfigDir redirects the config-dir for the duration of t. Restores
// the prior override on Cleanup. Used by every test that touches the pidfile.
func withTempConfigDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	prev := os.Getenv("GSUITE_MCP_CONFIG_DIR")
	t.Setenv("GSUITE_MCP_CONFIG_DIR", dir)
	config.SetConfigDir(dir)
	t.Cleanup(func() {
		if prev != "" {
			os.Setenv("GSUITE_MCP_CONFIG_DIR", prev)
			config.SetConfigDir(prev)
		} else {
			os.Unsetenv("GSUITE_MCP_CONFIG_DIR")
			config.SetConfigDir("") // reset to default
		}
	})
	return dir
}

// === pidfile lifecycle ===

func TestPidFile_WriteReadRemove(t *testing.T) {
	dir := withTempConfigDir(t)

	if err := WritePidFile(); err != nil {
		t.Fatalf("WritePidFile: %v", err)
	}

	// File at expected location with our PID.
	path := filepath.Join(dir, "server.pid")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading pidfile at %s: %v", path, err)
	}
	got, err := strconv.Atoi(string(data[:len(data)-1])) // strip trailing newline
	if err != nil {
		t.Fatalf("pidfile contents not a PID: %v (raw: %q)", err, string(data))
	}
	if got != os.Getpid() {
		t.Errorf("WritePidFile: file has pid %d, want %d", got, os.Getpid())
	}

	// ReadPidFile round-trip.
	readPID, err := ReadPidFile()
	if err != nil {
		t.Fatalf("ReadPidFile: %v", err)
	}
	if readPID != os.Getpid() {
		t.Errorf("ReadPidFile: got %d, want %d", readPID, os.Getpid())
	}

	if err := RemovePidFile(); err != nil {
		t.Fatalf("RemovePidFile: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("after RemovePidFile, expected file gone; stat err: %v", err)
	}

	// RemovePidFile is idempotent — second call must not error.
	if err := RemovePidFile(); err != nil {
		t.Errorf("RemovePidFile (second call): expected no error, got %v", err)
	}
}

func TestReadPidFile_MissingReturnsError(t *testing.T) {
	withTempConfigDir(t)
	if _, err := ReadPidFile(); err == nil {
		t.Error("ReadPidFile on missing file: expected error, got nil")
	}
}

func TestReadPidFile_EmptyReturnsError(t *testing.T) {
	dir := withTempConfigDir(t)
	if err := os.WriteFile(filepath.Join(dir, "server.pid"), []byte("   \n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, err := ReadPidFile(); err == nil {
		t.Error("ReadPidFile on empty file: expected error, got nil")
	}
}

func TestReadPidFile_NonNumericReturnsError(t *testing.T) {
	dir := withTempConfigDir(t)
	if err := os.WriteFile(filepath.Join(dir, "server.pid"), []byte("not-a-pid"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if _, err := ReadPidFile(); err == nil {
		t.Error("ReadPidFile on non-numeric file: expected error, got nil")
	}
}

// === IsOurDaemon classification ===

func TestIsOurDaemon_PidFileMatchWins(t *testing.T) {
	withTempConfigDir(t)
	// Pidfile says 12345 is ours. Even with a generic command name, the
	// pidfile match is authoritative.
	if err := os.WriteFile(PidFilePath(), []byte("12345\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	origCmd := commandForPID
	defer func() { commandForPID = origCmd }()
	commandForPID = func(pid int) (string, error) { return "some-other-binary", nil }

	if !IsOurDaemon(12345) {
		t.Error("IsOurDaemon: pidfile match should win regardless of command name")
	}
}

func TestIsOurDaemon_FallbackToCommandName(t *testing.T) {
	withTempConfigDir(t) // pidfile absent
	origCmd := commandForPID
	defer func() { commandForPID = origCmd }()
	commandForPID = func(pid int) (string, error) { return "gsuite-mc", nil } // Linux 15-char truncation

	if !IsOurDaemon(99999) {
		t.Error("IsOurDaemon: should match the Linux-truncated gsuite-mc name")
	}
}

func TestIsOurDaemon_RejectsForeignCommand(t *testing.T) {
	withTempConfigDir(t)
	origCmd := commandForPID
	defer func() { commandForPID = origCmd }()
	commandForPID = func(pid int) (string, error) { return "vim", nil }

	if IsOurDaemon(99999) {
		t.Error("IsOurDaemon: vim should NOT be classified as our daemon")
	}
}

func TestIsOurDaemon_PidFileStaleFallsBackToName(t *testing.T) {
	withTempConfigDir(t)
	// Pidfile says 11111 is ours, but we ask about 22222. Fallback should
	// kick in and use the command name.
	if err := os.WriteFile(PidFilePath(), []byte("11111\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	origCmd := commandForPID
	defer func() { commandForPID = origCmd }()
	commandForPID = func(pid int) (string, error) { return "GSUITE-MCP", nil } // case-insensitive match

	if !IsOurDaemon(22222) {
		t.Error("IsOurDaemon: should fall back to command-name match when pidfile points elsewhere")
	}
}

func TestIsOurDaemon_RejectsZeroPid(t *testing.T) {
	if IsOurDaemon(0) {
		t.Error("IsOurDaemon(0): must reject non-positive PIDs")
	}
	if IsOurDaemon(-1) {
		t.Error("IsOurDaemon(-1): must reject non-positive PIDs")
	}
}

// === TakeOverPort behaviour ===

func TestTakeOverPort_NoHolderIsNoOp(t *testing.T) {
	origFind := findPortHolder
	defer func() { findPortHolder = origFind }()
	findPortHolder = func(port int) (*PortHolder, error) { return nil, nil }

	if err := TakeOverPort(38917); err != nil {
		t.Errorf("TakeOverPort with no holder: expected nil, got %v", err)
	}
}

func TestTakeOverPort_ForeignProcessNeverKilled(t *testing.T) {
	withTempConfigDir(t) // no pidfile
	origFind := findPortHolder
	origCmd := commandForPID
	origSig := signalProcess
	defer func() {
		findPortHolder = origFind
		commandForPID = origCmd
		signalProcess = origSig
	}()

	findPortHolder = func(port int) (*PortHolder, error) {
		return &PortHolder{PID: 4242, Command: "ssh-agent"}, nil
	}
	commandForPID = func(pid int) (string, error) { return "ssh-agent", nil }

	var signalled atomic.Bool
	signalProcess = func(pid int, sig syscall.Signal) error {
		signalled.Store(true)
		return nil
	}

	err := TakeOverPort(38917)
	if err == nil {
		t.Fatal("TakeOverPort with foreign holder: expected error, got nil")
	}
	if signalled.Load() {
		t.Error("TakeOverPort sent a signal to a foreign process — this must never happen")
	}
	// Error message must name the foreign command + PID so the operator can act.
	msg := err.Error()
	if !contains(msg, "ssh-agent") || !contains(msg, "4242") {
		t.Errorf("error message missing command/PID context: %q", msg)
	}
}

func TestTakeOverPort_OurDaemonSigTermThenFreed(t *testing.T) {
	withTempConfigDir(t)
	// Pidfile says 7777 is ours.
	if err := os.WriteFile(PidFilePath(), []byte("7777\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	origFind := findPortHolder
	origSig := signalProcess
	origExists := processExists
	defer func() {
		findPortHolder = origFind
		signalProcess = origSig
		processExists = origExists
	}()

	var (
		sigterm   atomic.Int32
		sigkill   atomic.Int32
		findCalls atomic.Int32
	)

	findPortHolder = func(port int) (*PortHolder, error) {
		n := findCalls.Add(1)
		if n == 1 {
			// First call: report our daemon holding the port.
			return &PortHolder{PID: 7777, Command: "gsuite-mcp"}, nil
		}
		// Subsequent calls (after SIGTERM): port is free.
		return nil, nil
	}
	signalProcess = func(pid int, sig syscall.Signal) error {
		switch sig {
		case syscall.SIGTERM:
			sigterm.Add(1)
		case syscall.SIGKILL:
			sigkill.Add(1)
		}
		return nil
	}
	// Process "exits" immediately after SIGTERM.
	processExists = func(pid int) bool { return false }

	if err := TakeOverPort(38917); err != nil {
		t.Fatalf("TakeOverPort with our daemon: expected nil, got %v", err)
	}
	if sigterm.Load() != 1 {
		t.Errorf("expected 1 SIGTERM, got %d", sigterm.Load())
	}
	if sigkill.Load() != 0 {
		t.Errorf("expected 0 SIGKILL (process exited gracefully), got %d", sigkill.Load())
	}
}

func TestTakeOverPort_OurDaemonEscalatesToSigKill(t *testing.T) {
	withTempConfigDir(t)
	if err := os.WriteFile(PidFilePath(), []byte("8888\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	origFind := findPortHolder
	origSig := signalProcess
	origExists := processExists
	defer func() {
		findPortHolder = origFind
		signalProcess = origSig
		processExists = origExists
	}()

	var (
		sigterm atomic.Int32
		sigkill atomic.Int32
	)
	// First findPortHolder call reports our daemon; final check after SIGKILL says port free.
	findCalls := atomic.Int32{}
	findPortHolder = func(port int) (*PortHolder, error) {
		n := findCalls.Add(1)
		if n == 1 {
			return &PortHolder{PID: 8888, Command: "gsuite-mcp"}, nil
		}
		return nil, nil
	}
	signalProcess = func(pid int, sig syscall.Signal) error {
		switch sig {
		case syscall.SIGTERM:
			sigterm.Add(1)
		case syscall.SIGKILL:
			sigkill.Add(1)
		}
		return nil
	}
	// Process refuses to exit after SIGTERM; becomes "gone" only after the
	// SIGKILL stage waits.
	var killed atomic.Bool
	processExists = func(pid int) bool { return !killed.Load() }
	prevSig := signalProcess
	signalProcess = func(pid int, sig syscall.Signal) error {
		if sig == syscall.SIGKILL {
			killed.Store(true)
		}
		return prevSig(pid, sig)
	}

	if err := TakeOverPort(38917); err != nil {
		t.Fatalf("TakeOverPort: expected nil, got %v", err)
	}
	if sigterm.Load() != 1 {
		t.Errorf("expected exactly 1 SIGTERM, got %d", sigterm.Load())
	}
	if sigkill.Load() != 1 {
		t.Errorf("expected exactly 1 SIGKILL (process did not exit gracefully), got %d", sigkill.Load())
	}
}

// helper that doesn't pull in strings.Contains import noise into the test file
func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
