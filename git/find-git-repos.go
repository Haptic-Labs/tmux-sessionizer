package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// findGitRepos searches for git repos recursively from the given root
func FindGitRepos(root string) ([]string, error) {
	var repos []string

	// Check if the root directory itself is a git repository
	if IsGitRepo(root) {
		repos = append(repos, root)
	}

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip directories we can't access
			if os.IsPermission(err) {
				fmt.Fprintf(os.Stderr, "Permission denied: %v\n", path)
				return filepath.SkipDir
			}
			return err
		}

		// Skip the root directory since we've already checked it
		if path == root {
			return nil
		}

		// Skip hidden directories (those starting with .)
		if info.IsDir() && strings.HasPrefix(filepath.Base(path), ".") && path != root {
			return filepath.SkipDir
		}

		// If this directory is a git repository, add it to our list
		if info.IsDir() && IsGitRepo(path) {
			repos = append(repos, path)
			return filepath.SkipDir // Skip traversing into git repositories
		}

		return nil
	})

	return repos, err
}
