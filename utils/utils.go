package utils

import (
	"path/filepath"
	"runtime"
)

// AbsPath add project root path before a relative path
func AbsPath(rel string) string {
	_, projectRoot, _, _ := runtime.Caller(0)                           // get dir of current file setting.go
	projectRoot = filepath.Dir(filepath.Dir(filepath.Dir(projectRoot))) // get project root path
	if !filepath.IsAbs(rel) {
		return filepath.Join(projectRoot, rel)
	}
	return rel
}
