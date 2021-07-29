package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/vmware-tanzu-labs/operator-builder/pkg/cli"
)

func main() {
	command, err := cli.NewKubebuilderCLI()
	if err != nil {
		log.Fatal(err)
	}

	if err := command.Run(); err != nil {
		log.Fatal(err)
	}
}
