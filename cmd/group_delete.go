package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var groupDeleteYes bool

var groupDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a target group by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		group, err := client.GetGroupByName(args[0])
		if err != nil {
			return err
		}

		if !groupDeleteYes {
			fmt.Printf("Delete group %q (ID: %d, %d targets)? [y/N]: ", group.Name, group.ID, len(group.Targets))
			r := bufio.NewReader(os.Stdin)
			answer, _ := r.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(answer)) != "y" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		if err := client.DeleteGroup(group.ID); err != nil {
			return err
		}

		fmt.Printf("Deleted group %q (ID: %d)\n", group.Name, group.ID)
		return nil
	},
}

func init() {
	groupCmd.AddCommand(groupDeleteCmd)
	groupDeleteCmd.Flags().BoolVarP(&groupDeleteYes, "yes", "y", false, "skip confirmation prompt")
}
