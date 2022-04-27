package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ava-labs/avalanchego/utils/wrappers"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"github.com/ava-labs/apm/config"
	"github.com/ava-labs/apm/constant"
)

var (
	homeDir = os.ExpandEnv("$HOME")
	apmDir  = filepath.Join(homeDir, fmt.Sprintf(".%s", constant.AppName))
	goPath  = os.ExpandEnv("$GOPATH")
)

const (
	ConfigFileKey      = "config-file"
	ApmPathKey         = "apm-path"
	PluginPathKey      = "plugin-path"
	CredentialsFileKey = "credentials-file"
	AdminApiEndpoint   = "admin-api-endpoint"
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
	if errs.Errored() {
		return nil, errs.Err
	}

	rootCmd.AddCommand(
		install(),
		listRepositories(),
		joinSubnet(),
	)

	fmt.Printf("credentials file: %s\n", viper.GetString(CredentialsFileKey))

	return rootCmd, nil
}

// initializes config from file, if available.
func initializeConfig() error {
	if viper.IsSet(ConfigFileKey) {
		cfgFile := os.ExpandEnv(viper.GetString(ConfigFileKey))
		viper.SetConfigFile(cfgFile)

		return viper.ReadInConfig()
	}

	return nil
}

// If we need to use custom git credentials (say for private repos).
// nil credentials is safe to use.
func getCredentials() (http.BasicAuth, error) {
	result := http.BasicAuth{}

	if viper.IsSet(CredentialsFileKey) {
		credentials := &config.Credential{}

		bytes, err := os.ReadFile(viper.GetString(CredentialsFileKey))
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
