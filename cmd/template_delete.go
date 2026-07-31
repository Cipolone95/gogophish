package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/cipolone95/gogophish/internal/gophish"
	"github.com/spf13/cobra"
)

var templateDeleteYes bool

var templateDeleteCmd = &cobra.Command{
	Use:   "delete <name|id|pattern>",
	Short: "Delete template(s) by name, ID, or wildcard pattern",
	Example: `  gogophish template delete example-template1  # exact name
  gogophish template delete 42                  # by ID
  gogophish template delete "example-*"         # wildcard (quote to avoid shell expansion)
  gogophish template delete "example-*" -y      # skip confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		all, err := client.GetTemplates()
		if err != nil {
			return err
		}

		matches, err := matchTemplates(all, args[0])
		if err != nil {
			return err
		}
		if len(matches) == 0 {
			return fmt.Errorf("no templates found matching %q", args[0])
		}

		if !templateDeleteYes {
			if len(matches) == 1 {
				fmt.Printf("Delete template %q (ID: %d)? [y/N]: ",
					matches[0].Name, matches[0].ID)
			} else {
				fmt.Println("The following templates will be deleted:")
				for _, t := range matches {
					fmt.Printf("  [%d] %s\n", t.ID, t.Name)
				}
				fmt.Printf("Delete %d templates? [y/N]: ", len(matches))
			}
			r := bufio.NewReader(os.Stdin)
			answer, _ := r.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(answer)) != "y" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		for _, t := range matches {
			if err := client.DeleteTemplate(t.ID); err != nil {
				fmt.Fprintf(os.Stderr, "error deleting %q (ID: %d): %v\n", t.Name, t.ID, err)
				continue
			}
			fmt.Printf("Deleted template %q (ID: %d)\n", t.Name, t.ID)
		}
		return nil
	},
}

func init() {
	templateCmd.AddCommand(templateDeleteCmd)
	templateDeleteCmd.Flags().BoolVarP(&templateDeleteYes, "yes", "y", false, "skip confirmation prompt")
}

// matchTemplates resolves a numeric ID, wildcard pattern, or exact name against the template list.
func matchTemplates(templates []gophish.Template, arg string) ([]gophish.Template, error) {
	idx, err := matchByIDOrName(len(templates), func(i int) (int64, string) {
		return templates[i].ID, templates[i].Name
	}, arg)
	if err != nil {
		return nil, err
	}
	matches := make([]gophish.Template, len(idx))
	for i, j := range idx {
		matches[i] = templates[j]
	}
	return matches, nil
}
