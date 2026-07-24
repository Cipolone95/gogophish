package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/cipolone95/gogophish/internal/gophish"
	"github.com/spf13/cobra"
)

var deleteYes bool

var campaignDeleteCmd = &cobra.Command{
	Use:   "delete <name|id|pattern>",
	Short: "Delete campaign(s) by name, ID, or wildcard pattern",
	Example: `  gogophish campaign delete example-email1      # exact name
  gogophish campaign delete 42                  # by ID
  gogophish campaign delete "example-*"         # wildcard (quote to avoid shell expansion)
  gogophish campaign delete "example-*" -y      # skip confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		all, err := client.GetCampaigns()
		if err != nil {
			return err
		}

		matches, err := matchCampaigns(all, args[0])
		if err != nil {
			return err
		}
		if len(matches) == 0 {
			return fmt.Errorf("no campaigns found matching %q", args[0])
		}

		if !deleteYes {
			if len(matches) == 1 {
				fmt.Printf("Delete %q (ID: %d, status: %s)? [y/N]: ",
					matches[0].Name, matches[0].ID, matches[0].Status)
			} else {
				fmt.Println("The following campaigns will be deleted:")
				for _, c := range matches {
					fmt.Printf("  [%d] %s (%s)\n", c.ID, c.Name, c.Status)
				}
				fmt.Printf("Delete %d campaigns? [y/N]: ", len(matches))
			}
			r := bufio.NewReader(os.Stdin)
			answer, _ := r.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(answer)) != "y" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		for _, c := range matches {
			if err := client.DeleteCampaign(c.ID); err != nil {
				fmt.Fprintf(os.Stderr, "error deleting %q (ID: %d): %v\n", c.Name, c.ID, err)
				continue
			}
			fmt.Printf("Deleted %q (ID: %d)\n", c.Name, c.ID)
		}
		return nil
	},
}

func init() {
	campaignCmd.AddCommand(campaignDeleteCmd)
	campaignDeleteCmd.Flags().BoolVarP(&deleteYes, "yes", "y", false, "skip confirmation prompt")
}

// matchCampaigns resolves a numeric ID, wildcard pattern, or exact name against the campaign list.
func matchCampaigns(campaigns []gophish.Campaign, arg string) ([]gophish.Campaign, error) {
	if id, err := strconv.ParseInt(arg, 10, 64); err == nil {
		for _, c := range campaigns {
			if c.ID == id {
				return []gophish.Campaign{c}, nil
			}
		}
		return nil, nil
	}

	if strings.ContainsAny(arg, "*?[") {
		var matches []gophish.Campaign
		for _, c := range campaigns {
			ok, err := path.Match(arg, c.Name)
			if err != nil {
				return nil, fmt.Errorf("invalid pattern %q: %w", arg, err)
			}
			if ok {
				matches = append(matches, c)
			}
		}
		return matches, nil
	}

	for _, c := range campaigns {
		if c.Name == arg {
			return []gophish.Campaign{c}, nil
		}
	}
	return nil, nil
}
