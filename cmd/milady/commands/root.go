package commands

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	version     = "v0.0.0"
	versionFile = GetMiladyDir() + "/.milady/.github/version"
)

// NewRootCMD command entry
func NewRootCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use: "milady",
		Long: fmt.Sprintf(`
A powerful and easy-to-use Go development framework that enables you to effortlessly 
build stable, reliable, and high-performance backend services with a "low-code" approach.
Repo: %s
Docs: %s`,
			color.HiCyanString("https://github.com/moweilong/milady"),
			color.HiCyanString("https://milady.moweilong.com")),
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       getVersion(),
	}

	cmd.AddCommand(
		InitCommand(),
		UpgradeCommand(),
		PluginsCommand(),
	)

	return cmd
}

// getVersion get milady version
// need execute command "milady init" first, it will create version file in milady home directory
// e.g. ~/.milady/.github/version
func getVersion() string {
	data, _ := os.ReadFile(versionFile)
	v := string(data)
	if v != "" {
		return v
	}
	return "unknown, execute command \"milady init\" to get version"
}

// GetMiladyDir get milady home directory
func GetMiladyDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("can't get home directory'")
		return ""
	}

	return dir
}
