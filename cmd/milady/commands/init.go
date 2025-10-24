package commands

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	latestVersion = "latest"
)

// InitCommand initial milady
func InitCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize milady",
		Long:  "Initialize milady",
		Example: color.HiBlackString(`  # Run init, download code and install plugins.
  milady init`),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			targetVersion := latestVersion
			// download milady template code
			_, err := runUpgrade(targetVersion)
			if err != nil {
				return err
			}

			// installing dependency plugins
			_, lackNames := checkInstallPlugins()
			installPlugins(lackNames)

			return nil
		},
	}

	return cmd
}
