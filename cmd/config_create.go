package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	configCreateURL      string
	configCreateAPIKey   string
	configCreateInsecure bool
	configCreateOutput   string
	configCreateForce    bool
)

var configCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		path := configCreateOutput
		if path == "" {
			path = defaultConfigPath()
		}

		if _, err := os.Stat(path); err == nil && !configCreateForce {
			return fmt.Errorf("config file %s already exists, use --force to overwrite", path)
		}

		content := sampleConfig
		if configCreateURL != "" || configCreateAPIKey != "" || cmd.Flags().Changed("insecure") {
			content = fmt.Sprintf(
				"# gogophish configuration file\n\nurl: %s\napi_key: %s\ninsecure: %t\n",
				configCreateURL, configCreateAPIKey, configCreateInsecure,
			)
		}

		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			return err
		}

		fmt.Printf("Created config file at %s\n", path)
		return nil
	},
}

func init() {
	configCreateCmd.Flags().StringVar(&configCreateURL, "url", "", "GoPhish server URL")
	configCreateCmd.Flags().StringVar(&configCreateAPIKey, "api-key", "", "GoPhish API key")
	configCreateCmd.Flags().BoolVar(&configCreateInsecure, "insecure", false, "skip TLS certificate verification")
	configCreateCmd.Flags().StringVarP(&configCreateOutput, "output", "o", "", "path to write config file (default: ./.gogophish.yaml)")
	configCreateCmd.Flags().BoolVarP(&configCreateForce, "force", "f", false, "overwrite existing config file")
	configCmd.AddCommand(configCreateCmd)
}
