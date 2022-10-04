// Copyright 2022 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package templates

import (
	"sigs.k8s.io/kubebuilder/v3/pkg/machinery"
)

var _ machinery.Template = &Readme{}

// Readme scaffolds a file that defines the templated README.md instructions for a custom workload.
type Readme struct {
	machinery.TemplateMixin

	RootCmdName string
}

// SetTemplateDefaults implements file.Template.
func (f *Readme) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = "README.md"
	}

	f.IfExistsAction = machinery.OverwriteFile
	f.TemplateBody = readmefileTemplate

	return nil
}

const readmefileTemplate = `A Kubernetes operator built with
[operator-builder](https://github.com/nukleros/operator-builder).

## Local Development & Testing

To install the custom resource/s for this operator, make sure you have a
kubeconfig set up for a test cluster, then run:

    make install

To run the controller locally against a test cluster:

    make run

You can then test the operator by creating the sample manifest/s:

    kubectl apply -f config/samples

To clean up:

    make uninstall

## Deploy the Controller Manager

First, set the image:

    export IMG=myrepo/myproject:v0.1.0

Now you can build and push the image:

    make docker-build
    make docker-push

Then deploy:

    make deploy

To clean up:

    make undeploy

{{ if ne .RootCmdName "" -}}
## Companion CLI

To build the companion CLI:

    make build-cli

The CLI binary will get saved to the bin directory.  You can see the help
message with:

    ./bin/{{ .RootCmdName }} help
{{ end -}}
`
