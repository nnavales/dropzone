package service

import (
	"strings"
	"testing"
)

func TestUnitContentHasNoNetworkDependency(t *testing.T) {
	out := unitContent("/usr/bin/dropzone")
	if strings.Contains(out, "network-online") {
		t.Fatalf("unit should not wait on network:\n%s", out)
	}
	if !strings.Contains(out, "ExecStart=/usr/bin/dropzone run") {
		t.Fatalf("unit missing ExecStart:\n%s", out)
	}
}

func TestUninstallIdempotentWhenMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	path, err := Uninstall()
	if err != nil {
		t.Fatalf("second uninstall should succeed: %v", err)
	}
	if path == "" {
		t.Fatal("expected unit path back")
	}

	// Repeat: must still succeed.
	if _, err := Uninstall(); err != nil {
		t.Fatalf("repeated uninstall should succeed: %v", err)
	}
}
