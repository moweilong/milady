package commands

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/moweilong/milady/pkg/gobash"
	"github.com/moweilong/milady/pkg/gofile"
	"github.com/moweilong/milady/pkg/utils"
	"github.com/spf13/cobra"
)

// UpgradeCommand 升级 milady 版本, 包括 milady 二进制文件、模板代码、内置插件
func UpgradeCommand() *cobra.Command {
	var targetVersion string

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade milady version",
		Long:  "Upgrade milady version.",
		Example: color.HiBlackString(`  # Upgrade to latest version
  milady upgrade

  # Upgrade to specified version
  milady upgrade --version=v1.5.6`),
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if targetVersion == "" {
				targetVersion = latestVersion
			}
			ver, err := runUpgrade(targetVersion)
			if err != nil {
				return err
			}
			fmt.Printf("upgraded version to %s successfully.\n", ver)
			return nil
		},
	}

	cmd.Flags().StringVarP(&targetVersion, "version", "v", latestVersion, "upgrade milady version")
	return cmd
}

// runUpgrade 升级 milady 相关文件, 包括 milady 二进制文件、模板代码、内置插件
func runUpgrade(targetVersion string) (string, error) {
	runningTip := "Upgrading milady binary "
	finishTip := "Upgrade milady binary done " + installedSymbol
	failedTip := "Upgrade milady binary failed " + lackSymbol
	p := utils.NewWaitPrinter(time.Millisecond * 100)
	p.LoopPrint(runningTip)
	err := runUpgradeCommand(targetVersion)
	if err != nil {
		p.StopPrint(failedTip + "\nError: " + err.Error())
		return "", err
	}
	p.StopPrint(finishTip)

	runningTip = "upgrading template code "
	finishTip = "upgrade template code done " + installedSymbol
	failedTip = "upgrade template code failed " + lackSymbol
	p = utils.NewWaitPrinter(time.Millisecond * 500)
	p.LoopPrint(runningTip)
	ver, err := copyToTempDir(targetVersion)
	if err != nil {
		p.StopPrint(failedTip + "\nError: " + err.Error())
		return "", err
	}
	p.StopPrint(finishTip)

	runningTip = "upgrading the built-in plugins of milady "
	finishTip = "upgrade the built-in plugins of milady done " + installedSymbol
	failedTip = "upgrade the built-in plugins of milady failed " + lackSymbol
	p = utils.NewWaitPrinter(time.Millisecond * 500)
	p.LoopPrint(runningTip)
	err = updateMiladyInternalPlugin(ver)
	if err != nil {
		p.StopPrint(failedTip + "\nError: " + err.Error())
		return "", err
	}
	p.StopPrint(finishTip)
	return ver, nil
}

// runUpgradeCommand upgrade milady binary
func runUpgradeCommand(targetVersion string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*3)
	defer cancel()
	// github.com/moweilong/milady/cmd/milady@latest
	miladyVersion := "github.com/moweilong/milady/cmd/milady@" + targetVersion
	result := gobash.Run(ctx, "go", "install", miladyVersion)
	for v := range result.StdOut {
		// fmt.Println(v)
		_ = v
	}
	if result.Err != nil {
		return result.Err
	}
	return nil
}

