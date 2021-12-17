// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func ReadStream(fileName string) (io.ReadCloser, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to %w", err)
	}

	return file, nil
}

// CloseFile safely closes a file handle.
func CloseFile(file io.ReadCloser) {
	if err := file.Close(); err != nil {
		log.Fatalf("error closing file!: %s", err)
	}
}

// Glob adds double-star support to the core path/filepath Glob function.
// It's useful when your globs might have double-stars, but you're not sure.
func Glob(pattern string) ([]string, error) {
	if !strings.Contains(pattern, "**") {
		// passthru to core package if no double-star
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return matches, fmt.Errorf("unable to expand glob, %w", err)
		}

		return matches, nil
	}

	return expand(strings.Split(pattern, "**"))
}

// expand finds matches for the provided Globs.
func expand(g []string) ([]string, error) {
	matches := []string{""}

	for _, glob := range g {
		var hits []string

		hitMap := map[string]bool{}

		for _, match := range matches {
			paths, err := filepath.Glob(match + glob)
			if err != nil {
				return nil, fmt.Errorf("unable to expand glob, %w", err)
			}

			for _, path := range paths {
				err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					// save deduped match from current iteration
					if _, ok := hitMap[path]; !ok {
						hits = append(hits, path)
						hitMap[path] = true
					}

					return nil
				})

				if err != nil {
					return nil, fmt.Errorf("unable to expand glob, %w", err)
				}
			}
		}

		matches = hits
	}

	// fix up return value for nil input
	if g == nil && len(matches) > 0 && matches[0] == "" {
		matches = matches[1:]
	}

	return matches, nil
}
