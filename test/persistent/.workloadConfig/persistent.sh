#!/bin/bash

operator-builder init \
    --workload-config .workloadConfig/workload.yaml \
    --repo github.com/acme/acme-webstore-mgr \
    --skip-go-version-check

operator-builder create api \
    --workload-config .workloadConfig/workload.yaml \
    --controller \
    --resource

