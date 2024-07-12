// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/nukleros/operator-builder/internal/utils"
	"github.com/nukleros/operator-builder/internal/workload/v1/kinds"
)

var (
	ErrParseConfig          = errors.New("error parsing workload config")
	ErrParseComponentConfig = errors.New("error parsing workload component config")
	ErrConvertComponent     = errors.New("error converting workload to component workload")
	ErrConvertCollection    = errors.New("error converting workload to collection")
	ErrCollectionRequired   = errors.New("a WorkloadCollection is required when using WorkloadComponents")
	ErrMultipleConfigs      = errors.New("multiple configs found - please provide only one standalone or collection workload")
	ErrMissingWorkload      = errors.New("could not find either standalone or collection workload, please provide one")
	ErrMissingDependencies  = errors.New("missing dependencies - no workload config provided")
)

// Parse will parse and individual workload config given the path at which it exists.  It also
// returns the processor and all of its attributes that were parsed and set during parsing.
func Parse(configPath string) (*Processor, error) {
	processor, err := NewProcessor(configPath)
	if err != nil {
		return nil, fmt.Errorf("%s - error creating new processor - %w", ErrParseConfig.Error(), err)
	}

	// create the validator we need to track workloads as they are parsed and fail fast
	validator := &inlineValidator{
		names:         make(map[string]bool),
		kindsInGroups: make(map[string][]string),
	}

	// parse the workload configuration from the newly created object
	if err := processor.parse(validator); err != nil {
		return nil, fmt.Errorf("%s at config path %s - %w", ErrParseConfig.Error(), configPath, err)
	}

	// we must ensure the top-level workload that we created is not a component, as the top-level
	// workload should always represent either a collection or a standalone
	if processor.Workload.IsComponent() {
		return nil, fmt.Errorf(
			"%s - no %s found at config path %s - %w",
			ErrParseConfig.Error(),
			kinds.WorkloadKindCollection,
			configPath,
			ErrCollectionRequired,
		)
	}

	// finally, we must ensure that any dependencies specified exist within this configuration
	// bundle.
	for _, component := range processor.Children {
		if err := setDependencies(component.Workload, processor.GetWorkloads()); err != nil {
			return nil, fmt.Errorf("%w; unable to set dependencies for component: %s", err, component.Workload.GetName())
		}
	}

	return processor, nil
}

// parse will parse a given workload config into its appropriate workload object
// definitions.
func (processor *Processor) parse(validator *inlineValidator) error {
	file, err := utils.ReadStream(processor.Path)
	if err != nil {
		return fmt.Errorf("%w; error reading file %s", err, processor.Path)
	}

	defer utils.CloseFile(file)

	var kindReader bytes.Buffer
	reader := io.TeeReader(file, &kindReader)

	sharedDecoder, kindDecoder := yaml.NewDecoder(reader), yaml.NewDecoder(&kindReader)

	kindDecoder.KnownFields(true)

	for {
		var workloadID kinds.WorkloadShared

		// decode the particular kind
		if err := sharedDecoder.Decode(&workloadID); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return fmt.Errorf("failed to read file %s: %w", processor.Path, err)
		}

		// decode the particular kind into its appropriate workload
		workload, err := kinds.Decode(workloadID.Kind, kindDecoder)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", processor.Path, err)
		}

		// perform validation of the configuration after decoding the config object
		if err := validator.validate(workload, processor); err != nil {
			return err
		}

		// update the inline validator so we can appropriate validate future items in the loop
		validator.names[workloadID.Name] = true
		validator.kindsInGroups[workload.GetAPIGroup()] = append(
			validator.kindsInGroups[workload.GetAPIGroup()], workload.GetAPIKind(),
		)

		// update the processor for this workload
		workload.SetNames()
		processor.Workload = workload

		// parse the child components
		if workload.IsCollection() {
			collection, ok := workload.(*kinds.WorkloadCollection)
			if !ok {
				return fmt.Errorf("%w for workload %s labeled as collection", ErrConvertCollection, workload.GetName())
			}

			if err := processor.parseComponents(collection, processor.Path, validator); err != nil {
				return err
			}
		}
	}

	return nil
}

func (processor *Processor) parseComponents(
	workload *kinds.WorkloadCollection,
	workloadConfig string,
	validator *inlineValidator,
) error {
	for _, componentFile := range workload.Spec.ComponentFiles {
		// get each of the component paths for a glob pattern
		componentPaths, err := utils.Glob(filepath.Join(filepath.Dir(workloadConfig), componentFile))
		if err != nil {
			return fmt.Errorf("%w; error globbing workload config at path %s", err, componentFile)
		}

		// parse each file path from the glob
		for _, componentPath := range componentPaths {
			componentProcessor, err := NewProcessor(componentPath)
			if err != nil {
				return err
			}

			// add the component processor as a child
			processor.Children = append(processor.Children, componentProcessor)

			if err := componentProcessor.parse(validator); err != nil {
				return fmt.Errorf("%w; %s at path %s", err, ErrParseComponentConfig.Error(), componentPath)
			}

			// set the config path
			c := componentProcessor.Workload
			if component, ok := c.(*kinds.ComponentWorkload); ok {
				component.Spec.ConfigPath = componentPath
			}
		}
	}

	return nil
}

// setDependencies will set the dependencies for a particular workload.
func setDependencies(workload kinds.WorkloadBuilder, workloads []kinds.WorkloadBuilder) error {
	component, ok := workload.(*kinds.ComponentWorkload)
	if !ok {
		return fmt.Errorf("%w for workload [%s]", ErrConvertComponent, workload.GetName())
	}

	component.Spec.ComponentDependencies = []*kinds.ComponentWorkload{}

	missing := []string{}

	for _, expected := range component.Spec.Dependencies {
		if dependency := getDependency(expected, workloads); dependency != nil {
			component.Spec.ComponentDependencies = append(
				component.Spec.ComponentDependencies,
				dependency,
			)
		} else {
			missing = append(missing, expected)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("%w; missing [%v] for component: [%s]", ErrMissingDependencies, missing, component.Name)
	}

	return nil
}

// getDependency returns a dependency as a component workload.
func getDependency(name string, workloads []kinds.WorkloadBuilder) *kinds.ComponentWorkload {
	for this := range workloads {
		if workloads[this].GetName() == name {
			component, ok := workloads[this].(*kinds.ComponentWorkload)
			if !ok {
				return nil
			}

			return component
		}
	}

	return nil
}
