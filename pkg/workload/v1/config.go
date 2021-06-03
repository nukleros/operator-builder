package v1

import (
	"errors"
	"io/ioutil"

	"sigs.k8s.io/yaml"
)

func ProcessInitConfig(standaloneConfig, collectionConfig string) (WorkloadInitializer, error) {

	// must provide standalone OR collection
	if standaloneConfig != "" && collectionConfig != "" {
		return nil, errors.New("Must provide a standalone config OR a collection config, not both")
	}

	if standaloneConfig != "" {

		var workload StandaloneWorkload

		config, err := ioutil.ReadFile(standaloneConfig)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(config, &workload)
		if err != nil {
			return nil, err
		}

		return workload, nil

	} else if collectionConfig != "" {

		var workload WorkloadCollection

		config, err := ioutil.ReadFile(collectionConfig)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(config, &workload)
		if err != nil {
			return nil, err
		}

		return workload, nil

	} else {
		return nil, errors.New("No workload config proviced, must provide a standalone config OR a collection config")
	}

}

func ProcessAPIConfig(standaloneConfig, componentConfig string) (WorkloadAPIBuilder, string, error) {

	// must provide standalone OR collection
	if standaloneConfig != "" && componentConfig != "" {
		return nil, "", errors.New("Must provide a standalone config OR a component config, not both")
	}

	if standaloneConfig != "" {

		var workload StandaloneWorkload

		config, err := ioutil.ReadFile(standaloneConfig)
		if err != nil {
			return nil, "", err
		}
		err = yaml.Unmarshal(config, &workload)
		if err != nil {
			return nil, "", err
		}

		return workload, standaloneConfig, nil

	} else if componentConfig != "" {

		var workload ComponentWorkload

		config, err := ioutil.ReadFile(componentConfig)
		if err != nil {
			return nil, "", err
		}
		err = yaml.Unmarshal(config, &workload)
		if err != nil {
			return nil, "", err
		}

		return workload, componentConfig, nil

	} else {
		return nil, "", errors.New("No workload config proviced, must provide a standalone config OR a component config")
	}
}
