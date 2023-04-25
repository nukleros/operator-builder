// Copyright 2023 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package e2e

import (
	"fmt"
	"strings"

	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
	"sigs.k8s.io/kubebuilder/v3/pkg/model/resource"

	"github.com/nukleros/operator-builder/internal/utils"
	"github.com/nukleros/operator-builder/internal/workload/v1/kinds"
)

const (
	e2eTestWorkloadPath = "test/e2e/%s_%s_%s_test.go"
)

var (
	_ machinery.Template = &WorkloadTest{}
)

// WorkloadTest adds API-specific scaffolding for each workload test case.
type WorkloadTest struct {
	machinery.TemplateMixin
	machinery.BoilerplateMixin
	machinery.DomainMixin
	machinery.RepositoryMixin
	machinery.ComponentConfigMixin
	machinery.ResourceMixin

	// input fields
	Builder kinds.WorkloadBuilder

	// template fields
	TesterName                string
	TesterNamespace           string
	TesterSamplePath          string
	TesterCollectionName      string
	TesterCollectionNamespace string
}

func (f *WorkloadTest) SetTemplateDefaults() error {
	// set template fields
	f.TesterNamespace = getTesterNamespace(f.Builder)
	f.TesterSamplePath = getTesterSamplePath(f.Resource)
	f.TesterName = getTesterName(f.Resource)

	if f.Builder.GetCollection() != nil {
		f.TesterCollectionName = getTesterCollectionName(f.Builder.GetCollection())
		f.TesterCollectionNamespace = getTesterNamespace(f.Builder.GetCollection())
	}

	// set interface fields
	f.Path = fmt.Sprintf(
		e2eTestWorkloadPath,
		f.Resource.Group,
		f.Resource.Version,
		strings.ToLower(f.Resource.Kind),
	)

	f.IfExistsAction = machinery.SkipFile
	f.TemplateBody = e2eWorkloadsTemplate

	return nil
}

//nolint:lll
const e2eWorkloadsTemplate = `// +build e2e_test

{{ .Boilerplate }}

package e2e_test

import (
	"fmt"
	"os"

	"github.com/stretchr/testify/require"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	{{ .Resource.ImportAlias }} "{{ .Resource.Path }}"
	"{{ .Resource.Path }}/{{ .Builder.GetPackageName }}"
)

//
// {{ .TesterName }} tests
//
func {{ .TesterName }}ChildrenFuncs(tester *E2ETest) error {
	// TODO: need to run r.GetResources(request) on the reconciler to get the mutated resources
	if len({{ .Builder.GetPackageName }}.CreateFuncs) == 0 {
		return nil
	}

	workload, {{ if .Builder.IsComponent }}collection,{{ end }}err := {{ .Builder.GetPackageName }}.ConvertWorkload(tester.workload{{ if .Builder.IsComponent }},tester.collectionTester.workload){{ else }}){{ end }}
	if err != nil {
		return fmt.Errorf("error in workload conversion; %w", err)
	}

	resourceObjects, err := {{ .Builder.GetPackageName }}.Generate(*workload{{ if .Builder.IsComponent }}, *collection, nil, nil){{ else }}, nil, nil){{ end }}
	if err != nil {
		return fmt.Errorf("unable to create objects in memory; %w", err)
	}

	tester.children = resourceObjects

	return nil
}

func {{ .TesterName }}NewHarness(namespace string) *E2ETest {
	return &E2ETest{
		namespace:          namespace,
		unstructured:       &unstructured.Unstructured{},
		workload:           &{{ .Resource.ImportAlias }}.{{ .Resource.Kind }}{},
		sampleManifestFile: "{{ .TesterSamplePath }}",
		getChildrenFunc:    {{ .TesterName }}ChildrenFuncs,
		logSyntax:          "controllers.{{ .Resource.Group }}.{{ .Resource.Kind }}",
		{{ if .Builder.IsComponent -}}
		collectionTester:   {{ .TesterCollectionName }}NewHarness("{{ .TesterCollectionNamespace }}"),     
		{{ end }}
	}
}

{{ if .Builder.IsCollection -}}
func (tester *E2ETest) {{ .TesterName }}Test(testSuite *E2ECollectionTestSuite) {
{{ else }}
func (tester *E2ETest) {{ .TesterName }}Test(testSuite *E2EComponentTestSuite) {
{{ end -}}
	testSuite.suiteConfig.tests = append(testSuite.suiteConfig.tests, tester)
	tester.suiteConfig = &testSuite.suiteConfig
	require.NoErrorf(testSuite.T(), tester.setup(), "failed to setup test")

	// create the custom resource
	require.NoErrorf(testSuite.T(), testCreateCustomResource(tester), "failed to create custom resource")

	// test the deletion of a child object
	require.NoErrorf(testSuite.T(), testDeleteChildResource(tester), "failed to reconcile deletion of a child resource")

	// test the update of a child object
	// TODO: need immutable fields so that we can predict which managed fields we can modify to test reconciliation
	// see https://github.com/nukleros/operator-builder/issues/24

	// test the update of a parent object
	// TODO: need immutable fields so that we can predict which managed fields we can modify to test reconciliation
	// see https://github.com/nukleros/operator-builder/issues/24

	// test that controller logs do not contain errors
	if os.Getenv("DEPLOY_IN_CLUSTER") == "true" {
		require.NoErrorf(testSuite.T(), testControllerLogsNoErrors(tester.suiteConfig, tester.logSyntax), "found errors in controller logs")
	}
}

{{ if .Builder.IsCollection -}}
func (testSuite *E2ECollectionTestSuite) Test_{{ .TesterName }}() {
{{ else }}
func (testSuite *E2EComponentTestSuite) Test_{{ .TesterName }}() {
{{ end -}}
	tester := {{ .TesterName }}NewHarness("{{ .TesterNamespace }}")
	tester.{{ .TesterName }}Test(testSuite)
}

{{ if and (not .Builder.IsClusterScoped) (not .Builder.IsCollection) }}
func (testSuite *E2EComponentTestSuite) Test_{{ .TesterName }}Multi() {
	tester := {{ .TesterName }}NewHarness("{{ .TesterNamespace }}-2")
	tester.{{ .TesterName }}Test(testSuite)
}
{{ end }}
`

func getTesterSamplePath(r *resource.Resource) string {
	return strings.Join([]string{
		"../..",
		"config",
		"samples",
		fmt.Sprintf(
			"%s_%s_%s.yaml",
			r.Group,
			r.Version,
			utils.ToFileName(r.Kind),
		),
	}, "/",
	)
}

func getTesterNamespace(builder kinds.WorkloadBuilder) (namespace string) {
	if !builder.IsClusterScoped() {
		namespaceElements := []string{
			"test",
			strings.ToLower(builder.GetAPIGroup()),
			strings.ToLower(builder.GetAPIVersion()),
			strings.ToLower(builder.GetAPIKind()),
		}
		namespace = strings.Join(namespaceElements, "-")
	}

	return namespace
}

func getTesterName(r *resource.Resource) string {
	return r.ImportAlias() + r.Kind
}

func getTesterCollectionName(collection *kinds.WorkloadCollection) string {
	return strings.ToLower(collection.Spec.API.Group) +
		strings.ToLower(collection.GetAPIVersion()) +
		collection.Spec.API.Kind
}
