# License Management

Manage the creation and update of licensing for your Kubebuilder project.

## Try it Out

Create two license files for testing:

    cat > /tmp/project.txt <<EOF
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

    cat > /tmp/source-header.txt <<EOF
    // Copyright 2006-2021 Acme Inc.
    // SPDX-License-Identifier: MIT
    EOF

Now initialize a new Kubebuilder project and reference your license files.

    operator-builder init \
        --domain apps.acme.com \
        --project-license /tmp/project.txt \
        --source-header-license /tmp/source-header.txt

You will now have a `LICENSE` file in your project which has the contents of
`/tmp/project.txt`.  The `hack/boilerplate.go.txt` file will have the contents
of `/tmp/source-header.txt` which you will also find at the top of `main.go`.

## Update Existing Project

If you have an existing project that you would like to update licensing for:

    operator-builder update license \
        --project-license /tmp/project.txt \
        --source-header-license /tmp/source-header.txt


