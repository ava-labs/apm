package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"github.com/ava-labs/apm/config"
	"github.com/ava-labs/apm/constant"
)

var (
	homeDir = os.ExpandEnv("$HOME")
	apmDir  = filepath.Join(homeDir, fmt.Sprintf(".%s", constant.AppName))
	goPath  = os.ExpandEnv("$GOPATH")
)

const (
	configFileKey       = "config-file"
	apmPathKey          = "apm-path"
	pluginPathKey       = "plugin-path"
	credentialsFileKey  = "credentials-file"
	adminApiEndpointKey = "admin-api-endpoint"
)

func New() (*cobra.Command, error) {
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
	rootCmd.PersistentFlags().String(adminApiEndpointKey, "127.0.0.1:9650/ext/admin", "endpoint for the avalanche admin api")

	errs := wrappers.Errs{}
	errs.Add(
		viper.BindPFlag(configFileKey, rootCmd.PersistentFlags().Lookup(configFileKey)),
		viper.BindPFlag(apmPathKey, rootCmd.PersistentFlags().Lookup(apmPathKey)),
		viper.BindPFlag(pluginPathKey, rootCmd.PersistentFlags().Lookup(pluginPathKey)),
		viper.BindPFlag(credentialsFileKey, rootCmd.PersistentFlags().Lookup(credentialsFileKey)),
		viper.BindPFlag(adminApiEndpointKey, rootCmd.PersistentFlags().Lookup(adminApiEndpointKey)),
	)
	if errs.Errored() {
		return nil, errs.Err
	}

	rootCmd.AddCommand(
		install(),
		listRepositories(),
		joinSubnet(),
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
// nil credentials is safe to use.
func getCredentials() (http.BasicAuth, error) {
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
