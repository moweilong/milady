package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/moweilong/milady/pkg/core"
	"github.com/moweilong/milady/pkg/log"
	"github.com/moweilong/milady/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	genericapiserver "k8s.io/apiserver/pkg/server"
	logsv1 "k8s.io/component-base/logs/api/v1"

	"github.com/moweilong/milady/cmd/milady-apiserver/app/options"
)

const (
	// defaultHomeDir defines the default directory to store the configuration for the milady-apiserver service.
	defaultHomeDir = ".milady"

	// defaultConfigName specifies the default configuration file name for the milady-apiserver service.
	defaultConfigName = "milady-apiserver.yaml"
)

// Path to the configuration file
var configFile string

// NewWebServerCommand creates a *cobra.Command object used to start the application.
func NewWebServerCommand() *cobra.Command {
	// Create default application command-line options
	opts := options.NewServerOptions()

	cmd := &cobra.Command{
		// Specify the name of the command, which will appear in the help information
		Use: "milady-apiserver",
		// A short description of the command
		Short: "milady",
		// A detailed description of the command
		Long: `milady long`,
		// Do not print help information when the command encounters an error.
		// Setting this to true ensures that errors are immediately visible.
		SilenceUsage: true,
		// Specify the Run function to execute when cmd.Execute() is called
		RunE: func(cmd *cobra.Command, args []string) error {
			// If the --version flag is passed, print version information and exit
			version.PrintAndExitIfRequested()

			// Unmarshal the configuration from viper into opts
			if err := viper.Unmarshal(opts); err != nil {
				return fmt.Errorf("failed to unmarshal configuration: %w", err)
			}

			if err := logsv1.ValidateAndApply(opts.LogOptions.Native(), nil); err != nil {
				return err
			}

			// Validate command-line options
			if err := opts.Validate(); err != nil {
				return fmt.Errorf("invalid options: %w", err)
			}

			ctx := genericapiserver.SetupSignalContext()

			return run(ctx, opts)
		},
		// Set argument validation for the command. No command-line arguments are required.
		// For example: ./mwl-apiserver param1 param2
		Args: cobra.NoArgs,
	}

	// Initialize configuration function, called when each command runs
	cobra.OnInitialize(core.OnInitialize(&configFile, "MILADY_APISERVER", searchDirs(), defaultConfigName))

	// cobra supports persistent flags, which apply to the assigned command and all its subcommands.
	// It is recommended to use configuration files for application configuration to make it easier to manage configuration items.
	cmd.PersistentFlags().StringVarP(&configFile, "config", "c", filePath(), "Path to the milady-apiserver configuration file.")

	// Add server options as flags
	opts.AddFlags(cmd.PersistentFlags())

	// Add the --version flag
	version.AddFlags(cmd.PersistentFlags())

	return cmd
}

// run contains the main logic for initializing and running the server.
func run(ctx context.Context, opts *options.ServerOptions) error {
	// 初始化日志
	log.Init(logOptions())
	defer log.Sync() // 确保日志在退出时被刷新到磁盘

	// Retrieve application configuration
	// Separating command-line options and application configuration allows more flexible handling of these two types of configurations.
	cfg, err := opts.Config()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create and start the server
	server, err := cfg.NewServer(ctx)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	// Run the server
	return server.Run(ctx)
}

// searchDirs returns the default directories to search for the configuration file.
func searchDirs() []string {
	// Get the user's home directory.
	homeDir, err := os.UserHomeDir()
	// If unable to get the user's home directory, print an error message and exit the program.
	cobra.CheckErr(err)
	return []string{filepath.Join(homeDir, defaultHomeDir), "."}
}

// filePath retrieves the full path to the default configuration file.
func filePath() string {
	home, err := os.UserHomeDir()
	// If the user's home directory cannot be retrieved, log an error and return an empty path.
	cobra.CheckErr(err)
	return filepath.Join(home, defaultHomeDir, defaultConfigName)
}

// logOptions 从 viper 中读取日志配置，构建 *log.Options 并返回.
// 注意：viper.Get<Type>() 中 key 的名字需要使用 . 分割，以跟 YAML 中保持相同的缩进.
func logOptions() *log.Options {
	opts := log.NewOptions()
	if viper.IsSet("log.disable-caller") {
		opts.DisableCaller = viper.GetBool("log.disable-caller")
	}
	if viper.IsSet("log.disable-stacktrace") {
		opts.DisableStacktrace = viper.GetBool("log.disable-stacktrace")
	}
	if viper.IsSet("log.level") {
		opts.Level = viper.GetString("log.level")
	}
	if viper.IsSet("log.format") {
		opts.Format = viper.GetString("log.format")
	}
	if viper.IsSet("log.output-paths") {
		opts.OutputPaths = viper.GetStringSlice("log.output-paths")
	}
	return opts
}
