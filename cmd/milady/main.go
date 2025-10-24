package main

import (
	"fmt"
	"os"

	"github.com/moweilong/milady/cmd/milady/commands"
	"github.com/moweilong/milady/cmd/milady/commands/generate"
)

func main() {
	err := generate.Init()
	if err != nil {
		fmt.Printf("\n    %v\n\n", err)
		return
	}

	rootCMD := commands.NewRootCMD()
	if err := rootCMD.Execute(); err != nil {
		rootCMD.PrintErrln("Error:", err)
		os.Exit(1)
	}
}
