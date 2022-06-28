// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package utils

func MustWrite(n int, err error) {
	if err != nil {
		panic(err)
	}
}
