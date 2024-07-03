// Copyright 2024 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package templates

import (
	"sigs.k8s.io/kubebuilder/v4/pkg/machinery"

	"github.com/nukleros/operator-builder/internal/utils"
)

const (
	defaultDockerfilePath = "Dockerfile"
)

var _ machinery.Template = &Dockerfile{}

// Dockerfile scaffolds a file that defines the containerized build process.
type Dockerfile struct {
	machinery.TemplateMixin

	GoVersion string
}

// SetTemplateDefaults implements file.Template.
func (f *Dockerfile) SetTemplateDefaults() error {
	if f.Path == "" {
		f.Path = defaultDockerfilePath
	}

	f.GoVersion = utils.GeneratedGoVersionPreferred
	f.IfExistsAction = machinery.OverwriteFile
	f.TemplateBody = dockerfileTemplate

	return nil
}

const dockerfileTemplate = `# Build the manager binary
FROM golang:{{ .GoVersion }} as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY apis/ apis/
COPY controllers/ controllers/
COPY internal/ internal/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o manager main.go

# Use distroless as minimal base image to package the manager binary
# Refer to https://github.com/GoogleContainerTools/distroless for more details
FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager .
USER 65532:65532

ENTRYPOINT ["/manager"]
`
