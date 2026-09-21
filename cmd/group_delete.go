package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/cipolone95/gogophish/internal/gophish"
	"github.com/spf13/cobra"
)

var groupDeleteYes bool

var groupDeleteCmd = &cobra.Command{
	Use:   "delete <name|id|pattern>",
	Short: "Delete group(s) by name, ID, or wildcard pattern",
	Example: `  gogophish group delete "Sales Team"    # exact name
  gogophish group delete 42               # by ID
  gogophish group delete "Sales *"        # wildcard (quote to avoid shell expansion)
  gogophish group delete "Sales *" -y     # skip confirmation`,
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

		if !groupDeleteYes {
			if len(matches) == 1 {
				fmt.Printf("Delete group %q (ID: %d, %d targets)? [y/N]: ",
					matches[0].Name, matches[0].ID, len(matches[0].Targets))
			} else {
				fmt.Println("The following groups will be deleted:")
				for _, g := range matches {
					fmt.Printf("  [%d] %s (%d targets)\n", g.ID, g.Name, len(g.Targets))
				}
				fmt.Printf("Delete %d groups? [y/N]: ", len(matches))
			}
			r := bufio.NewReader(os.Stdin)
			answer, _ := r.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(answer)) != "y" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		for _, g := range matches {
			if err := client.DeleteGroup(g.ID); err != nil {
				fmt.Fprintf(os.Stderr, "error deleting %q (ID: %d): %v\n", g.Name, g.ID, err)
				continue
			}
			fmt.Printf("Deleted group %q (ID: %d)\n", g.Name, g.ID)
		}
		return nil
	},
}

func init() {
	groupCmd.AddCommand(groupDeleteCmd)
	groupDeleteCmd.Flags().BoolVarP(&groupDeleteYes, "yes", "y", false, "skip confirmation prompt")
}

// matchGroups resolves a numeric ID, wildcard pattern, or exact name against the group list.
func matchGroups(groups []gophish.Group, arg string) ([]gophish.Group, error) {
	idx, err := matchByIDOrName(len(groups), func(i int) (int64, string) {
		return groups[i].ID, groups[i].Name
	}, arg)
	if err != nil {
		return nil, err
	}
	matches := make([]gophish.Group, len(idx))
	for i, j := range idx {
		matches[i] = groups[j]
	}
	return matches, nil
}
