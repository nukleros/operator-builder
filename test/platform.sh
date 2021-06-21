#!/bin/bash

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
  componentNames:
  - ns-operator-component
  - contour-component
  #componentFiles:
  #- ns_operator-component.yaml
  #- contour-component.yaml
---
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
  - tenancy/ns-operator-deploy.yaml
---
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
  - ingress/contour-deploy.yaml
  - ingress/contour-svc.yaml
  - ingress/envoy-ds.yaml
  dependencies:
  - ns-operator-component
EOF

mkdir .test/tenancy
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

## ingress
#cat > .test/contour-workload.yaml <<EOF
#name: contour-component
#spec:
#  apiGroup: ingress
#  apiVersion: v1alpha1
#  apiKind: Contour
#  clusterScoped: true
#  companionCliSubcmd:
#    name: contour
#    description: Manage contour component
#  resources:
#  - contour-deploy.yaml
#  - contour-svc.yaml
#  - envoy-ds.yaml
#EOF

mkdir .test/ingress
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

operator-builder init \
    --workload-config .test/cnp-workload-collection.yaml

operator-builder create api \
    --workload-config .test/cnp-workload-collection.yaml \
    --controller \
    --resource

