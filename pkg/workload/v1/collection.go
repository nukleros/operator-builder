package v1

import (
	"errors"
	"fmt"
)

func (c WorkloadCollection) Validate() error {

	missingFields := []string{}

	// required fields
	if c.Name == "" {
		missingFields = append(missingFields, "name")
	}
	if c.Spec.Domain == "" {
		missingFields = append(missingFields, "spec.domain")
	}
	if c.Spec.Group == "" {
		missingFields = append(missingFields, "spec.group")
	}
	if c.Spec.Version == "" {
		missingFields = append(missingFields, "spec.version")
	}
	if c.Spec.Kind == "" {
		missingFields = append(missingFields, "spec.kind")
	}
	if len(missingFields) > 0 {
		msg := fmt.Sprintf("Missing required fields: %s", missingFields)
		return errors.New(msg)
	}

	return nil
}

func (c WorkloadCollection) GetDomain() string {
	return c.Spec.Domain
}

func (c WorkloadCollection) HasRootCmdName() bool {
	if c.Spec.CompanionCliRootcmd.Name != "" {
		return true
	} else {
		return false
	}
}

func (c WorkloadCollection) GetRootCmdName() string {
	return c.Spec.CompanionCliRootcmd.Name
}

func (c WorkloadCollection) GetRootCmdDescr() string {
	return c.Spec.CompanionCliRootcmd.Description
}
