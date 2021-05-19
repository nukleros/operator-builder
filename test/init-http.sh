#!/bin/bash

kbl init \
    --domain apps.acme.com \
    --project-license https://raw.githubusercontent.com/lander2k2/license/master/project.txt \
    --source-header-license https://raw.githubusercontent.com/lander2k2/license/master/source-header.txt

kbl create api \
    --group workloads \
    --version v1alpha1 \
    --kind WebApp \
    --controller \
    --resource

