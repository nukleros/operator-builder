// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/go-playground/validator"
	"gopkg.in/yaml.v3"
)

var (
	ErrNamesMustBeUnique   = errors.New("each workload name must be unique")
	ErrConfigMustExist     = errors.New("no workload config provided - workload config required")
	ErrInvalidKind         = errors.New("unrecognized workload kind in workload config")
	ErrMultipleConfigs     = errors.New("multiple configs found - please provide only one standalone or collection workload")
	ErrCollectionRequired  = errors.New("a WorkloadCollection is required when using WorkloadComponents")
	ErrMissingWorkload     = errors.New("could not find either standalone or collection workload, please provide one")
	ErrMissingDependencies = errors.New("missing dependencies - no workload config provided")
)

func ProcessInitConfig(workloadConfig string) (WorkloadInitializer, error) {
	workloads, err := parseConfig(workloadConfig)
	if err != nil {
		return nil, err
	}

	if len(workloads[WorkloadKindComponent]) != 0 && len(workloads[WorkloadKindCollection]) != 1 {
		return nil, fmt.Errorf("no %s found - %w", WorkloadKindCollection, ErrCollectionRequired)
	}

	var workload WorkloadInitializer

	for k := range workloads {
		for _, w := range workloads[k] {
			switch v := w.(type) {
			case *StandaloneWorkload:
				workload = v
			case *WorkloadCollection:
				workload = v
			case *ComponentWorkload:
				continue
			}
		}
	}

	workload.SetNames()

	return workload, nil
}

//nolint:gocyclo // this will be refactored later
func ProcessAPIConfig(workloadConfig string) (WorkloadAPIBuilder, error) {
	workloads, err := parseConfig(workloadConfig)
	if err != nil {
		return nil, err
	}

	if len(workloads[WorkloadKindComponent]) != 0 && len(workloads[WorkloadKindCollection]) != 1 {
		return nil, fmt.Errorf("no %s found - %w", WorkloadKindCollection, ErrCollectionRequired)
	}

	var workload WorkloadAPIBuilder

	var components []*ComponentWorkload

	for kind := range workloads {
		for _, w := range workloads[kind] {
			switch v := w.(type) {
			case *StandaloneWorkload:
				workload = v
				if err := workload.SetResources(workloadConfig); err != nil {
					return nil, fmt.Errorf("%w", err)
				}

				workload.SetNames()
			case *WorkloadCollection:
				workload = v
				if err := workload.SetResources(workloadConfig); err != nil {
					return nil, fmt.Errorf("%w", err)
				}

				workload.SetNames()
			case *ComponentWorkload:
				if err := v.SetResources(v.Spec.ConfigPath); err != nil {
					return nil, err
				}

				v.SetNames()
				components = append(components, v)
			}
		}
	}

	if err := handleDependencies(&components); err != nil {
		return nil, err
	}

	if len(workloads[WorkloadKindCollection]) == 1 {
		if err := workload.SetComponents(components); err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		if err := workload.SetResources(workloadConfig); err != nil {
			return nil, fmt.Errorf("%w", err)
		}
	}

	return workload, nil
}

func missingDependencies(expected, actual []string) []string {
	var missing []string

	for _, expectedDependency := range expected {
		var found bool

		for _, actualDependency := range actual {
			if actualDependency == expectedDependency {
				found = true

				break
			}
		}

		if !found {
			missing = append(missing, expectedDependency)
		}
	}

	return missing
}

