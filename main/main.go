package main

import "github.com/ava-labs/apm/cmd"

func main() {
	//var err error
	//
	//fmt.Println("-----------------------------------------------")
	//fmt.Println("Bootstrap:")
	//cmd, err := cmd.New(
	//	cmd.Config{
	//		WorkingDir: workingDir,
	//	},
	//)
	//fmt.Println("-----------------------------------------------")
	//fmt.Println("Update:")
	//cmd.Update()
	//fmt.Println("-----------------------------------------------")
	//fmt.Println("ListRepositories:")
	//cmd.ListRepositories()
	//fmt.Println("-----------------------------------------------")
	//fmt.Println("Install:")
	//cmd.Install("spacesvm")
	//fmt.Println("-----------------------------------------------")
	//fmt.Println("Install Again:")
	//cmd.Install("spacesvm")
	//fmt.Println("-----------------------------------------------")
	//
	//if err != nil {
	//	fmt.Printf("unexpected error: %s\n", err)
	//}
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
