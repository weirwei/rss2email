package helpers

import (
	"os"
	"path/filepath"
)

// ResolveProjectRoot resolves project root by searching for db/rss.sql upward.
func ResolveProjectRoot() string {
	var candidates []string

	if envRoot := os.Getenv("RSS2EMAIL_ROOT"); envRoot != "" {
		candidates = append(candidates, envRoot)
	}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates, cwd)
	}
	if exe, err := os.Executable(); err == nil {
		candidates = append(candidates, filepath.Dir(exe))
	}

	for _, base := range candidates {
		if root, ok := findProjectRoot(base); ok {
			return root
		}
	}
	return "."
}

func findProjectRoot(start string) (string, bool) {
	if start == "" {
		return "", false
	}
	dir := start
	for {
		marker := filepath.Join(dir, "db", "rss.sql")
		if st, err := os.Stat(marker); err == nil && !st.IsDir() {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}
