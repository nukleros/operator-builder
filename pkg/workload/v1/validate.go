package v1

import (
	"errors"
	"fmt"
)

// Validate validates a workload configuration
func (w Workload) Validate() error {

	missingFields := []string{}

	// required fields
	if w.Spec.Group == "" {
		missingFields = append(missingFields, "spec.group")
	}
	if w.Spec.Version == "" {
		missingFields = append(missingFields, "spec.version")
	}
	if w.Spec.Kind == "" {
		missingFields = append(missingFields, "spec.kind")
	}
	if len(missingFields) > 0 {
		msg := fmt.Sprintf("Missing required fields: %s", missingFields)
		return errors.New(msg)
	}

	// children only valid if a collection
	if !w.Spec.Collection && len(w.Spec.Children) > 0 {
		return errors.New("Cannot define spec.Children if spec.collection is false")
	}

	return nil
}
