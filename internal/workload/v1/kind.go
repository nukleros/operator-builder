// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package v1

import (
	"errors"

	"gopkg.in/yaml.v3"
)

var ErrInvalidKind = errors.New("unrecognized workload kind in workload config")

// WorkloadKind indicates which of the supported workload kinds are being used.
type WorkloadKind int32

const (
	WorkloadKindUnknown WorkloadKind = iota
	WorkloadKindStandalone
	WorkloadKindCollection
	WorkloadKindComponent
)

func (wk WorkloadKind) String() string {
	kinds := map[WorkloadKind]string{
		WorkloadKindStandalone: "StandaloneWorkload",
		WorkloadKindCollection: "WorkloadCollection",
		WorkloadKindComponent:  "ComponentWorkload",
		WorkloadKindUnknown:    "Unknown Workload Type",
	}

	return kinds[wk]
}

func (wk *WorkloadKind) UnmarshalJSON(data []byte) error {
	kinds := map[string]WorkloadKind{
		"StandaloneWorkload": WorkloadKindStandalone,
		"WorkloadCollection": WorkloadKindCollection,
		"ComponentWorkload":  WorkloadKindComponent,
	}

	kind, ok := kinds[string(data)]
	if !ok {
		return ErrInvalidKind
	}

	*wk = kind

	return nil
}

func (wk *WorkloadKind) UnmarshalYAML(node *yaml.Node) error {
	kinds := map[string]WorkloadKind{
		"StandaloneWorkload": WorkloadKindStandalone,
		"WorkloadCollection": WorkloadKindCollection,
		"ComponentWorkload":  WorkloadKindComponent,
	}

	kind, ok := kinds[node.Value]
	if !ok {
		return ErrInvalidKind
	}

	*wk = kind

	return nil
}