func parseConfig(workloadConfig string) (map[WorkloadKind][]WorkloadIdentifier, error) {
	if workloadConfig == "" {
		return nil, ErrConfigMustExist
	}

	file, err := ReadStream(workloadConfig)
	if err != nil {
		return nil, err
	}

	defer CloseFile(file)

	var kindReader bytes.Buffer
	reader := io.TeeReader(file, &kindReader)

	sharedDecoder := yaml.NewDecoder(reader)

	kindDecoder := yaml.NewDecoder(&kindReader)
	kindDecoder.KnownFields(true)

	workloads := make(map[WorkloadKind][]WorkloadIdentifier)

	workloadMap := make(map[string]bool)

	for {
		var workloadID WorkloadShared

		if err := sharedDecoder.Decode(&workloadID); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", workloadConfig, err)
		}

		if _, found := workloadMap[workloadID.Name]; found {
			return nil, fmt.Errorf(
				"%s name used on multiple workloads - %w",
				workloadID.Name,
				ErrNamesMustBeUnique,
			)
		}

		workloadMap[workloadID.Name] = true

		workload, err := decodeKind(workloadID.Kind, kindDecoder)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", workloadConfig, err)
		}

		workloads[workload.GetWorkloadKind()] = append(workloads[workload.GetWorkloadKind()], workload)

		if collection, ok := workload.(*WorkloadCollection); ok {
			cws, err := parseCollectionComponents(collection, workloadConfig)
			if err != nil {
				return nil, err
			}

			workloads[WorkloadKindComponent] = append(workloads[WorkloadKindComponent], cws...)
		}
	}

	if err := validateConfigs(workloads); err != nil {
		return nil, err
	}

	return workloads, nil
}

func parseCollectionComponents(workload *WorkloadCollection, workloadConfig string) ([]WorkloadIdentifier, error) {
	var workloads []WorkloadIdentifier

	for _, componentFile := range workload.Spec.ComponentFiles {
		componentPath := filepath.Join(filepath.Dir(workloadConfig), componentFile)

		w, err := parseConfig(componentPath)
		if err != nil {
			return nil, err
		}

		for _, component := range w[WorkloadKindComponent] {
			if cw, ok := component.(*ComponentWorkload); ok {
				cw.Spec.ConfigPath = componentPath
				workloads = append(workloads, cw)
			}
		}
	}

	return workloads, nil
}

func decodeKind(kind WorkloadKind, dc *yaml.Decoder) (WorkloadIdentifier, error) {
	switch kind {
	case WorkloadKindStandalone:
		v := &StandaloneWorkload{}
		if err := dc.Decode(v); err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return v, nil
	case WorkloadKindComponent:
		v := &ComponentWorkload{}
		if err := dc.Decode(v); err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return v, nil
	case WorkloadKindCollection:
		v := &WorkloadCollection{}
		if err := dc.Decode(v); err != nil {
			return nil, fmt.Errorf("%w", err)
		}

		return v, nil
	default:
		return nil, fmt.Errorf(
			"%w - valid kinds: %s, %s, %s,",
			ErrInvalidKind,
			WorkloadKindStandalone,
			WorkloadKindCollection,
			WorkloadKindComponent,
		)
	}
}

func handleDependencies(components *[]*ComponentWorkload) error {
	c := *components
	// get a list of existing component names in the config
	componentNames := make([]string, len(c))
	for i := range c {
		componentNames[i] = c[i].Name
	}

	// check the dependencies against the actual components
	for i := range c {
		missingDependencies := missingDependencies(c[i].Spec.Dependencies, componentNames)

		// return an error if any dependencies are not satisfied
		if len(missingDependencies) > 0 {
			return fmt.Errorf(
				"%w for components(s) %s listed as dependencies for %s",
				ErrMissingDependencies,
				missingDependencies,
				c[i].Name,
			)
		}

		// add the component dependencies to the object
		for _, dependency := range c[i].Spec.Dependencies {
			for j := range c {
				if c[j].Name == dependency {
					// add the component object to ComponentDependencies
					c[i].Spec.ComponentDependencies = append(
						c[i].Spec.ComponentDependencies,
						c[j],
					)
				}
			}
		}
	}

	*components = c

	return nil
}

func validateConfigs(workloads map[WorkloadKind][]WorkloadIdentifier) error {
	validate := validator.New()

	for kind := range workloads {
		for _, w := range workloads[kind] {
			if err := validate.Struct(w); err != nil {
				return fmt.Errorf("%w", err)
			}
		}
	}

	if len(workloads[WorkloadKindCollection])+len(workloads[WorkloadKindStandalone]) > 1 {
		return ErrMultipleConfigs
	}

	return nil
}
