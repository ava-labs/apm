package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ava-labs/apm/apm"
	"github.com/ava-labs/apm/constant"
)

var (
	homeDir    = os.ExpandEnv("$HOME")
	workingDir = filepath.Join(homeDir, fmt.Sprintf(".%s", constant.AppName))

	rootCmd *cobra.Command

	// arguments
	vmAlias string
)

func init() {
	cobra.EnablePrefixMatching = true

	rootCmd = &cobra.Command{
		Use:   "apm",
		Short: "apm is a plugin manager to help manage virtual machines and subnets",
	}
	rootCmd.AddCommand(
		Install(),
	)
}

func Run() error {
	return rootCmd.Execute()
}

func Install() *cobra.Command {
	command := &cobra.Command{
		Use:   "install",
		Short: "installs a virtual machine by its alias",
	}
	command.PersistentFlags().StringVar(&vmAlias, "vm-alias", "", "vm alias to install")

	install := func(_ *cobra.Command, _ []string) error {
		apm, err := apm.New(apm.Config{})
		if err != nil {
			return err
		}

		return apm.Install(vmAlias)
	}

	command.RunE = install
	return command
}
