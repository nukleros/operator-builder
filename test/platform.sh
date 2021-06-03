#!/bin/bash

# collection
cat > .test/cnp-workload-collection.yaml <<EOF
name: cloud-native-platform
kind: WorkloadCollection
spec:
  group: platforms
  version: v1alpha1
  kind: CloudNativePlatform
  clusterScoped: true
  companionCliRootcmd:
    name: cnpctl
    description: Manage platform stuff like a boss
  components:
  - ns-operator-component
  - contour-component
EOF

# tenancy
cat > .test/ns-operator-workload.yaml <<EOF
name: ns-operator-component
kind: ComponentWorkload
spec:
  group: tenancy
  version: v1alpha1
  kind: NsOperator
  clusterScoped: true
  companionCliSubcmd:
    name: ns-operator
    description: Manage namespace operator component
  resources:
  - ns-operator-deploy.yaml
EOF

cat > .test/ns-operator-deploy.yaml <<EOF
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
      serviceAccountName: namespace-operator
      containers:
        - name: namespace-operator
          image: nginx:1.17  # +workload:nsOperatorImage:type=string
          securityContext:
            runAsNonRoot: true
            runAsUser: 65532
            runAsGroup: 65532
EOF

## ingress
#cat > .test/contour-workload.yaml <<EOF
#name: contour-component
#spec:
#  group: ingress
#  version: v1alpha1
#  kind: Contour
#  clusterScoped: true
#  companionCliSubcmd:
#    name: contour
#    description: Manage contour component
#  resources:
#  - contour-deploy.yaml
#  - contour-svc.yaml
#  - envoy-ds.yaml
#EOF
#
#cat > .test/contour-deploy.yaml <<EOF
#apiVersion: apps/v1
#kind: Deployment
#metadata:
#  name: contour-deploy
#  namespace: ingress-system  # +workload:namespace:default=ingress-system:type=string
#spec:
#  replicas: 2  # +workload:ContourReplicas:default=2:type=int
#  selector:
#    matchLabels:
#      app: contour
#  template:
#    metadata:
#      labels:
#        app: contour
#    spec:
#      containers:
#      - name: contour
#        image: nginx:1.17  # +workload:ContourImage:type=string
#        ports:
#        - containerPort: 8080
#EOF
#
#cat > .test/contour-svc.yaml <<EOF
#kind: Service
#apiVersion: v1
#metadata:
#  name: contour-svc
#  namespace: ingress-system  # +workload:namespace:default=ingress-system:type=string
#spec:
#  selector:
#    app: contour
#  ports:
#  - protocol: TCP
#    port: 80
#    targetPort: 8080
#EOF
#
#cat > .test/envoy-ds.yaml <<EOF
#apiVersion: apps/v1
#kind: DaemonSet
#metadata:
#  labels:
#    app.kubernetes.io/name: envoy
#  name: envoy-ds
#  namespace: ingress-system  # +workload:namespace:default=ingress-system:type=string
#spec:
#  selector:
#    matchLabels:
#      app.kubernetes.io/name: envoy
#  template:
#    metadata:
#      labels:
#        app.kubernetes.io/name: envoy
#    spec:
#      containers:
#      - name: envoy
#        image: nginx:1.17  # +workload:EnvoyImage:type=string
#EOF

operator-builder init \
    --domain platform.acme.com \
    --workload-config .test/cnp-workload-collection.yaml

operator-builder create api \
    --workload-config .test/cnp-workload-collection.yaml \
    --group services \
    --version v1alpha1 \
    --kind CloudNativePlatform \
    --controller \
    --resource

