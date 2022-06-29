// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package config

import (
	"errors"
	"fmt"

	structvalidator "github.com/go-playground/validator"

	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/kinds"
)

var (
	ErrUniqueNames = errors.New("each workload name must be unique")
	ErrUniqueKinds = errors.New("each kind within a group must be unique")
)

type inlineValidator struct {
	names         map[string]bool
	kindsInGroups map[string][]string
}

// validate validates a workload inline during processing.
func (validator *inlineValidator) validate(workload kinds.WorkloadBuilder, processor *Processor) error {
	// validate that a workload is a struct
	validate := structvalidator.New()
	if err := validate.Struct(workload); err != nil {
		return fmt.Errorf("%w", err)
	}

	// validate that we do not have overlapping names
	if err := validator.validateName(workload.GetName()); err != nil {
		return err
	}

	// run through the individual validation for the workload
	if err := workload.Validate(); err != nil {
		return fmt.Errorf("error validating workload at path %s: %w", processor.Path, err)
	}

	// validate that we do not have overlapping kinds within a group
	if err := validator.validateKind(workload.GetAPIGroup(), workload.GetAPIKind()); err != nil {
		return err
	}

	return nil
}

// validateName validates a name inline during processing.
func (validator *inlineValidator) validateName(name string) error {
	// ensure that we do not have overlapping names
	if _, found := validator.names[name]; found {
		return fmt.Errorf("%s name used on multiple workloads - %w", name, ErrUniqueNames)
	}

	return nil
}

// validateKind validates kinds within a particular group during processing.
func (validator *inlineValidator) validateKind(group, kind string) error {
	// ensure that we do not have overlapping kinds within a particular group.  overlapping kinds
	// can exist so long as they are in separate groups.
	existingKinds := validator.kindsInGroups[group]
	if len(existingKinds) == 0 {
		return nil
	}

	for i := range existingKinds {
		if existingKinds[i] == kind {
			return fmt.Errorf("%s already exists in group %s - %w", kind, group, ErrUniqueKinds)
		}
	}

	return nil
}
