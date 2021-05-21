#!/bin/bash

cat > .test/workload.yaml <<EOF
apiVersion: workload.cnr.vmware.com/v1alpha1
kind: Workload
metadata:
  name: WebApp
spec:
  clusterScoped: false
  importPath: gitlab.eng.vmware.com/landerr/acme-webapp-mgr
  companionCliRootcmd:
    name: webappctl
    description: Manage webapp stuff like a boss
  resources:
  - app.yaml
EOF

cat > .test/app.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webapp-deploy
  labels:
    production: false  #+workload:production:default=false:type=bool
spec:
  replicas: 2  # +workload:webAppReplicas:default=2:type=int
  selector:
    matchLabels:
      app: webapp
  template:
    metadata:
      labels:
        app: webapp
    spec:
      containers:
      - name: webapp-container
        image: nginx:1.17  # +workload:webAppImage:type=string
        ports:
        - containerPort: 8080
---
apiVersion: networking.k8s.io/v1beta1
kind: Ingress
metadata:
  name: webapp-ing
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: app.acme.com
    http:
      paths:
      - path: /
        backend:
          serviceName: webappp-svc
          servicePort: 80
---
kind: Service
apiVersion: v1
metadata:
  name: webapp-svc
spec:
  selector:
    app: webapp
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
EOF

operator-builder init \
    --domain apps.acme.com \
    --workload-config .test/workload.yaml

#operator-builder create api \
#    --group workloads \
#    --version v1alpha1 \
#    --kind WebApp \
#    --controller \
#    --resource

