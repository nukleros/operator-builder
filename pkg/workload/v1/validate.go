package v1

//import (
//	"errors"
//	"fmt"
//)
//
//// Validate validates a workload configuration
//func (wc *WorkloadConfig) Validate() error {
//
//	missingFields := []string{}
//
//	// required fields
//	if wc.Spec.Group == "" {
//		missingFields = append(missingFields, "spec.group")
//	}
//	if wc.Spec.Version == "" {
//		missingFields = append(missingFields, "spec.version")
//	}
//	if wc.Spec.Kind == "" {
//		missingFields = append(missingFields, "spec.kind")
//	}
//	if len(missingFields) > 0 {
//		msg := fmt.Sprintf("Missing required fields: %s", missingFields)
//		return errors.New(msg)
//	}
//
//	// children only valid if a collection
//	if !wc.Spec.Collection && len(wc.Spec.Children) > 0 {
//		return errors.New("Cannot define spec.Children if spec.collection is false")
//	}
//
//	return nil
//}
