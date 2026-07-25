package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var configSampleCmd = &cobra.Command{
	Use:   "sample",
	Short: "Print a sample config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print(sampleConfig)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSampleCmd)
}
