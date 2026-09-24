package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var groupEmailsOutput string

var groupEmailsCmd = &cobra.Command{
	Use:   "emails <name|id>",
	Short: "Print a group's target email addresses",
	Example: `  gogophish group emails "Sales Team"
  gogophish group emails "Sales Team" --output sales-team.txt
  gogophish group emails 42 -o targets.txt`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		all, err := client.GetGroups()
		if err != nil {
			return err
		}

		matches, err := matchGroups(all, args[0])
		if err != nil {
			return err
		}
		if len(matches) == 0 {
			return fmt.Errorf("no groups found matching %q", args[0])
		}
		if len(matches) > 1 {
			return fmt.Errorf("%q matches %d groups; use the exact name or ID", args[0], len(matches))
		}
		group := matches[0]

		if len(group.Targets) == 0 {
			fmt.Printf("No targets found in %q.\n", group.Name)
			return nil
		}

		emails := make([]string, len(group.Targets))
		for i, t := range group.Targets {
			emails[i] = t.Email
		}
		output := strings.Join(emails, "\n") + "\n"

		if groupEmailsOutput == "" {
			fmt.Print(output)
			return nil
		}

		if err := os.WriteFile(groupEmailsOutput, []byte(output), 0o644); err != nil {
			return err
		}
		fmt.Printf("Wrote %d email address(es) from %q to %s\n", len(emails), group.Name, groupEmailsOutput)
		return nil
	},
}

func init() {
	groupCmd.AddCommand(groupEmailsCmd)
	groupEmailsCmd.Flags().StringVarP(&groupEmailsOutput, "output", "o", "", "write email addresses to this file instead of stdout (one per line)")
}
