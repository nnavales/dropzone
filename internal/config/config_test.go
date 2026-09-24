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

//go:embed testdata/invalid/relative_path.yml
var relativePathYAML []byte

//go:embed testdata/invalid/duplicate_path.yml
var duplicatePathYAML []byte

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

	if cfg.Settings.StableForSeconds != 2 || cfg.Settings.OnConflict != "rename" {
		t.Errorf("settings = %+v, want {2 rename}", cfg.Settings)
	}
	rules := cfg.Zones[0].Rules
	if len(rules) != 2 {
		t.Fatalf("len(rules) = %d, want 2", len(rules))
	}
	if rules[1].Action.Kind != "delete" {
		t.Errorf("delete action did not parse, kind = %q", rules[1].Action.Kind)
	}
}

func TestParseDefaults(t *testing.T) {
	cfg, err := Parse(defaultsYAML)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if cfg.Settings.StableForSeconds != 2 {
		t.Errorf("StableForSeconds = %d, want default 2", cfg.Settings.StableForSeconds)
	}
	if cfg.Settings.OnConflict != "rename" {
		t.Errorf("OnConflict = %q, want default rename", cfg.Settings.OnConflict)
	}
}

func TestParseEmptyZones(t *testing.T) {
	cfg, err := Parse(emptyZonesYAML)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(cfg.Zones) != 0 {
		t.Errorf("len(zones) = %d, want 0", len(cfg.Zones))
	}
}

func TestParseInvalid(t *testing.T) {
	cases := []struct {
		name    string
		yml     []byte
		wantErr string
	}{
		{"negative wait", negativeWaitYAML, "settings.stable_for_seconds"},
		{"bad conflict", badConflictYAML, "settings.on_conflict"},
		{"missing path", missingPathYAML, "path is required"},
		{"relative path", relativePathYAML, "path must be absolute"},
		{"duplicate path", duplicatePathYAML, "duplicate path"},
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
