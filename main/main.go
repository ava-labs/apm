package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ava-labs/apm/cmd"
	"github.com/ava-labs/apm/constant"
)

var (
	homeDir    = os.ExpandEnv("$HOME")
	workingDir = filepath.Join(homeDir, fmt.Sprintf(".%s", constant.AppName))
)

func main() {
	var err error

	fmt.Println("-----------------------------------------------")
	fmt.Println("Bootstrap:")
	cmd, err := cmd.New(
		cmd.Config{
			WorkingDir: workingDir,
		},
	)
	fmt.Println("-----------------------------------------------")
	fmt.Println("Update:")
	cmd.Update()
	fmt.Println("-----------------------------------------------")
	fmt.Println("ListRepositories:")
	cmd.ListRepositories()
	fmt.Println("-----------------------------------------------")
	fmt.Println("Install:")
	cmd.Install("spacesvm")
	fmt.Println("-----------------------------------------------")
	fmt.Println("Install Again:")
	cmd.Install("spacesvm")
	fmt.Println("-----------------------------------------------")

	if err != nil {
		fmt.Printf("unexpected error: %s\n", err)
	}
}
