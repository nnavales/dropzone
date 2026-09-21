package rules

import (
	"path/filepath"
)

// File is the DTO used for rules application.
type File struct {
	Path string
	Name string
}

// NewFile builds a File from a path without touching disk.
func NewFile(path string) File {
	return File{
		Path: path,
		Name: filepath.Base(path),
	}
}
