package commands

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	version     = "v0.0.0"
	versionFile = GetUserHomeDir() + "/.milady/.github/version"
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

// getVersion 获取 milady 版本
// 版本文件存储在 milady 主目录下的 .milady/.github/version 文件中
func getVersion() string {
	data, _ := os.ReadFile(versionFile)
	v := string(data)
	if v != "" {
		return v
	}
	return "unknown, execute command \"milady init\" to get version"
}

// GetUserHomeDir 获取用户主目录
func GetUserHomeDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("can't get user home directory")
		return ""
	}

	return dir
}
