#!/bin/bash

operator-builder init \
    --domain apps.acme.com \
    --project-license https://raw.githubusercontent.com/lander2k2/license/master/project.txt \
    --source-header-license https://raw.githubusercontent.com/lander2k2/license/master/source-header.txt

operator-builder create api \
    --group workloads \
    --version v1alpha1 \
    --kind WebApp \
    --controller \
    --resource

operator-builder update license \
    --project-license https://raw.githubusercontent.com/lander2k2/license/master/project.txt \
    --source-header-license https://raw.githubusercontent.com/lander2k2/license/master/source-header.txt

