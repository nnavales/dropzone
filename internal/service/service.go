// Package service manages the systemd user unit for dropzone.
package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const name = "dropzone"
const unitFile = "dropzone.service"

// Status describes the current installation state.
type Status struct {
	UnitPath  string
	Binary    string
	Installed bool
	Enabled   string // enabled, disabled, unknown
	Active    string // active, inactive, unknown
}

// Install writes the unit file and enables + starts it via systemctl --user.
func Install() (string, error) {
	bin, err := binaryPath()
	if err != nil {
		return "", err
	}
	p, err := unitPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(p, []byte(unitContent(bin)), 0o644); err != nil {
		return "", err
	}
	if out, err := runSystemctl("--user", "daemon-reload"); err != nil {
		return "", fmt.Errorf("daemon-reload failed: %s: %w", out, err)
	}
	if out, err := runSystemctl("--user", "enable", "--now", name); err != nil {
		return "", fmt.Errorf("enable --now failed: %s: %w", out, err)
	}
	return p, nil
}

// Uninstall stops and disables the unit, then removes the unit file. (idempotent)
func Uninstall() (string, error) {
	p, err := unitPath()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(p); err != nil {
		if os.IsNotExist(err) {
			return p, nil
		}
		return "", err
	}
	if out, err := runSystemctl("--user", "disable", "--now", name); err != nil {
		return "", fmt.Errorf("disable --now failed: %s: %w", out, err)
	}
	if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
		return "", err
	}
	// Best effort: clear a stale failed state so status reads clean.
	_, _ = runSystemctl("--user", "reset-failed", name)
	if out, err := runSystemctl("--user", "daemon-reload"); err != nil {
		return "", fmt.Errorf("daemon-reload failed: %s: %w", out, err)
	}
	return p, nil
}

// Inspect returns the current status without changing anything.
func Inspect() (Status, error) {
	var st Status
	p, err := unitPath()
	if err != nil {
		return st, err
	}
	st.UnitPath = p
	st.Binary, _ = binaryPath()

	if _, err := os.Stat(p); err == nil {
		st.Installed = true
	} else if !os.IsNotExist(err) {
		return st, err
	}

	st.Enabled = "unknown"
	if out, err := runSystemctl("--user", "is-enabled", name); err == nil {
		st.Enabled = out
	} else if out != "" {
		// is-enabled exits non-zero for disabled/missing; output still tells state.
		st.Enabled = out
	}

	st.Active = "unknown"
	if out, err := runSystemctl("--user", "is-active", name); err == nil {
		st.Active = out
	} else if out != "" {
		st.Active = out
	}
	return st, nil
}

// Logs streams the service logs directly to stdout/stderr.
func Logs(lines int, follow bool) error {
	if _, err := exec.LookPath("journalctl"); err != nil {
		return fmt.Errorf("journalctl not found: %w", err)
	}
	args := []string{"--user", "-u", name, "--no-pager"}
	if lines > 0 {
		args = append(args, "-n", strconv.Itoa(lines))
	}
	if follow {
		args = append(args, "-f")
	}
	cmd := exec.Command("journalctl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// unitPath returns the path to the systemd user unit.
func unitPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "systemd", "user", unitFile), nil
}

// unitContent renders the unit file for the given binary path.
func unitContent(bin string) string {
	return fmt.Sprintf(`[Unit]
Description=dropzone - watch directories and act on stable files

[Service]
ExecStart=%s run
Restart=on-failure
RestartSec=2

[Install]
WantedBy=default.target
`, bin)
}

// binaryPath resolves the current executable to an absolute path.
func binaryPath() (string, error) {
	p, err := os.Executable()
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		p = resolved
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	return abs, nil
}

func runSystemctl(args ...string) (string, error) {
	if _, err := exec.LookPath("systemctl"); err != nil {
		return "", fmt.Errorf("systemctl not found: %w", err)
	}
	cmd := exec.Command("systemctl", args...)
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}
