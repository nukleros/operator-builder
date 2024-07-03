// Copyright 2024 Nukleros
// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: MIT

package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nukleros/operator-builder/internal/license"
)

func NewUpdateLicenseCmd() *cobra.Command {
	var projectLicensePath string

	var sourceHeaderPath string

	cmd := &cobra.Command{
		Use:   "license",
		Short: "Update a project license",
		Long:  `Update a project license.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// project license
			if projectLicensePath != "" {
				if err := license.UpdateProjectLicense(projectLicensePath); err != nil {
					return fmt.Errorf("unable to update project license at %s, %w", projectLicensePath, err)
				}
			}

			// source header license
			if sourceHeaderPath != "" {
				// boilerplate
				if err := license.UpdateSourceHeader(sourceHeaderPath); err != nil {
					return fmt.Errorf("unable to update source header file at %s, %w", sourceHeaderPath, err)
				}
				// existing source code files
				if err := license.UpdateExistingSourceHeader(sourceHeaderPath); err != nil {
					return fmt.Errorf("unable to update source header file at %s, %w", sourceHeaderPath, err)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&projectLicensePath, "project-license", "p", "", "path to project license file")
	cmd.Flags().StringVarP(&sourceHeaderPath, "source-header-license", "s", "", "path to file with source code header license text")

	return cmd
}
