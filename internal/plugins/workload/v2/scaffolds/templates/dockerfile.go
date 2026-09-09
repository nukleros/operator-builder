// Copyright 2024 Nukleros
// SPDX-License-Identifier: Apache-2.0

package templates

import (
	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"
)

const (
	defaultDockerfilePath = "Dockerfile"
)

var _ machinery.Template = &Dockerfile{}

// Dockerfile scaffolds a file that defines the containerized build process.
type Dockerfile struct {
	machinery.TemplateMixin
}

// SetTemplateDefaults implements file.Template.
func (f *Dockerfile) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = defaultDockerfilePath
	}

	f.IfExistsAction = machinery.OverwriteFile
	f.TemplateBody = dockerfileTemplate

	return nil
}

const dockerfileTemplate = `# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
#
# NOTE: this expects a pre-built "manager" binary in the build context (see
# the "docker-build" target in the Makefile) rather than compiling it here,
# since release tooling such as goreleaser's docker pipe builds this image
# from a context that only contains the release binary, not the full repo
# source.
#
# TARGETOS/TARGETARCH are unused by default (goreleaser only ever builds one
# platform per image) but are declared here so "make docker-buildx" can
# rewrite the COPY line below to pull in a per-platform binary.
FROM gcr.io/distroless/static:nonroot
WORKDIR /
ARG TARGETOS
ARG TARGETARCH
COPY manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
`
