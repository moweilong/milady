package commands

import (
	"github.com/spf13/cobra"

	"github.com/moweilong/milady/cmd/milady/commands/template"
)

// TemplateCommand generate code based on custom templates
func TemplateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "template",
		Short:         "Generate code based on custom templates",
		Long:          `Generate code based on custom templates.`,
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	cmd.AddCommand(
		template.FieldCommand(),
		template.SQLCommand(),
		template.ProtobufCommand(),
	)

	return cmd
}
