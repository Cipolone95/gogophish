package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var groupCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new target group",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		group, err := client.CreateGroup(args[0], nil)
		if err != nil {
			return err
		}

		fmt.Printf("Created group %q (ID: %d)\n", group.Name, group.ID)
		return nil
	},
}

func init() {
	groupCmd.AddCommand(groupCreateCmd)
}
