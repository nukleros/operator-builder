// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"

	"github.com/nukleros/operator-builder/internal/workload/v1/kinds"
)

var ErrConfigMustExist = errors.New("no workload config provided - workload config required")

// Processor is an object that stores information necessary for generating object source
// code.
type Processor struct {
	Path string

	// Workload represents the top-level configuration (e.g. as passed in via the --workload-config) flag
	// from the command line, while Children represents subordinate configurations that the parent files such
	// as the componentFiles field.
	Workload kinds.WorkloadBuilder
	Children []*Processor
}

// NewProcessor will return a new workload config processor given a path.  An error is returned if the workload config
// does not exist at a path.
func NewProcessor(configPath string) (*Processor, error) {
	if configPath == "" {
		return nil, ErrConfigMustExist
	}

	return &Processor{Path: configPath}, nil
}

// GetWorkloads gets all of the workloads for a config processor in a flattened
// fashion.
func (processor *Processor) GetWorkloads() []kinds.WorkloadBuilder {
	workloads := []kinds.WorkloadBuilder{processor.Workload}

	// return the array with the single workload if we have no children
	if len(processor.Children) == 0 {
		return workloads
	}

	// get the workloads for a child and append it to the array
	for i := range processor.Children {
		workloads = append(workloads, processor.Children[i].GetWorkloads()...)
	}

	return workloads
}

// GetProcessors gets all of the processors to include the parent and children processors.
func (processor *Processor) GetProcessors() []*Processor {
	processors := []*Processor{processor}

	// return array with single processor if we have no children
	if len(processor.Children) == 0 {
		return processors
	}

	// get the processors for a child and append it to the array
	for i := range processor.Children {
		processors = append(processors, processor.Children[i].GetProcessors()...)
	}

	return processors
}
