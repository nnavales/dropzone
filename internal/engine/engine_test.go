package engine

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/nnavales/dropzone/internal/config"
)

func TestResolveAction(t *testing.T) {
	now := time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)
	src := filepath.Join(t.TempDir(), "report.txt")

	cases := map[string]struct {
		cmd  string
		want string
	}{
		"file":      {"cp {file} /dst", "cp " + src + " /dst"},
		"filename":  {"echo {filename}", "echo report.txt"},
		"basename":  {"echo {basename}", "echo report"},
		"extension": {"echo {extension}", "echo .txt"},
		"date":      {"echo {date}", "echo 2026-09-23"},
		"timestamp": {"echo {timestamp}", "echo " + strconv.FormatInt(now.Unix(), 10)},
		"multiple":  {"{basename}{extension}", "report.txt"},
		"none":      {"echo hi", "echo hi"},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := resolveAction(tc.cmd, src, now); got != tc.want {
				t.Errorf("resolveAction() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestZoneFor(t *testing.T) {
	dir := t.TempDir()
	e := New(config.Config{Zones: []config.Zone{
		{Name: "a", Path: filepath.Join(dir, "a")},
		{Name: "b", Path: filepath.Join(dir, "b")},
	}})

	t.Run("file inside zone", func(t *testing.T) {
		z := e.zoneFor(filepath.Join(dir, "a", "f.txt"))
		if z == nil || z.Name != "a" {
			t.Errorf("zoneFor() = %+v, want zone a", z)
		}
	})

	t.Run("sibling prefix does not match", func(t *testing.T) {
		if z := e.zoneFor(filepath.Join(dir, "abc", "f.txt")); z != nil {
			t.Errorf("zoneFor() = %+v, want nil", z)
		}
	})

	t.Run("outside all zones", func(t *testing.T) {
		if z := e.zoneFor(filepath.Join(dir, "c", "f.txt")); z != nil {
			t.Errorf("zoneFor() = %+v, want nil", z)
		}
	})
}

func TestHandleFile(t *testing.T) {
	ctx := context.Background()

	newEngine := func(t *testing.T, rules []config.Rule) (*Engine, string) {
		t.Helper()
		zone := t.TempDir()
		e := New(config.Config{Zones: []config.Zone{
			{Name: "z", Path: zone, Rules: rules},
		}})
		return e, zone
	}

	writeFile := func(t *testing.T, dir, name, content string) string {
		t.Helper()
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}

	t.Run("moves matching file", func(t *testing.T) {
		dst := t.TempDir()
		e, zone := newEngine(t, []config.Rule{{
			Name:  "r",
			Match: config.Match{Extensions: []string{".txt"}},
			Action: config.Action{Kind: "move", Target: dst},
		}})
		src := writeFile(t, zone, "a.txt", "data")

		e.handleFile(ctx, src)

		if _, err := os.Stat(src); !os.IsNotExist(err) {
			t.Error("source should have been moved")
		}
		if data, err := os.ReadFile(filepath.Join(dst, "a.txt")); err != nil || string(data) != "data" {
			t.Errorf("destination = %q, %v; want %q", data, err, "data")
		}
	})

	t.Run("ignores non-matching file", func(t *testing.T) {
		dst := t.TempDir()
		e, zone := newEngine(t, []config.Rule{{
			Name:  "r",
			Match: config.Match{Extensions: []string{".txt"}},
			Action: config.Action{Kind: "move", Target: dst},
		}})
		src := writeFile(t, zone, "a.log", "data")

		e.handleFile(ctx, src)

		if _, err := os.Stat(src); err != nil {
			t.Errorf("non-matching file should be left alone: %v", err)
		}
	})

	t.Run("ignores missing file and directory", func(t *testing.T) {
		e, zone := newEngine(t, []config.Rule{{
			Name:  "r",
			Match: config.Match{Extensions: []string{".txt"}},
			Action: config.Action{Kind: "delete"},
		}})

		e.handleFile(ctx, filepath.Join(zone, "nope.txt"))
		e.handleFile(ctx, zone)
	})
}
