package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"path/filepath"
	"runtime"
)

// StringHash return sha256 checksum encoded by base64
func StringHash(source string) string {
	hashed := sha256.Sum256([]byte(source))
	return base64.StdEncoding.EncodeToString(hashed[:])
}

// Hash returns SHA256 hash
func Hash(source string) [32]byte {
	return sha256.Sum256([]byte(source))
}

// AbsPath add project root path before a relative path
func AbsPath(rel string) string {
	_, projectRoot, _, _ := runtime.Caller(0)             // get dir of current file setting.go
	projectRoot = filepath.Dir(filepath.Dir(projectRoot)) // get project root path
	if !filepath.IsAbs(rel) {
		return filepath.Join(projectRoot, rel)
	}
	return rel
}
