package v1

import (
	"errors"
	"fmt"
	"io/ioutil"

	"sigs.k8s.io/yaml"
)

func ProcessInitConfig(workloadConfig string) (WorkloadInitializer, error) {

	kind, err := getKind(workloadConfig)
	if err != nil {
		return nil, err
	}

	switch kind {
	case WorkloadKindStandalone:

		var workload StandaloneWorkload

		config, err := ioutil.ReadFile(workloadConfig)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(config, &workload)
		if err != nil {
			return nil, err
		}

		return workload, nil

	case WorkloadKindCollection:

		var workload WorkloadCollection

		config, err := ioutil.ReadFile(workloadConfig)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(config, &workload)
		if err != nil {
			return nil, err
		}

		return workload, nil

	default:
		msg := fmt.Sprintf(
			"Project initialization requires a %s or %s workload config",
			WorkloadKindStandalone,
			WorkloadKindCollection,
		)
		return nil, errors.New(msg)
	}
}

func ProcessAPIConfig(workloadConfig string) (WorkloadAPIBuilder, error) {

	kind, err := getKind(workloadConfig)
	if err != nil {
		return nil, err
	}

	switch kind {
	case WorkloadKindStandalone:

		var workload StandaloneWorkload

		config, err := ioutil.ReadFile(workloadConfig)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(config, &workload)
		if err != nil {
			return nil, err
		}

		return workload, nil

	case WorkloadKindComponent:

		var workload ComponentWorkload

		config, err := ioutil.ReadFile(workloadConfig)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(config, &workload)
		if err != nil {
			return nil, err
		}

		return workload, nil

	default:
		msg := fmt.Sprintf(
			"API creation requires a %s or %s workload config",
			WorkloadKindStandalone,
			WorkloadKindComponent,
		)
		return nil, errors.New(msg)
	}
}

func getKind(workloadConfig string) (WorkloadKind, error) {

	if workloadConfig == "" {
		return "", errors.New("No workload config provided - workload config required")
	}

	workload := struct {
		Kind WorkloadKind `json:"kind"`
	}{}

	config, err := ioutil.ReadFile(workloadConfig)
	if err != nil {
		return "", err
	}
	err = yaml.Unmarshal(config, &workload)
	if err != nil {
		return "", err
	}

	switch workload.Kind {
	case WorkloadKindStandalone:
		return WorkloadKindStandalone, nil
	case WorkloadKindCollection:
		return WorkloadKindCollection, nil
	case WorkloadKindComponent:
		return WorkloadKindComponent, nil
	default:
		msg := fmt.Sprintf(
			"Unrecognized workload kind in workload config - valid kinds: %s, %s, %s,",
			WorkloadKindStandalone,
			WorkloadKindCollection,
			WorkloadKindComponent,
		)
		return "", errors.New(msg)
	}
}
