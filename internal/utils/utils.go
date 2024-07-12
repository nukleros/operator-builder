// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package utils

func MustWrite(n int, err error) {
	if err != nil {
		panic(err)
	}
}
