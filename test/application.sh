#!/bin/bash

cat > .test/workload.yaml <<EOF
name: webstore
kind: StandaloneWorkload
spec:
  domain: acme.com
  apiGroup: apps
  apiVersion: v1alpha1
  apiKind: WebStore
  clusterScoped: false
  companionCliRootcmd:
    name: webstorectl
    description: Manage webstore stuff like a boss
  resources:
  - app.yaml
EOF

cat > .test/app.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webstore-deploy
spec:
  replicas: 2  # +workload:webStoreReplicas:default=2:type=int
  selector:
    matchLabels:
      app: webstore
  template:
    metadata:
      labels:
        app: webstore
    spec:
      containers:
      - name: webstore-container
        #+docs: Defines the web store image
        image: nginx:1.17  # +workload:webStoreImage:type=string
        ports:
        - containerPort: 8080
---
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: webstore-ing
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: app.acme.com
    http:
      paths:
      - path: /
        backend:
          serviceName: webstorep-svc
          servicePort: 80
---
kind: Service
apiVersion: v1
metadata:
  name: webstore-svc
spec:
  selector:
    app: webstore
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
EOF

operator-builder init \
    --workload-config .test/workload.yaml \
    --repo github.com/acme/acme-webstore-mgr \
    --skip-go-version-check

operator-builder create api \
    --workload-config .test/workload.yaml \
    --controller \
    --resource

