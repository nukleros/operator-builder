// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package subcommand

import (
	"github.com/nukleros/operator-builder/internal/workload/v1/config"
)

// Init runs the process logic for a config processor when running the `init`
// subcommand.
func Init(processor *config.Processor) error {
	workload := processor.Workload

	workload.SetNames()

	return nil
}
