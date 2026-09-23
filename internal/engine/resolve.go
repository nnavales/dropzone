package engine

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type resolver func(src string, now time.Time) string

var funcMap = map[string]resolver{
	"{file}": func(src string, now time.Time) string {
		return src
	},
	"{filename}": func(src string, now time.Time) string {
		return filepath.Base(src)
	},
	"{basename}": func(src string, now time.Time) string {
		name := filepath.Base(src)
		return strings.TrimSuffix(name, filepath.Ext(name))
	},
	"{extension}": func(src string, now time.Time) string {
		return filepath.Ext(src)
	},
	"{date}": func(src string, now time.Time) string {
		return now.Format("2006-01-02")
	},
	"{timestamp}": func(src string, now time.Time) string {
		return strconv.FormatInt(now.Unix(), 10)
	},
}

// resolveAction expands placeholders like {filename} in cmd.
func resolveAction(cmd, src string, now time.Time) string {
	for placeholder, resolve := range funcMap {
		cmd = strings.ReplaceAll(cmd, placeholder, resolve(src, now))
	}

	return cmd
}
