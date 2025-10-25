package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/moweilong/milady/pkg/gobash"
	"github.com/spf13/cobra"
)

const (
	installedSymbol = "✔ "
	lackSymbol      = "❌ "
	warnSymbol      = "⚠ "
)

// pluginNames milady 依赖插件
var pluginNames = []string{
	"go",
	"protoc",
	"protoc-gen-go",
	"protoc-gen-go-grpc",
	"protoc-gen-validate",
	"protoc-gen-gotag",
	"protoc-gen-go-gin",
	"protoc-gen-go-rpc-tmpl",
	"protoc-gen-json-field",
	"protoc-gen-openapiv2",
	"protoc-gen-doc",
	"swag",
	"golangci-lint",
	"go-callvis",
}

// installPluginCommands 插件安装命令
var installPluginCommands = map[string]string{
	"go":                     "go: please install manually yourself, download url is https://go.dev/dl/ or https://golang.google.cn/dl/",
	"protoc":                 "protoc: please install manually yourself, download url is https://github.com/protocolbuffers/protobuf/releases/tag/v31.1", // TODO 获取最新版本号
	"protoc-gen-go":          "google.golang.org/protobuf/cmd/protoc-gen-go@latest",
	"protoc-gen-go-grpc":     "google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest",
	"protoc-gen-validate":    "github.com/envoyproxy/protoc-gen-validate@latest",
	"protoc-gen-gotag":       "github.com/srikrsna/protoc-gen-gotag@latest",
	"protoc-gen-go-gin":      "github.com/moweilong/milady/cmd/protoc-gen-go-gin@latest",
	"protoc-gen-go-rpc-tmpl": "github.com/moweilong/milady/cmd/protoc-gen-go-rpc-tmpl@latest",
	"protoc-gen-json-field":  "github.com/moweilong/milady/cmd/protoc-gen-json-field@latest",
	"protoc-gen-openapiv2":   "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest",
	"protoc-gen-doc":         "github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc@latest",
	"swag":                   "github.com/swaggo/swag/cmd/swag@v1.8.12", // TODO 获取最新版本号
	"golangci-lint":          "github.com/golangci/golangci-lint/cmd/golangci-lint@latest",
	"go-callvis":             "github.com/ofabry/go-callvis@latest",
}

// PluginsCommand 插件管理, 包括插件安装、插件升级等
func PluginsCommand() *cobra.Command {
	var installFlag bool
	var skipPluginName string

	cmd := &cobra.Command{
		Use:   "plugins",
		Short: "Manage milady dependency plugins",
		Long:  "Manage milady dependency plugins.",
		Example: color.HiBlackString(`  # Show all dependency plugins.
  milady plugins

  # Install all dependency plugins.
  milady plugins --install

  # Skip installing dependency plugins, multiple plugin names separated by commas
  milady plugins --install --skip=go-callvis`),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			installedNames, lackNames := checkInstallPlugins()
			lackNames = filterLackNames(lackNames, skipPluginName)
			if installFlag {
				installPlugins(lackNames)
			} else {
				showDependencyPlugins(installedNames, lackNames)
			}

			return nil
		},
	}
	cmd.Flags().BoolVarP(&installFlag, "install", "i", false, "install dependency plugins")
	cmd.Flags().StringVarP(&skipPluginName, "skip", "s", "", "skip installing dependency plugins")

	return cmd
}

// checkInstallPlugins 检查插件是否已安装
// 返回已经安装的插件名称列表和未安装的插件名称列表
func checkInstallPlugins() (installedNames []string, lackNames []string) {
	for _, name := range pluginNames {
		_, err := gobash.Exec("which", name)
		if err != nil {
			lackNames = append(lackNames, name)
			continue
		}
		installedNames = append(installedNames, name)
	}

	data, _ := os.ReadFile(versionFile)
	v := string(data)
	if v != "" {
		version = v
	}

	return installedNames, lackNames
}

// filterLackNames 忽略指定的插件, 返回未忽略的插件名称列表
func filterLackNames(lackNames []string, skipPluginName string) []string {
	if skipPluginName == "" {
		return lackNames
	}
	skipPluginNames := strings.Split(skipPluginName, ",")

	names := []string{}
	for _, name := range lackNames {
		isMatch := false
		for _, pluginName := range skipPluginNames {
			if name == pluginName {
				isMatch = true
				continue
			}
		}
		if !isMatch {
			names = append(names, name)
		}
	}
	return names
}

// installPlugins 安装插件
func installPlugins(lackNames []string) {
	if len(lackNames) == 0 {
		fmt.Printf("\n    all dependency plugins installed.\n\n")
		return
	}
	fmt.Printf("\ninstalling %d dependency plugins, please wait a moment.\n\n", len(lackNames))

	var wg = &sync.WaitGroup{}
	var manuallyNames []string
	for _, name := range lackNames {
		// go 和 protoc 插件需要手动安装
		if name == "go" || name == "protoc" {
			manuallyNames = append(manuallyNames, name)
			continue
		}

		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3) //nolint
			defer cancel()
			pkgAddr, ok := installPluginCommands[name]
			if !ok {
				return
			}
			pkgAddr = adaptInternalCommand(name, pkgAddr)
			result := gobash.Run(ctx, "go", "install", pkgAddr)
			for v := range result.StdOut {
				_ = v
			}
			if result.Err != nil {
				fmt.Printf("%s %s, %v\n", lackSymbol, name, result.Err)
			} else {
				fmt.Printf("%s %s\n", installedSymbol, name)
			}
		}(name)
	}

	wg.Wait()

	for _, name := range manuallyNames {
		fmt.Println(warnSymbol + " " + installPluginCommands[name])
	}
	fmt.Println()
}

// adaptInternalCommand 适配内部插件, 如果版本不是 v0.0.0, 则替换为指定版本
func adaptInternalCommand(name string, pkgAddr string) string {
	if name == "protoc-gen-go-gin" || name == "protoc-gen-go-rpc-tmpl" ||
		name == "protoc-gen-json-field" {
		if version != "v0.0.0" {
			return strings.ReplaceAll(pkgAddr, "@latest", "@"+version)
		}
	}

	return pkgAddr
}

// showDependencyPlugins 显示插件安装情况, 已安装的插件显示为已安装, 未安装的插件显示为未安装
func showDependencyPlugins(installedNames []string, lackNames []string) {
	var content string

	if len(installedNames) > 0 {
		content = "installed dependency plugins:\n"
		for _, name := range installedNames {
			content += "    " + installedSymbol + " " + name + "\n"
		}
	}

	if len(lackNames) > 0 {
		content += "\nuninstalled dependency plugins:\n"
		for _, name := range lackNames {
			content += "    " + lackSymbol + " " + name + "\n"
		}
		content += "\ninstalling dependency plugins using the command: milady plugins --install\n"
	} else {
		content += "\nall dependency plugins installed.\n"
	}

	fmt.Println(content)
}
