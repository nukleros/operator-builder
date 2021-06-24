package cli

import (
	"github.com/spf13/cobra"
)

func NewUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update an existing project",
		Long:  `Update an existing project.`,
	}

	cmd.AddCommand(
		NewUpdateLicenseCmd(),
	)

	return cmd
}
