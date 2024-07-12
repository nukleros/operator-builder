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

cat > /tmp/updated-project.txt <<EOF
    !! UPDATED !!
    MIT License

    Copyright (c) Acme Inc. All rights reserved.

    Permission is hereby granted, free of charge, to any person obtaining a copy
    of this software and associated documentation files (the "Software"), to deal
    in the Software without restriction, including without limitation the rights
    to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
    copies of the Software, and to permit persons to whom the Software is
    furnished to do so, subject to the following conditions:

    The above copyright notice and this permission notice shall be included in all
    copies or substantial portions of the Software.

    THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
    IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
    FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
    AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
    LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
    OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
    SOFTWARE

EOF

cat > /tmp/updated-source-header.txt <<EOF
// !! UPDATED !!
// Copyright 2006-2021 Acme Inc.
// SPDX-License-Identifier: Apache-2.0
EOF

operator-builder update license \
    --project-license /tmp/updated-project.txt \
    --source-header-license /tmp/updated-source-header.txt

