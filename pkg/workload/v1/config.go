package v1

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"sigs.k8s.io/yaml"
)

func ProcessInitConfig(workloadConfig string) (WorkloadInitializer, error) {

	workloads, err := parseConfig(workloadConfig)
	if err != nil {
		return nil, err
	}

	var workload WorkloadInitializer
	standaloneFound := false
	collectionFound := false

	for _, w := range *workloads {

		switch w.GetWorkloadKind() {
		case WorkloadKindStandalone:
			if standaloneFound {
				msg := fmt.Sprintf(
					"Multiple %s configs provided - must provide only one",
					WorkloadKindStandalone,
				)
				return nil, errors.New(msg)
			}
			workload = w.(*StandaloneWorkload)
			standaloneFound = true

		case WorkloadKindCollection:
			if collectionFound {
				msg := fmt.Sprintf(
					"Multiple %s configs provided - must provide only one",
					WorkloadKindCollection,
				)
				return nil, errors.New(msg)
			}
			workload = w.(*WorkloadCollection)
			collectionFound = true
		}
	}

	if standaloneFound == true && collectionFound == true {
		msg := fmt.Sprintf(
			"%s and %s both provided - must provide one *or* the other",
			WorkloadKindStandalone,
			WorkloadKindComponent,
		)
		return nil, errors.New(msg)
	}

	workload.SetNames()
	return workload, nil
}

func ProcessAPIConfig(workloadConfig string) (WorkloadAPIBuilder, error) {

	workloads, err := parseConfig(workloadConfig)
	if err != nil {
		return nil, err
	}

	var workload WorkloadAPIBuilder
	var components []ComponentWorkload
	standaloneFound := false
	collectionFound := false

	for _, w := range *workloads {

		switch w.GetWorkloadKind() {
		case WorkloadKindStandalone:
			if standaloneFound {
				msg := fmt.Sprintf(
					"Multiple %s configs provided - must provide only one",
					WorkloadKindStandalone,
				)
				return nil, errors.New(msg)
			}

			workload = w.(*StandaloneWorkload)
			workload.SetSpecFields(workloadConfig)
			workload.SetResources(workloadConfig)
			workload.SetNames()
			standaloneFound = true

		case WorkloadKindCollection:
			if collectionFound {
				msg := fmt.Sprintf(
					"Multiple %s configs provided - must provide only one",
					WorkloadKindCollection,
				)
				return nil, errors.New(msg)
			}
			workload = w.(*WorkloadCollection)
			workload.SetNames()
			collectionFound = true

		case WorkloadKindComponent:
			component := w.(*ComponentWorkload)
			component.SetSpecFields(workloadConfig)
			component.SetResources(workloadConfig)
			component.SetNames()
			components = append(components, *component)
		}
	}

	// attach component dependencies
	for i, component := range components {
		for _, dependencyName := range component.Spec.Dependencies {
			for _, comp := range components {
				if comp.Name == dependencyName {
					components[i].Spec.ComponentDependencies = append(
						components[i].Spec.ComponentDependencies,
						comp,
					)
				}
			}
			if len(components[i].Spec.ComponentDependencies) < 1 {
				msg := fmt.Sprintf(
					"%s component listed as a dependency for %s but no component workload config for %s provied",
					dependencyName,
					component.Name,
					dependencyName,
				)
				return nil, errors.New(msg)
			}
		}
	}

	if standaloneFound == true && collectionFound == true {
		msg := fmt.Sprintf(
			"%s and %s both provided - must provide one *or* the other",
			WorkloadKindStandalone,
			WorkloadKindComponent,
		)
		return nil, errors.New(msg)
	} else if collectionFound == true {
		if err := workload.SetComponents(&components); err != nil {
			return nil, err
		}
	}

	return workload, nil
}

func parseConfig(workloadConfig string) (*[]WorkloadIdentifier, error) {

	if workloadConfig == "" {
		return nil, errors.New("No workload config provided - workload config required")
	}

	configContent, err := ioutil.ReadFile(workloadConfig)
	if err != nil {
		return nil, err
	}

	var configs []string

	lines := strings.Split(string(configContent), "\n")

	var config string
	for _, line := range lines {
		if strings.TrimSpace(line) == "---" {
			if len(config) > 0 {
				configs = append(configs, config)
				config = ""
			}

		} else {
			config = config + "\n" + line
		}
	}
	if len(config) > 0 {
		configs = append(configs, config)
	}

	var workloads []WorkloadIdentifier
	var workloadNames []string

	for _, c := range configs {

		var workloadID WorkloadShared

		err := yaml.Unmarshal([]byte(c), &workloadID)
		if err != nil {
			return nil, err
		}

		for _, n := range workloadNames {
			if workloadID.Name == n {
				msg := fmt.Sprintf(
					"%s name used on multiple workloads - each workload name must be unique",
					workloadID.Name,
				)
				return nil, errors.New(msg)
			}
		}
		workloadNames = append(workloadNames, workloadID.Name)

		switch workloadID.Kind {
		case WorkloadKindStandalone:

			workload := &StandaloneWorkload{}

			err = yaml.Unmarshal([]byte(c), workload)
			if err != nil {
				return nil, err
			}

			workloads = append(workloads, workload)

		case WorkloadKindComponent:

			workload := &ComponentWorkload{}

			err = yaml.Unmarshal([]byte(c), workload)
			if err != nil {
				return nil, err
			}

			workloads = append(workloads, workload)

		case WorkloadKindCollection:

			workload := &WorkloadCollection{}

			err = yaml.Unmarshal([]byte(c), workload)
			if err != nil {
				return nil, err
			}

			workloads = append(workloads, workload)

		default:
			msg := fmt.Sprintf(
				"Unrecognized workload kind in workload config - valid kinds: %s, %s, %s,",
				WorkloadKindStandalone,
				WorkloadKindCollection,
				WorkloadKindComponent,
			)
			return nil, errors.New(msg)
		}
	}

	return &workloads, nil
}
