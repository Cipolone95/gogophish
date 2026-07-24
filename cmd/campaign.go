package cmd

import "github.com/spf13/cobra"

var campaignCmd = &cobra.Command{
	Use:   "campaign",
	Short: "Manage GoPhish campaigns",
}

func init() {
	rootCmd.AddCommand(campaignCmd)
}
