package config

import (
	_ "embed"
	"strings"
	"testing"
)

//go:embed testdata/valid.yml
var validYAML []byte

//go:embed testdata/defaults.yml
var defaultsYAML []byte

//go:embed testdata/invalid/empty_zones.yml
var emptyZonesYAML []byte

//go:embed testdata/invalid/negative_wait.yml
var negativeWaitYAML []byte

//go:embed testdata/invalid/bad_conflict.yml
var badConflictYAML []byte

//go:embed testdata/invalid/missing_path.yml
var missingPathYAML []byte

//go:embed testdata/invalid/empty_match.yml
var emptyMatchYAML []byte

//go:embed testdata/invalid/empty_action.yml
var emptyActionYAML []byte

//go:embed testdata/invalid/two_actions.yml
var twoActionsYAML []byte

func TestParseValid(t *testing.T) {
	cfg, err := Parse(validYAML)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if cfg.Settings.WaitSeconds != 2 || cfg.Settings.Conflict != "rename" {
		t.Errorf("settings = %+v, want {2 rename}", cfg.Settings)
	}
	rules := cfg.Zones[0].Rules
	if len(rules) != 2 {
		t.Fatalf("len(rules) = %d, want 2", len(rules))
	}
	if rules[1].Action.Kind != ActionDelete {
		t.Errorf("delete action did not parse, kind = %q", rules[1].Action.Kind)
	}
}

func TestParseDefaults(t *testing.T) {
	cfg, err := Parse(defaultsYAML)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if cfg.Settings.WaitSeconds != 2 {
		t.Errorf("WaitSeconds = %d, want default 2", cfg.Settings.WaitSeconds)
	}
	if cfg.Settings.Conflict != "rename" {
		t.Errorf("Conflict = %q, want default rename", cfg.Settings.Conflict)
	}
}

func TestParseInvalid(t *testing.T) {
	cases := []struct {
		name    string
		yml     []byte
		wantErr string
	}{
		{"empty zones", emptyZonesYAML, "zones must not be empty"},
		{"negative wait", negativeWaitYAML, "settings.wait_seconds"},
		{"bad conflict", badConflictYAML, "settings.conflict"},
		{"missing path", missingPathYAML, "path is required"},
		{"empty match", emptyMatchYAML, "extensions|glob"},
		{"empty action", emptyActionYAML, "one of move|copy|rename|run|delete is required"},
		{"two actions in one entry", twoActionsYAML, "only one of"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(tc.yml)
			if err == nil {
				t.Fatalf("Parse() expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %q, want substring %q", err.Error(), tc.wantErr)
			}
		})
	}
}
