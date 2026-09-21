package cmd

import "github.com/spf13/cobra"

var smtpCmd = &cobra.Command{
	Use:   "smtp",
	Short: "Manage GoPhish sending profiles (SMTP)",
}

func init() {
	rootCmd.AddCommand(smtpCmd)
}
