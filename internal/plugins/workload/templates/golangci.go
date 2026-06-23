package templates

const Linter = `version: "2"
run:
  allow-parallel-runners: true
linters:
  default: none
  enable:
    - copyloopvar
    - depguard
    - dupl
    - errcheck
    - ginkgolinter
    - goconst
    - gocyclo
    - govet
    - ineffassign
    - lll
    - modernize
    - misspell
    - nakedret
    - prealloc
    - revive
    - staticcheck
    - unconvert
    - unparam
    - unused
    # - logcheck
  settings:
    # NOTE: we skip logcheck linting here as it includes an external dependency which is not
    #       vendored in the project. 
    # custom:
    #   logcheck:
    #     type: "module"
    #     description: Checks Go logging calls for Kubernetes logging conventions.
    depguard:
      rules:
        forbid-sort-pkg:
          deny:
            - pkg: sort
              desc: Should be replaced with slices package
    revive:
      rules:
        - name: comment-spacings
        - name: import-shadowing
    modernize:
      disable:
        - omitzero
  exclusions:
    generated: lax
    rules:
      - linters:
          - lll
          - goconst
          - modernize
          - revive
        path: api/*
      - linters:
          - lll
          - goconst
          - modernize
          - revive
        path: apis/*
      # NOTE: this ignores a deprecates scheme builder.  when we upgrade the kubebuilder package to something
      #       later than v4.14.0, we will need to fix this as well.
      - linters:
          - staticcheck
        path: apis/.*groupversion_info\.go$
      # NOTE: this ignores a deprecates event recorder.  when we upgrade the kubebuilder package to something
      #       later than v4.14.0, we will need to fix this as well.  this also requires an operator-builder-tools
      #       package fix.
      - linters:
          - staticcheck
        path: controllers/.*_controller\.go$
      - linters:
          - dupl
          - lll
        path: internal/*
    paths:
      - third_party$
      - builtin$
      - examples$
formatters:
  enable:
    - gofmt
    - goimports
  exclusions:
    generated: lax
    paths:
      - third_party$
      - builtin$
      - examples$
`
