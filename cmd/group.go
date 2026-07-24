package cmd

import "github.com/spf13/cobra"

var groupCmd = &cobra.Command{
	Use:   "group",
	Short: "Manage GoPhish target groups",
}

func init() {
	rootCmd.AddCommand(groupCmd)
}
