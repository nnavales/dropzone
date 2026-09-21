package actions

import (
	"path/filepath"
	"strings"
)

// renderCommand expands per-file placeholders in run commands.
func renderCommand(cmd, source string) string {
	return renderPlaceholders(cmd, source)
}

// renderName expands per-file placeholders in rename targets.
func renderName(name, source string) string {
	return renderPlaceholders(name, source)
}

func renderPlaceholders(s, source string) string {
	name := filepath.Base(source)
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	r := strings.NewReplacer(
		"{path}", source,
		"{name}", name,
		"{dir}", filepath.Dir(source),
		"{stem}", stem,
		"{ext}", strings.TrimPrefix(ext, "."),
	)
	return r.Replace(s)
}
