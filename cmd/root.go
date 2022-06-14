// Copyright (C) 2019-2021, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/ava-labs/apm/apm"
	"github.com/ava-labs/apm/config"
	"github.com/ava-labs/apm/constant"
)

var (
	goPath  = os.ExpandEnv("$GOPATH")
	homeDir = os.ExpandEnv("$HOME")
	apmDir  = filepath.Join(homeDir, fmt.Sprintf(".%s", constant.AppName))
)

const (
	configFileKey       = "config-file"
	apmPathKey          = "apm-path"
	pluginPathKey       = "plugin-path"
	credentialsFileKey  = "credentials-file"
	adminAPIEndpointKey = "admin-api-endpoint"
)

func New(fs afero.Fs) (*cobra.Command, error) {
	rootCmd := &cobra.Command{
		Use:   "apm",
		Short: "apm is a plugin manager to help manage virtual machines and subnets",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// we need to initialize our config here before each command starts,
			// since Cobra doesn't actually parse any of the flags until
			// cobra.Execute() is called.
			return initializeConfig()
		},
	}

	rootCmd.PersistentFlags().String(configFileKey, "", "path to configuration file for the apm")
	rootCmd.PersistentFlags().String(apmPathKey, apmDir, "path to the directory apm creates its artifacts")
	rootCmd.PersistentFlags().String(pluginPathKey, filepath.Join(goPath, "src", "github.com", "ava-labs", "avalanchego", "build", "plugins"), "path to avalanche plugin directory")
	rootCmd.PersistentFlags().String(credentialsFileKey, "", "path to credentials file")
	rootCmd.PersistentFlags().String(adminAPIEndpointKey, "127.0.0.1:9650/ext/admin", "endpoint for the avalanche admin api")

	errs := wrappers.Errs{}
	errs.Add(
		viper.BindPFlag(configFileKey, rootCmd.PersistentFlags().Lookup(configFileKey)),
		viper.BindPFlag(apmPathKey, rootCmd.PersistentFlags().Lookup(apmPathKey)),
		viper.BindPFlag(pluginPathKey, rootCmd.PersistentFlags().Lookup(pluginPathKey)),
		viper.BindPFlag(credentialsFileKey, rootCmd.PersistentFlags().Lookup(credentialsFileKey)),
		viper.BindPFlag(adminAPIEndpointKey, rootCmd.PersistentFlags().Lookup(adminAPIEndpointKey)),
	)
	if errs.Errored() {
		return nil, errs.Err
	}

	rootCmd.AddCommand(
		install(fs),
		uninstall(fs),
		update(fs),
		listRepositories(fs),
		joinSubnet(fs),
	)

	return rootCmd, nil
}

// initializes config from file, if available.
func initializeConfig() error {
	if viper.IsSet(configFileKey) {
		cfgFile := os.ExpandEnv(viper.GetString(configFileKey))
		viper.SetConfigFile(cfgFile)

		return viper.ReadInConfig()
	}

	return nil
}

// If we need to use custom git credentials (say for private repos).
// the zero value for credentials is safe to use.
func initCredentials() (http.BasicAuth, error) {
	result := http.BasicAuth{}

	if viper.IsSet(credentialsFileKey) {
		credentials := &config.Credential{}

		bytes, err := os.ReadFile(viper.GetString(credentialsFileKey))
		if err != nil {
			return result, err
		}
		if err := yaml.Unmarshal(bytes, credentials); err != nil {
			return result, err
		}

		result.Username = credentials.Username
		result.Password = credentials.Password
	}

	return result, nil
}

func initAPM(fs afero.Fs) (*apm.APM, error) {
	credentials, err := initCredentials()
	if err != nil {
		return nil, err
	}

	return apm.New(apm.Config{
		Directory:        viper.GetString(apmPathKey),
		Auth:             credentials,
		AdminAPIEndpoint: viper.GetString(adminAPIEndpointKey),
		PluginDir:        viper.GetString(pluginPathKey),
		Fs:               fs,
	})
}
