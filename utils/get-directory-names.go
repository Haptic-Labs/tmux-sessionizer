package utils

import (
	"path/filepath"
)

// GetDirectoryNames extracts just the directory names (not full paths) from a list of paths
func GetDirectoryNames(paths []string) map[string]string {
	// Using a map to store name->path mapping
	dirMap := make(map[string]string)

	for _, path := range paths {
		name := filepath.Base(path)
		dirMap[name] = path
	}

	return dirMap
}
