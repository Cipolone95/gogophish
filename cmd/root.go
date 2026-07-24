package cmd

import (
	"fmt"
	"os"

	"github.com/cipolone95/gogophish/internal/gophish"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "gogophish",
	Short: "CLI tool for managing GoPhish campaigns",
	Long:  `gogophish lets you list, copy, and delete GoPhish campaigns from the command line.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default $HOME/.gogophish.yaml)")
	rootCmd.PersistentFlags().String("url", "", "GoPhish server URL (e.g. https://localhost:3333)")
	rootCmd.PersistentFlags().String("api-key", "", "GoPhish API key")
	rootCmd.PersistentFlags().Bool("insecure", false, "skip TLS certificate verification")

	viper.BindPFlag("url", rootCmd.PersistentFlags().Lookup("url"))         //nolint:errcheck
	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key")) //nolint:errcheck
	viper.BindPFlag("insecure", rootCmd.PersistentFlags().Lookup("insecure")) //nolint:errcheck
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(".gogophish")
		viper.SetConfigType("yaml")
	}
	viper.ReadInConfig() //nolint:errcheck
}

func newClient() (*gophish.Client, error) {
	serverURL := viper.GetString("url")
	apiKey := viper.GetString("api_key")
	insecure := viper.GetBool("insecure")

	if serverURL == "" {
		return nil, fmt.Errorf("GoPhish URL not set — use --url flag or set 'url' in ~/.gogophish.yaml")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("GoPhish API key not set — use --api-key flag or set 'api_key' in ~/.gogophish.yaml")
	}
	return gophish.NewClient(serverURL, apiKey, insecure), nil
}
