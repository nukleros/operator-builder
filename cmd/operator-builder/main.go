// Copyright 2024 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/nukleros/operator-builder/internal/plugins/workload"
	"github.com/nukleros/operator-builder/pkg/cli"
)

func main() {
	command, err := cli.NewKubebuilderCLI(workload.FromEnv())
	if err != nil {
		log.Fatal(err)
	}

	if err := command.Run(); err != nil {
		log.Fatal(err)
	}
}
