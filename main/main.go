package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("I need a subcommand")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "install":
		if len(os.Args) != 3 {
			fmt.Println("Invalid number of parameters for install")
			os.Exit(1)
		}
		fmt.Println("Not implemented yet")
	case "uninstall":
		if len(os.Args) != 3 {
			fmt.Println("Invalid number of parameters for uninstall")
			os.Exit(1)
		}
		fmt.Println("Not implemented yet")
	case "search":
		if len(os.Args) != 3 {
			fmt.Println("Invalid number of parameters for search")
			os.Exit(1)
		}
		fmt.Println("Not implemented yet")
	case "info":
		if len(os.Args) != 3 {
			fmt.Println("Invalid number of parameters for info")
			os.Exit(1)
		}
		fmt.Println("Not implemented yet")

	case "sync":
		if len(os.Args) != 2 {
			fmt.Println("Invalid number of parameters for sync")
			os.Exit(1)
		}
		fmt.Println("Not implemented yet")
	case "update":
		if len(os.Args) < 2 {
			fmt.Println("Invalid number of parameters for update")
			os.Exit(1)
		}
		fmt.Println("Not implemented yet")
	case "add-repository":
		if len(os.Args) != 3 {
			fmt.Println("Invalid number of parameters for add-repository")
			os.Exit(1)
		}
		fmt.Println("Not implemented yet")
	case "remove-repository":
		if len(os.Args) != 3 {
			fmt.Println("Invalid number of parameters for remove-repository")
			os.Exit(1)
		}
		fmt.Println("Not implemented yet")
	case "list-repositories":
		if len(os.Args) != 3 {
			fmt.Println("Invalid number of parameters for list-repositories")
			os.Exit(1)
		}
		fmt.Println("Not implemented yet")
	default:
		fmt.Println("invalid command")
		os.Exit(1)
	}
}
