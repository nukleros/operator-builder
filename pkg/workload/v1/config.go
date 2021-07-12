package v1

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
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

	if standaloneFound && collectionFound {
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
			if err := workload.SetSpecFields(workloadConfig); err != nil {
				return nil, err
			}

			if err := workload.SetResources(workloadConfig); err != nil {
				return nil, err
			}

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
			if err := component.SetSpecFields(component.Spec.ConfigPath); err != nil {
				return nil, err
			}

			if err := component.SetResources(component.Spec.ConfigPath); err != nil {
				return nil, err
			}

			component.SetNames()
			components = append(components, *component)
		}
	}

	// get a list of existing component names in the config
	var componentNames []string
	for _, component := range components {
		componentNames = append(componentNames, component.Name)
	}

	// check the dependencies against the actual components
	for i, component := range components {
		missingDependencies := missingDependencies(component.Spec.Dependencies, componentNames)

		// return an error if any dependencies are not satisfied
		if len(missingDependencies) > 0 {
			msg := fmt.Sprintf(
				"%s component/s listed in dependencies for %s but no component workload config for %s provided",
				missingDependencies,
				component.Name,
				missingDependencies,
			)

			return nil, errors.New(msg)
		}

		// add the component dependencies to the object
		for _, dependency := range component.Spec.Dependencies {
			for _, innerComponent := range components {
				if innerComponent.Name == dependency {
					// add the component object to ComponentDependencies
					components[i].Spec.ComponentDependencies = append(
						components[i].Spec.ComponentDependencies,
						innerComponent,
					)
				}
			}
		}
	}

	if standaloneFound && collectionFound {
		msg := fmt.Sprintf(
			"%s and %s both provided - must provide one *or* the other",
			WorkloadKindStandalone,
			WorkloadKindComponent,
		)

		return nil, errors.New(msg)
	} else if collectionFound {
		if err := workload.SetComponents(&components); err != nil {
			return nil, err
		}
		if err := workload.SetSpecFields(workloadConfig); err != nil {
			return nil, err
		}
	}

	return workload, nil
}

func missingDependencies(expected []string, actual []string) []string {
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

			for _, componentFile := range workload.Spec.ComponentFiles {
				componentPath := filepath.Join(filepath.Dir(workloadConfig), componentFile)

				componentWorkloads, err := parseConfig(componentPath)
				if err != nil {
					return nil, err
				}

				for _, component := range *componentWorkloads {
					cw := component.(*ComponentWorkload)
					cw.Spec.ConfigPath = componentPath
					workloads = append(workloads, cw)
				}
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