// copyToTempDir copy the template files to a temporary directory
func copyToTempDir(targetVersion string) (string, error) {
	result, err := gobash.Exec("go", "env", "GOPATH")
	if err != nil {
		return "", fmt.Errorf("execute command failed, %v", err)
	}
	gopath := strings.ReplaceAll(string(result), "\n", "")
	if gopath == "" {
		return "", fmt.Errorf("$GOPATH is empty, you need set $GOPATH in your $PATH")
	}
	delimiter := ":"
	if gofile.IsWindows() {
		delimiter = ";"
	}
	if ss := strings.Split(gopath, delimiter); len(ss) > 1 {
		gopath = ss[0] // use the first $GOPATH
	}

	miladyDirName := ""
	if targetVersion == latestVersion {
		// find the new version of the milady code directory
		arg := fmt.Sprintf("%s/pkg/mod/github.com/moweilong", gopath)
		result, err = gobash.Exec("ls", adaptPathDelimiter(arg))
		if err != nil {
			return "", fmt.Errorf("execute command failed, %v", err)
		}

		miladyDirName = getLatestVersion(string(result))
		if miladyDirName == "" {
			return "", fmt.Errorf("not found milady directory in '$GOPATH/pkg/mod/github.com/moweilong'")
		}
	} else {
		miladyDirName = "milady@" + targetVersion
	}

	srcDir := adaptPathDelimiter(fmt.Sprintf("%s/pkg/mod/github.com/moweilong/%s", gopath, miladyDirName))
	destDir := adaptPathDelimiter(GetUserHomeDir() + "/")
	targetDir := adaptPathDelimiter(destDir + ".milady")

	err = executeCommand("rm", "-rf", targetDir)
	if err != nil {
		return "", err
	}
	err = executeCommand("cp", "-rf", srcDir, targetDir)
	if err != nil {
		return "", err
	}
	err = executeCommand("chmod", "-R", "744", targetDir)
	if err != nil {
		return "", err
	}
	_ = executeCommand("rm", "-rf", targetDir+"/cmd/milady")
	_ = executeCommand("rm", "-rf", targetDir+"/cmd/protoc-gen-go-gin")
	_ = executeCommand("rm", "-rf", targetDir+"/cmd/protoc-gen-go-rpc-tmpl")
	_ = executeCommand("rm", "-rf", targetDir+"/cmd/protoc-gen-json-field")
	_ = executeCommand("rm", "-rf", targetDir+"/pkg")
	_ = executeCommand("rm", "-rf", targetDir+"/test")
	_ = executeCommand("rm", "-rf", targetDir+"/assets")

	versionNum := strings.Replace(miladyDirName, "milady@", "", 1)
	err = os.WriteFile(versionFile, []byte(versionNum), 0644)
	if err != nil {
		return "", err
	}

	return versionNum, nil
}

// executeCommand execute command
func executeCommand(name string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	result := gobash.Run(ctx, name, args...)
	for v := range result.StdOut {
		_ = v
	}
	if result.Err != nil {
		return fmt.Errorf("execute command failed, %v", result.Err)
	}
	return nil
}

// adaptPathDelimiter adapt path delimiter to windows
func adaptPathDelimiter(filePath string) string {
	if gofile.IsWindows() {
		filePath = strings.ReplaceAll(filePath, "/", "\\")
	}
	return filePath
}

// getLatestVersion get the latest version of the milady code directory
func getLatestVersion(s string) string {
	var dirNames = make(map[int]string)
	var nums []int

	dirs := strings.SplitSeq(s, "\n")
	for dirName := range dirs {
		if strings.Contains(dirName, "milady@") {
			tmp := strings.ReplaceAll(dirName, "milady@", "")
			ss := strings.Split(tmp, ".")
			if len(ss) != 3 {
				continue
			}
			if strings.Contains(ss[2], "v0.0.0") {
				continue
			}
			num := utils.StrToInt(ss[0])*10000 + utils.StrToInt(ss[1])*100 + utils.StrToInt(ss[2])
			nums = append(nums, num)
			dirNames[num] = dirName
		}
	}
	if len(nums) == 0 {
		return ""
	}

	sort.Ints(nums)
	return dirNames[nums[len(nums)-1]]
}

func updateMiladyInternalPlugin(targetVersion string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	genGinVersion := "github.com/moweilong/milady/cmd/protoc-gen-go-gin@" + targetVersion
	result := gobash.Run(ctx, "go", "install", genGinVersion)
	for v := range result.StdOut {
		_ = v
	}
	if result.Err != nil {
		return result.Err
	}

	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	genRPCVersion := "github.com/moweilong/milady/cmd/protoc-gen-go-rpc-tmpl@" + targetVersion
	result = gobash.Run(ctx, "go", "install", genRPCVersion)
	for v := range result.StdOut {
		_ = v
	}
	if result.Err != nil {
		return result.Err
	}

	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	genJSONVersion := "github.com/moweilong/milady/cmd/protoc-gen-json-field@" + targetVersion
	result = gobash.Run(ctx, "go", "install", genJSONVersion)
	for v := range result.StdOut {
		_ = v
	}
	if result.Err != nil {
		return result.Err
	}

	return nil
}
