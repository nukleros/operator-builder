// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

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
