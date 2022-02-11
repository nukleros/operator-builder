// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrConvertArrayInterface     = errors.New("unable to convert to []interface{}")
	ErrConvertArrayString        = errors.New("unable to convert to []string")
	ErrConvertMapStringInterface = errors.New("unable to convert to map[string]interface{}")
	ErrConvertString             = errors.New("unable to convert to string")
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
	//nolint:nestif //refactor
	if !strings.Contains(pattern, "**") {
		// ensure the actual path exists if a glob pattern is not found
		if !strings.Contains(pattern, "*") {
			if _, err := os.Stat(pattern); errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("%w; file %s defined in spec.resources cannot be found", err, pattern)
			}
		}

		// passthru to core package if no double-star
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return matches, fmt.Errorf("unable to expand glob, %w", err)
		}

		if len(matches) == 0 {
			return nil, fmt.Errorf("%w; unable to find any files from glob pattern %s", os.ErrNotExist, pattern)
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

func toArrayInterface(in interface{}) ([]interface{}, error) {
	out, ok := in.([]interface{})
	if !ok {
		return nil, ErrConvertArrayInterface
	}

	return out, nil
}

func toArrayString(in interface{}) ([]string, error) {
	// attempt direct conversion
	out, ok := in.([]string)
	if !ok {
		// attempt conversion for each item
		outInterfaces, err := toArrayInterface(in)
		if err != nil {
			return nil, fmt.Errorf("%w; %s", err, ErrConvertArrayString)
		}

		outStrings := make([]string, len(outInterfaces))

		for i := range outInterfaces {
			outString, err := toString(outInterfaces[i])
			if err != nil {
				return nil, fmt.Errorf("%w; %s", err, ErrConvertArrayString)
			}

			outStrings[i] = outString
		}

		return outStrings, nil
	}

	return out, nil
}

func toMapStringInterface(in interface{}) (map[string]interface{}, error) {
	out, ok := in.(map[string]interface{})
	if !ok {
		return nil, ErrConvertMapStringInterface
	}

	return out, nil
}

func toString(in interface{}) (string, error) {
	out, ok := in.(string)
	if !ok {
		return "", ErrConvertString
	}

	return out, nil
}
