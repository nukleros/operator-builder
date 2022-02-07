// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

var (
	ErrFileExists  = errors.New("force was not requested and file exists")
	ErrWriteFile   = errors.New("unable to write to file")
	ErrWriteStdout = errors.New("unable to write to stdout")
)

const (
	indentLevel = 2
	permissions = 0644
)

type InitConfigOptions struct {
	Path           string
	Force          bool
	WorkloadConfig WorkloadInitializer
}

func WriteConfig(options *InitConfigOptions) error {
	// validate the configuration
	if err := options.WorkloadConfig.Validate(); err != nil {
		return fmt.Errorf("%w; invalid configuration", err)
	}

	// mutate typed fields
	yamlBytes, err := mutateConfig(options.WorkloadConfig)
	if err != nil {
		return fmt.Errorf("%w; unable to mutate configuration fields", err)
	}

	var yamlNode yaml.Node

	if err := yaml.Unmarshal(yamlBytes, &yamlNode); err != nil {
		return fmt.Errorf("%w; unable to unmarshal workload config as yaml.Node", err)
	}

	buf := new(bytes.Buffer)
	yamlEncoder := yaml.NewEncoder(buf)
	yamlEncoder.SetIndent(indentLevel)

	if err := yamlEncoder.Encode(&yamlNode); err != nil {
		return fmt.Errorf("%w; unable to encode yaml", err)
	}

	return outputFile(options, buf.Bytes())
}

func outputFile(options *InitConfigOptions, data []byte) error {
	// if we have requested stdout, write to stdout
	if options.Path == "-" {
		outputStream := os.Stdout

		if _, err := outputStream.WriteString(string(data)); err != nil {
			return fmt.Errorf("%w; %s", err, ErrWriteStdout)
		}

		return nil
	}

	// if the file exists without a force request, we return an error
	if _, err := os.Stat(options.Path); err == nil {
		if !options.Force {
			return fmt.Errorf("%w at location %s", ErrFileExists, options.Path)
		}
	}

	if err := os.WriteFile(options.Path, data, permissions); err != nil {
		return fmt.Errorf("%w; %s at location %s", err, ErrWriteFile, options.Path)
	}

	return nil
}

func mutateConfig(workloadConfig WorkloadInitializer) ([]byte, error) {
	// convert to bytes
	rawData, err := yaml.Marshal(&workloadConfig)
	if err != nil {
		return nil, fmt.Errorf("%w; unable to marshal workload config as []byte", err)
	}

	// convert to map[string]interface so that we can manipulate the fields
	// outside of their proper types
	var yamlData map[string]interface{}

	if err := yaml.Unmarshal(rawData, &yamlData); err != nil {
		return nil, fmt.Errorf("%w; unable to unmarshal workload config as yaml.Node", err)
	}

	// reset the kind field to a string
	yamlData["kind"] = workloadConfig.GetWorkloadKind().String()

	// convert the resources field to a flat array instead of an array of objects
	if err := mutateResources(yamlData); err != nil {
		return nil, fmt.Errorf("%w; unable to convert resources object to string array", err)
	}

	// convert to bytes
	yamlBytes, err := yaml.Marshal(yamlData)
	if err != nil {
		return nil, fmt.Errorf("%w; unable to marshal workload config as []byte", err)
	}

	return yamlBytes, nil
}

func mutateResources(yamlData map[string]interface{}) error {
	specField, err := toMapStringInterface(yamlData["spec"])
	if err != nil {
		return fmt.Errorf("%w; error converting workload config spec %v", err, yamlData["spec"])
	}

	resourceObjs, err := toArrayInterface(specField["resources"])
	if err != nil {
		return fmt.Errorf("%w; error converting spec.resources %v", err, specField["resources"])
	}

	resources := make([]string, len(resourceObjs))

	for i, resource := range resourceObjs {
		resourceMap, err := toMapStringInterface(resource)
		if err != nil {
			return fmt.Errorf("%w; error converting spec.resources item %v", err, resource)
		}

		filename, err := toString(resourceMap["filename"])
		if err != nil {
			return fmt.Errorf("%w; error converting spec.resources.filename %v", err, resource)
		}

		resources[i] = filename
	}

	specField["resources"] = resources

	return nil
}
