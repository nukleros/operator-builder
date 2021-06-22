#!/bin/bash

mkdir .test/tenancy
mkdir .test/ingress

cat > .test/cnp-workload-collection.yaml <<EOF
name: cloud-native-platform
kind: WorkloadCollection
spec:
  domain: acme.com
  apiGroup: platforms
  apiVersion: v1alpha1
  apiKind: CloudNativePlatform
  clusterScoped: true
  companionCliRootcmd:
    name: cnpctl
    description: Manage platform stuff like a boss
  componentFiles:
  - tenancy-common-component.yaml
  - ns-operator-component.yaml
  - contour-component.yaml
EOF

cat > .test/tenancy-common-component.yaml <<EOF
name: tenancy-common-component
kind: ComponentWorkload
spec:
  apiGroup: tenancy
  apiVersion: v1alpha1
  apiKind: TenancyCommon
  clusterScoped: true
  companionCliSubcmd:
    name: tenancy-common
    description: Manage common tenancy component
  resources:
  - ns-operator-ns.yaml
EOF

cat > .test/ns-operator-component.yaml <<EOF
name: ns-operator-component
kind: ComponentWorkload
spec:
  apiGroup: tenancy
  apiVersion: v1alpha1
  apiKind: NsOperator
  clusterScoped: true
  companionCliSubcmd:
    name: ns-operator
    description: Manage namespace operator component
  resources:
  - ns-operator-crd.yaml
  - ns-operator-deploy.yaml
  dependencies:
  - tenancy-common-component
EOF

cat > .test/ingress/contour-component.yaml <<EOF
name: contour-component
kind: ComponentWorkload
spec:
  apiGroup: ingress
  apiVersion: v1alpha1
  apiKind: Contour
  clusterScoped: true
  companionCliSubcmd:
    name: contour
    description: Manage contour component
  resources:
  - ingress-ns.yaml
  - contour-config.yaml
  - contour-deploy.yaml
  - contour-svc.yaml
  - envoy-ds.yaml
  dependencies:
  - ns-operator-component
EOF

cat > .test/tenancy/ns-operator-ns.yaml <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: tenancy-system  # +workload:namespace:default=tenancy-system:type=string
EOF

cat > .test/tenancy/ns-operator-crd.yaml <<EOF
---
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  annotations:
    controller-gen.kubebuilder.io/version: v0.2.5
  name: tanzunamespaces.tenancy.platform.cnr.vmware.com
spec:
  group: tenancy.platform.cnr.vmware.com
  names:
    kind: TanzuNamespace
    listKind: TanzuNamespaceList
    plural: tanzunamespaces
    singular: tanzunamespace
    shortNames:
      - tns
  scope: Cluster
  versions:
    - name: v1alpha1
      schema:
        openAPIV3Schema:
          description: TanzuNamespace is the Schema for the tanzunamespaces API
          properties:
            apiVersion:
              description: 'APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
              type: string
            kind:
              description: 'Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
              type: string
            metadata:
              type: object
            spec:
              description: TanzuNamespaceSpec defines the desired state of TanzuNamespace
              properties:
                #
                # common
                #
                # NOTE: pick one or the other, but not both.  name is primary. defaults to the name of the CRD object.
                name:
                  type: string
                tanzuNamespaceName:
                  type: string
                #
                # network policies
                #
                networkPolicies:
                  default: []
                  items:
                    description: NetworkPolicy defines an individual network policy which belongs to an array of NetworkPolicies
                    properties:
                      egressNamespaceLabels:
                        additionalProperties:
                          type: string
                        type: object
                        default: {}
                      egressPodLabels:
                        additionalProperties:
                          type: string
                        type: object
                        default: {}
                      egressTCPPorts:
                        items:
                          type: integer
                        type: array
                        default: []
                      egressUDPPorts:
                        items:
                          type: integer
                        type: array
                        default: []
                      ingressNamespaceLabels:
                        additionalProperties:
                          type: string
                        type: object
                        default: {}
                      ingressPodLabels:
                        additionalProperties:
                          type: string
                        type: object
                        default: {}
                      ingressTCPPorts:
                        items:
                          type: integer
                        type: array
                        default: []
                      ingressUDPPorts:
                        items:
                          type: integer
                        type: array
                        default: []
                      targetPodLabels:
                        additionalProperties:
                          type: string
                        type: object
                        default: {}
                    type: object
                  type: array
                #
                # limit range
                #
                # NOTE: backwards compatibility
                tanzuLimitRangeDefaultCpuLimit:
                  default: 125m
                  type: string
                tanzuLimitRangeDefaultCpuRequest:
                  default: 125m
                  type: string
                tanzuLimitRangeDefaultMemoryLimit:
                  default: 64Mi
                  type: string
                tanzuLimitRangeDefaultMemoryRequest:
                  default: 64Mi
                  type: string
                tanzuLimitRangeMaxCpuLimit:
                  default: 1000m
                  type: string
                tanzuLimitRangeMaxMemoryLimit:
                  default: 2Gi
                  type: string
                # NOTE: new object for limitRange*
                limitRange:
                  default: {}
                  type: object
                  properties:
                    defaultCPULimit:
                      type: string
                    defaultCPURequest:
                      type: string
                    defaultMemoryLimit:
                      type: string
                    defaultMemoryRequest:
                      type: string
                    maxCPULimit:
                      type: string
                    maxMemoryLimit:
                      type: string
                #
                # resource quota
                #
                # NOTE: backwards compatibility
                tanzuResourceQuotaCpuLimits:
                  default: 2000m
                  type: string
                tanzuResourceQuotaCpuRequests:
                  default: 2000m
                  type: string
                tanzuResourceQuotaMemoryLimits:
                  default: 4Gi
                  type: string
                tanzuResourceQuotaMemoryRequests:
                  default: 4Gi
                  type: string
                # NOTE: new object for resourceQuota*
                resourceQuota:
                  default: {}
                  type: object
                  properties:
                    limitsCPU:
                      type: string
                    limitsMemory:
                      type: string
                    requestsCPU:
                      type: string
                    requestsMemory:
                      type: string
                #
                # rbac
                #
                rbac:
                  default: []
                  items:
                    properties:
                      type:
                        type: string
                        enum:
                          - namespace-admin
                          - developer
                          - read-only
                          # - "custom"  # TODO: support this later
                      create:
                        type: boolean
                        default: false
                      user:
                        type: string
                        default: ""
                      role:
                        type: string
                        default: ""
                      roleBinding:
                        type: string
                        default: ""
                    type: object
                  type: array
              required: []
              type: object
          type: object
      served: true
      storage: true
      subresources:
        status: {}
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
EOF

