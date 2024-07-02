// Copyright 2024 Nukleros
// SPDX-License-Identifier: MIT

package controllergen

import (
	"errors"
	"fmt"

	"sigs.k8s.io/controller-tools/pkg/crd"
	"sigs.k8s.io/controller-tools/pkg/deepcopy"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/controller-tools/pkg/rbac"
	"sigs.k8s.io/controller-tools/pkg/schemapatcher"
	"sigs.k8s.io/controller-tools/pkg/webhook"
)

var (
	ErrGeneratorsMissing = errors.New("no generators specified")
	ErrGeneratorsFailed  = errors.New("not all generators ran successfully")
)

type Generator struct {
	Options  []string
	Registry *markers.Registry
}

// NewObjectGenerator returns a generator instance responsible for generating the DeepCopy methods.  This is the
// old `make generate` and `make manifests` commands in kubebuilder.
func NewObjectGenerator(options []string) (*Generator, error) {
	registry, err := newDefaultRegistry()
	if err != nil {
		return nil, err
	}

	return &Generator{
		Options:  options,
		Registry: registry,
	}, nil
}

// Execute executes the generator.
func (generator *Generator) Execute() error {
	// setup the runtime
	rt, err := genall.FromOptions(generator.Registry, generator.Options)
	if err != nil {
		return fmt.Errorf("unable to setup runtime for generator, %w", err)
	}

	if len(rt.Generators) == 0 {
		return ErrGeneratorsMissing
	}

	// execute
	if hadErrs := rt.Run(); hadErrs {
		// don't obscure the actual error with a bunch of usage
		return ErrGeneratorsFailed
	}

	return nil
}

// WithObjectGeneratorOptions returns an array of strings representing the object generator options.  This is
// equivalent to the old `make generate` command in kubebuilder.
func WithObjectGeneratorOptions(path string) []string {
	return []string{
		fmt.Sprintf(`object:headerFile="%s/hack/boilerplate.go.txt"`, path),
		fmt.Sprintf(`paths="%s/apis/..."`, path),
	}
}

// WithObjectGeneratorOptions returns an array of strings representing the object generator options.  This is
// equivalent to the old `make manifests` command in kubebuilder.
func WithManifestGeneratorOptions(path string) []string {
	return []string{
		`crd:crdVersions=v1`,
		`rbac:roleName=manager-role`,
		`webhook`,
		fmt.Sprintf(`paths="%s/apis/..."`, path),
		fmt.Sprintf(`paths="%s/controllers/..."`, path),
		fmt.Sprintf(`paths="%s/internal/..."`, path),
		fmt.Sprintf(`output:crd:artifacts:config=%s/config/crd/bases`, path),
		fmt.Sprintf(`output:rbac:artifacts:config=%s/config/rbac`, path),
	}
}

// newDefaultRegistry returns a registry with all defaults.  Logic taken from the controller-tools package.  See
// https://github.com/kubernetes-sigs/controller-tools/blob/master/cmd/controller-gen/main.go.
func newDefaultRegistry() (*markers.Registry, error) {
	allGenerators := map[string]genall.Generator{
		"crd":         crd.Generator{},
		"rbac":        rbac.Generator{},
		"object":      deepcopy.Generator{},
		"webhook":     webhook.Generator{},
		"schemapatch": schemapatcher.Generator{},
	}

	output := map[string]genall.OutputRule{
		"dir":       genall.OutputToDirectory(""),
		"artifacts": genall.OutputArtifacts{},
	}

	registry := &markers.Registry{}

	for genName, gen := range allGenerators {
		// make the generator options marker itself
		defn := markers.Must(markers.MakeDefinition(genName, markers.DescribesPackage, gen))
		if err := registry.Register(defn); err != nil {
			return nil, fmt.Errorf("unable to make definition for generator %s for registry, %w", genName, err)
		}

		// make per-generation output rule markers
		for ruleName, rule := range output {
			ruleMarker := markers.Must(markers.MakeDefinition(fmt.Sprintf("output:%s:%s", genName, ruleName), markers.DescribesPackage, rule))
			if err := registry.Register(ruleMarker); err != nil {
				return nil, fmt.Errorf("unable to make rule for generator %s for registry, %w", ruleName, err)
			}
		}
	}

	// make "default output" output rule markers
	for ruleName, rule := range output {
		ruleMarker := markers.Must(markers.MakeDefinition("output:"+ruleName, markers.DescribesPackage, rule))
		if err := registry.Register(ruleMarker); err != nil {
			return nil, fmt.Errorf("unable to make default rule output for generator %s for registry, %w", ruleName, err)
		}
	}

	// add in the common options markers
	if err := genall.RegisterOptionsMarkers(registry); err != nil {
		return nil, fmt.Errorf("unable to register options for registry, %w", err)
	}

	return registry, nil
}
