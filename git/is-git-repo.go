package git

import (
	"os"
	"path/filepath"
)

func IsGitRepo(dir string) bool {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}