cat > .test/tenancy/ns-operator-deploy.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app.kubernetes.io/name: namespace-operator
  name: namespace-operator
  namespace: tenancy-system  # +workload:namespace:default=tenancy-system:type=string
spec:
  replicas: 2  # +workload:nsOperatorReplicas:default=2:type=int
  selector:
    matchLabels:
      app.kubernetes.io/name: namespace-operator
  template:
    metadata:
      labels:
        app.kubernetes.io/name: namespace-operator
      name: namespace-operator
    spec:
      containers:
        - name: namespace-operator
          image: nginx:1.17  # +workload:nsOperatorImage:type=string
EOF

cat > .test/ingress/ingress-ns.yaml <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: ingress-system  # +workload:namespace:default=ingress-system:type=string
EOF

cat > .test/ingress/contour-config.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: contour-configmap
  namespace: ingress-system  # +workload:namespace:default=ingress-system:type=string
data:
  config.yaml: |
    someoption: myoption
    anotheroption: another
    justtesting: multistringyaml
---
apiVersion: v1
kind: Secret
metadata:
  name: contour-secret
  namespace: ingress-system  # +workload:namespace:default=ingress-system:type=string
stringData:
  some: secretstuff
EOF

cat > .test/ingress/contour-deploy.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: contour-deploy
  namespace: ingress-system  # +workload:namespace:default=ingress-system:type=string
spec:
  replicas: 2  # +workload:ContourReplicas:default=2:type=int
  selector:
    matchLabels:
      app: contour
  template:
    metadata:
      labels:
        app: contour
    spec:
      containers:
      - name: contour
        image: nginx:1.17  # +workload:ContourImage:type=string
        ports:
        - containerPort: 8080
EOF

cat > .test/ingress/contour-svc.yaml <<EOF
kind: Service
apiVersion: v1
metadata:
  name: contour-svc
  namespace: ingress-system  # +workload:namespace:default=ingress-system:type=string
spec:
  selector:
    app: contour
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
EOF

cat > .test/ingress/envoy-ds.yaml <<EOF
apiVersion: apps/v1
kind: DaemonSet
metadata:
  labels:
    app.kubernetes.io/name: envoy
  name: envoy-ds
  namespace: ingress-system  # +workload:namespace:default=ingress-system:type=string
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: envoy
  template:
    metadata:
      labels:
        app.kubernetes.io/name: envoy
    spec:
      containers:
      - name: envoy
        image: nginx:1.17  # +workload:EnvoyImage:type=string
EOF

# TODO: domain flag exists as part of workload-collection but is not
# properly being pulled in during init
# see https://github.com/vmware-tanzu-labs/operator-builder/issues/11
go mod init acme.com/operator-builder-test

operator-builder init \
    --workload-config .test/cnp-workload-collection.yaml

operator-builder create api \
    --workload-config .test/cnp-workload-collection.yaml \
    --controller \
    --resource

