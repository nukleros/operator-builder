// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package subcommand

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/config"
	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/kinds"
	"github.com/vmware-tanzu-labs/operator-builder/internal/workload/v1/markers"
)

var (
	ErrCreateAPIPreProcess    = errors.New("error pre-processing `create api` subcommand")
	ErrCreateAPIProcess       = errors.New("error processing `create api` subcommand")
	ErrCreateAPISetComponents = errors.New("error setting components for `create api` subcommand")
)

type createAPIProcessor struct {
	components       []*kinds.ComponentWorkload
	collection       *kinds.WorkloadCollection
	configProcessors []*config.Processor
}

// CreateAPI runs through the logic that happens when the `create api` subcommand is executed.  It is responsible
// for the processing of manifests and the markers within them, generating source code, and setting the values
// used during scaffolding.
func CreateAPI(processor *config.Processor) error {
	// run through pre-processing to gather the collection and the components
	apiProcessor := &createAPIProcessor{configProcessors: processor.GetProcessors()}
	if err := apiProcessor.preProcess(); err != nil {
		return fmt.Errorf("%w; %s", err, ErrCreateAPIPreProcess)
	}

	if len(apiProcessor.components) > 0 {
		if err := processor.Workload.SetComponents(apiProcessor.components); err != nil {
			return fmt.Errorf("%w; %s", err, ErrCreateAPISetComponents)
		}
	}

	// run through processing
	if err := apiProcessor.process(); err != nil {
		return fmt.Errorf("%w; %s", err, ErrCreateAPIProcess)
	}

	return nil
}

func (apiProcessor *createAPIProcessor) preProcess() error {
	for _, processor := range apiProcessor.configProcessors {
		// load the manifests for the workload
		if err := processor.Workload.LoadManifests(filepath.Dir(processor.Path)); err != nil {
			return fmt.Errorf("%w; error loading manifests for workload %s", err, processor.Workload.GetName())
		}

		// find the collection and the components
		switch workload := processor.Workload.(type) {
		case *kinds.StandaloneWorkload:
			continue
		case *kinds.WorkloadCollection:
			// a collection is still a collection to itself
			apiProcessor.collection = workload
			workload.Spec.Collection = workload
			workload.Spec.ForCollection = true
		case *kinds.ComponentWorkload:
			// add this component to the list of components
			apiProcessor.components = append(apiProcessor.components, workload)
		}
	}

	return nil
}

func (apiProcessor *createAPIProcessor) process() error {
	fieldMarkers := &markers.MarkerCollection{
		FieldMarkers:           []*markers.FieldMarker{},
		CollectionFieldMarkers: []*markers.CollectionFieldMarker{},
	}

	workloadSpecs := make([]*kinds.WorkloadSpec, len(apiProcessor.configProcessors))

	// set the resources and collect the markers and specs
	for i := range apiProcessor.configProcessors {
		// get the spec and set the collection on the components
		switch workload := apiProcessor.configProcessors[i].Workload.(type) {
		case *kinds.StandaloneWorkload:
			workloadSpecs[i] = &workload.Spec.WorkloadSpec
		case *kinds.WorkloadCollection:
			workloadSpecs[i] = &workload.Spec.WorkloadSpec
		case *kinds.ComponentWorkload:
			workloadSpecs[i] = &workload.Spec.WorkloadSpec
			workload.Spec.Collection = apiProcessor.collection
			workload.Spec.API.Domain = apiProcessor.collection.Spec.API.Domain
		}

		if err := apiProcessor.configProcessors[i].Workload.SetResources(apiProcessor.configProcessors[i].Path); err != nil {
			return fmt.Errorf(
				"%w; error setting resources for workload %s",
				err, apiProcessor.configProcessors[i].Workload.GetName(),
			)
		}

		apiProcessor.configProcessors[i].Workload.SetRBAC()

		fieldMarkers.FieldMarkers = append(fieldMarkers.FieldMarkers, workloadSpecs[i].FieldMarkers...)
		fieldMarkers.CollectionFieldMarkers = append(fieldMarkers.CollectionFieldMarkers, workloadSpecs[i].CollectionFieldMarkers...)
	}

	// loop through the collected workload specs and process the resource markers
	for i := range workloadSpecs {
		if err := workloadSpecs[i].ProcessResourceMarkers(fieldMarkers); err != nil {
			return fmt.Errorf("%w; error processing resource markers", err)
		}
	}

	return nil
}
