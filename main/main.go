package main

import (
	"fmt"
	"os"

	"github.com/ava-labs/apm/service"
)

func main() {

	api := service.New()
	var err error

	if len(os.Args) < 2 {
		fmt.Println("missing subcommand")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "install":
		err = api.Install(os.Args[2])
	case "uninstall":
		err = api.Uninstall(os.Args[2])
	case "search":
		err = api.Search(os.Args[2])
	case "info":
		err = api.Info(os.Args[2])
	case "sync":
		err = api.Sync(os.Args[2])
	case "update":
		err = api.Update(os.Args[2])
	case "add-repository":
		err = api.AddRepository(os.Args[2])
	case "remove-repository":
		err = api.RemoveRepository(os.Args[2])
	case "list-repositories":
		err = api.ListRepositories()
	default:
		fmt.Println("invalid command")
		os.Exit(2)
	}

	if err != nil {
		os.Exit(3)
	}
}
