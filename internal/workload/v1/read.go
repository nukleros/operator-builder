// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"fmt"
	"io"
	"log"
	"os"
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
