// Copyright 2006-2021 VMware, Inc.
// SPDX-License-Identifier: MIT
package main

import (
	log "github.com/sirupsen/logrus"

	"gitlab.eng.vmware.com/landerr/kb-license-plugin/pkg/cli"
)

func main() {
	c := cli.GetPluginsCLI()
	if err := c.Run(); err != nil {
		log.Fatal(err)
	}
}
