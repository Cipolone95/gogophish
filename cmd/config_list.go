package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show the currently loaded configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		if path := viper.ConfigFileUsed(); path != "" {
			fmt.Printf("Config file: %s\n\n", path)
		} else {
			fmt.Println("No config file found (using flags/defaults only).")
			fmt.Println()
		}

		fmt.Printf("url:      %s\n", valueOrUnset(viper.GetString("url")))
		fmt.Printf("api_key:  %s\n", maskAPIKey(viper.GetString("api_key")))
		fmt.Printf("insecure: %t\n", viper.GetBool("insecure"))
		return nil
	},
}

func init() {
	configCmd.AddCommand(configListCmd)
}

func valueOrUnset(v string) string {
	if v == "" {
		return "(not set)"
	}
	return v
}

func maskAPIKey(key string) string {
	if key == "" {
		return "(not set)"
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
