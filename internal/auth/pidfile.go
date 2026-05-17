package auth

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aliwatters/gsuite-mcp/internal/config"
)

// pidFileName is the filename used for the running MCP-server pidfile inside
// the config directory. Centralising the constant keeps Write/Read/Remove and
// IsOurDaemon agreeing on a single path.
const pidFileName = "server.pid"

// PidFilePath returns the absolute path to the pidfile used to identify the
// running gsuite-mcp daemon. Lives next to config.json so a single
// GSUITE_MCP_CONFIG_DIR override moves both.
func PidFilePath() string {
	return filepath.Join(config.DefaultConfigDir(), pidFileName)
}

// WritePidFile records the current process's PID so that a future
// `gsuite-mcp auth` invocation can reliably identify whether the holder of
// the OAuth port is our own daemon (gsuite-mcp#158).
//
// Returns an error if the parent directory cannot be created or the file
// cannot be written. Errors are typically surfaced as warnings, not fatal:
// the daemon runs fine without a pidfile, only the auth-time port-takeover
// loses its primary identification signal (it falls back to a process-name
// match in that case).
func WritePidFile() error {
	dir := config.DefaultConfigDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating config dir for pidfile: %w", err)
	}
	pid := strconv.Itoa(os.Getpid())
	if err := os.WriteFile(PidFilePath(), []byte(pid+"\n"), 0o644); err != nil {
		return fmt.Errorf("writing pidfile: %w", err)
	}
	return nil
}

// ReadPidFile returns the PID recorded in the pidfile. Returns (0, error) when
// the file is missing, empty, or unparseable — callers should fall back to
// process-name matching in those cases.
func ReadPidFile() (int, error) {
	data, err := os.ReadFile(PidFilePath())
	if err != nil {
		return 0, err
	}
	s := strings.TrimSpace(string(data))
	if s == "" {
		return 0, fmt.Errorf("pidfile %s is empty", PidFilePath())
	}
	pid, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("pidfile %s does not contain a valid PID (%q): %w", PidFilePath(), s, err)
	}
	if pid <= 0 {
		return 0, fmt.Errorf("pidfile %s contains non-positive PID %d", PidFilePath(), pid)
	}
	return pid, nil
}

// RemovePidFile deletes the pidfile. Best-effort: callers should ignore the
// error (the pidfile is non-authoritative state and a stale one is handled
// by the auth-time port-takeover logic).
func RemovePidFile() error {
	err := os.Remove(PidFilePath())
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
