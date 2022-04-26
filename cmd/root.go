package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ava-labs/apm/constant"
)

var (
	homeDir = os.ExpandEnv("$HOME")
	apmDir  = filepath.Join(homeDir, fmt.Sprintf(".%s", constant.AppName))
	goPath  = os.ExpandEnv("$GOPATH")

	authToken *http.BasicAuth
)

const (
	ConfigFileKey      = "config-file"
	ApmPathKey         = "apm-path"
	PluginPathKey      = "plugin-path"
	CredentialsFileKey = "credentials-file"
	AdminApiEndpoint   = "admin-api-endpoint"
)

func New() (*cobra.Command, error) {
	cobra.EnablePrefixMatching = true
	rootCmd := &cobra.Command{
		Use:   "apm",
		Short: "apm is a plugin manager to help manage virtual machines and subnets",
	}

	rootCmd.PersistentFlags().String(ConfigFileKey, "", "path to configuration file for the apm")
	rootCmd.PersistentFlags().String(ApmPathKey, apmDir, "path to the directory apm creates its artifacts")
	rootCmd.PersistentFlags().String(PluginPathKey, filepath.Join(goPath, "src", "github.com", "ava-labs", "avalanchego", "build", "plugins"), "path to avalanche plugin directory")
	rootCmd.PersistentFlags().String(CredentialsFileKey, "", "path to credentials file")
	rootCmd.PersistentFlags().String(AdminApiEndpoint, "127.0.0.1:9650/ext/admin", "endpoint for the avalanche admin api")

	errs := wrappers.Errs{}

	errs.Add(
		viper.BindPFlag(ConfigFileKey, rootCmd.PersistentFlags().Lookup(ConfigFileKey)),
		viper.BindPFlag(ApmPathKey, rootCmd.PersistentFlags().Lookup(ApmPathKey)),
		viper.BindPFlag(PluginPathKey, rootCmd.PersistentFlags().Lookup(PluginPathKey)),
		viper.BindPFlag(CredentialsFileKey, rootCmd.PersistentFlags().Lookup(CredentialsFileKey)),
		viper.BindPFlag(AdminApiEndpoint, rootCmd.PersistentFlags().Lookup(AdminApiEndpoint)),
	)

	if viper.IsSet(ConfigFileKey) {
		cfgFile := os.ExpandEnv(viper.GetString(ConfigFileKey))
		viper.SetConfigFile(cfgFile)

		if err := viper.ReadInConfig(); err != nil {
			return nil, err
		}
	}

	rootCmd.AddCommand(
		install(),
		listRepositories(),
		joinSubnet(),
	)

	return rootCmd, nil
}
