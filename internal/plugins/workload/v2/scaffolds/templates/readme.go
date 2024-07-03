// Copyright 2024 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package templates

import (
	"errors"
	"fmt"
	"strings"

	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"
)

const (
	missingVersionTag = "latest"
)

var _ machinery.Template = &Readme{}

var ErrInvalidImage = errors.New("invalid image")

// Readme scaffolds a file that defines the templated README.md instructions for a custom workload.
type Readme struct {
	machinery.TemplateMixin

	RootCmdName         string
	EnableOLM           bool
	ControllerImg       string
	ControllerBundleImg string
}

// SetTemplateDefaults implements file.Template.
func (f *Readme) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = "README.md"
	}

	controllerImgParts := strings.Split(f.ControllerImg, ":")
	switch len(controllerImgParts) {
	case 1:
		f.ControllerImg = fmt.Sprintf("%s:%s", controllerImgParts[0], missingVersionTag)
		f.ControllerBundleImg = fmt.Sprintf("%s-bundle:%s", controllerImgParts[0], missingVersionTag)
	case 2:
		f.ControllerImg = fmt.Sprintf("%s:%s", controllerImgParts[0], controllerImgParts[1])
		f.ControllerBundleImg = fmt.Sprintf("%s-bundle:%s", controllerImgParts[0], controllerImgParts[1])
	default:
		return fmt.Errorf("%s; %w", f.ControllerImg, ErrInvalidImage)
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

    export IMG={{ .ControllerImg }}

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
{{- end }}

{{ if .EnableOLM -}}
## Deploy the Operator Lifecycle Manager Bundle

First, build the bundle.  The bundle contains metadata that makes it 
compatible with Operator Lifecycle Manager and also makes the operator 
importable into OpenShift OperatorHub:

    make bundle

Next, set the bundle image.  This is the image that contains the packaged 
bundle:

    export BUNDLE_IMG={{ .ControllerBundleImg }}

Now you can build and push the bundle image:

    make bundle-build
    make bundle-push

To deploy the bundle (requires OLM to be running in the cluster):

    make operator-sdk
    bin/operator-sdk bundle validate $BUNDLE_IMG
    bin/operator-sdk run bundle $BUNDLE_IMG

To clean up:

    bin/operator-sdk cleanup --delete-all $BUNDLE_IMG
{{ end -}}
`
