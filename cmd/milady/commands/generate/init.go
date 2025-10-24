package generate

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"github.com/moweilong/milady/pkg/gofile"
	"github.com/moweilong/milady/pkg/replacer"
)

const warnSymbol = "⚠ "

// Replacers replacer name
var Replacers = map[string]replacer.Replacer{}

// MiladyDir milady directory
var MiladyDir = getHomeDir() + gofile.GetPathDelimiter() + ".milady"

// Template information
type Template struct {
	Name     string
	FS       embed.FS
	FilePath string
}

// Init initializing the template
func Init() error {
	// determine if the template directory exists, if not, prompt to initialize first
	if !gofile.IsExists(MiladyDir) {
		if isShowCommand() {
			return nil
		}
		return fmt.Errorf("%s not yet initialized, run the command \"milady init\"", warnSymbol)
	}

	var err error
	// determine if the template name already exists, if so, panic
	if _, ok := Replacers[TplNameMilady]; ok {
		panic(fmt.Sprintf("template name \"%s\" already exists", TplNameMilady))
	}
	// initialize the template
	Replacers[TplNameMilady], err = replacer.New(MiladyDir)
	if err != nil {
		return err
	}

	return nil
}

// InitFS initializing th FS templates
func InitFS(name string, filepath string, fs embed.FS) {
	var err error
	if _, ok := Replacers[name]; ok {
		panic(fmt.Sprintf("template name \"%s\" already exists", name))
	}
	Replacers[name], err = replacer.NewFS(filepath, fs)
	if err != nil {
		panic(err)
	}
}

func isShowCommand() bool {
	l := len(os.Args)

	// milady
	if l == 1 {
		return true
	}

	// milady init or milady -h
	if l == 2 {
		if os.Args[1] == "init" || os.Args[1] == "-h" {
			return true
		}
		return false
	}
	if l > 2 {
		return strings.Contains(strings.Join(os.Args[:3], ""), "init")
	}

	return false
}

func getHomeDir() string {
	dir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("can't get home directory'")
		return ""
	}

	return dir
}
