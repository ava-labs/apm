package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ava-labs/avalanchego/utils/filesystem"

	"github.com/ava-labs/apm/constant"
	"github.com/ava-labs/apm/service"
)

var (
	homeDir    = os.ExpandEnv("$HOME")
	workingDir = filepath.Join(homeDir, ".", constant.AppName)
)

func main() {

	var err error

	_, err = service.New(
		service.Config{
			WorkingDir: workingDir,
			FsReader:   filesystem.NewReader(),
		},
	)

	//switch os.Args[1] {
	//case "install":
	//	err = api.Install(os.Args[2])
	//case "uninstall":
	//	err = api.Uninstall(os.Args[2])
	//case "search":
	//	err = api.Search(os.Args[2])
	//case "info":
	//	err = api.Info(os.Args[2])
	//case "sync":
	//	err = api.Sync(os.Args[2])
	//case "update":
	//	err = api.Update(os.Args[2])
	//case "add-repository":
	//	_, err = api.AddRepository(os.Args[2])
	//case "remove-repository":
	//	err = api.RemoveRepository(os.Args[2])
	//case "list-repositories":
	//	err = api.ListRepositories()
	//default:
	//	fmt.Println("invalid command")
	//	os.Exit(2)
	//}

	if err != nil {
		fmt.Printf("unexpected error: %s\n", err)
		os.Exit(1)
	}
}
