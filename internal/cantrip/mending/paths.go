package mending

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/emiliopalmerini/grimorio/internal/lsp"
)

func ExpandPaths(patterns []string) ([]string, error) {
	var files []string
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		if strings.HasSuffix(pattern, "/...") {
			dir := strings.TrimSuffix(pattern, "/...")
			err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() && IsSupportedFile(path) {
					absPath, err := filepath.Abs(path)
					if err != nil {
						return fmt.Errorf("failed to get absolute path for %s: %w", path, err)
					}
					if !seen[absPath] {
						seen[absPath] = true
						files = append(files, path)
					}
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			info, err := os.Stat(pattern)
			if err != nil {
				return nil, err
			}
			if info.IsDir() {
				entries, err := os.ReadDir(pattern)
				if err != nil {
					return nil, err
				}
				for _, entry := range entries {
					if !entry.IsDir() {
						path := filepath.Join(pattern, entry.Name())
						if IsSupportedFile(path) {
							absPath, err := filepath.Abs(path)
							if err != nil {
								return nil, fmt.Errorf("failed to get absolute path for %s: %w", path, err)
							}
							if !seen[absPath] {
								seen[absPath] = true
								files = append(files, path)
							}
						}
					}
				}
			} else {
				absPath, err := filepath.Abs(pattern)
				if err != nil {
					return nil, fmt.Errorf("failed to get absolute path for %s: %w", pattern, err)
				}
				if !seen[absPath] {
					seen[absPath] = true
					files = append(files, pattern)
				}
			}
		}
	}

	return files, nil
}

func IsSupportedFile(path string) bool {
	return lsp.DetectLanguage(path) != nil
}
