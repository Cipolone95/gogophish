package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/cipolone95/gogophish/internal/gophish"
	"github.com/spf13/cobra"
)

var smtpDeleteYes bool

var smtpDeleteCmd = &cobra.Command{
	Use:   "delete <name|id|pattern>",
	Short: "Delete sending profile(s) by name, ID, or wildcard pattern",
	Example: `  gogophish smtp delete "Internal SMTP"   # exact name
  gogophish smtp delete 42                 # by ID
  gogophish smtp delete "Internal *"       # wildcard (quote to avoid shell expansion)
  gogophish smtp delete "Internal *" -y    # skip confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		all, err := client.GetSMTPProfiles()
		if err != nil {
			return err
		}

		matches, err := matchSMTPProfiles(all, args[0])
		if err != nil {
			return err
		}
		if len(matches) == 0 {
			return fmt.Errorf("no sending profiles found matching %q", args[0])
		}

		if !smtpDeleteYes {
			if len(matches) == 1 {
				fmt.Printf("Delete sending profile %q (ID: %d, host: %s)? [y/N]: ",
					matches[0].Name, matches[0].ID, matches[0].Host)
			} else {
				fmt.Println("The following sending profiles will be deleted:")
				for _, s := range matches {
					fmt.Printf("  [%d] %s (%s)\n", s.ID, s.Name, s.Host)
				}
				fmt.Printf("Delete %d sending profiles? [y/N]: ", len(matches))
			}
			r := bufio.NewReader(os.Stdin)
			answer, _ := r.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(answer)) != "y" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		for _, s := range matches {
			if err := client.DeleteSMTPProfile(s.ID); err != nil {
				fmt.Fprintf(os.Stderr, "error deleting %q (ID: %d): %v\n", s.Name, s.ID, err)
				continue
			}
			fmt.Printf("Deleted sending profile %q (ID: %d)\n", s.Name, s.ID)
		}
		return nil
	},
}

func init() {
	smtpCmd.AddCommand(smtpDeleteCmd)
	smtpDeleteCmd.Flags().BoolVarP(&smtpDeleteYes, "yes", "y", false, "skip confirmation prompt")
}

// matchSMTPProfiles resolves a numeric ID, wildcard pattern, or exact name against the sending profile list.
func matchSMTPProfiles(profiles []gophish.SMTP, arg string) ([]gophish.SMTP, error) {
	idx, err := matchByIDOrName(len(profiles), func(i int) (int64, string) {
		return profiles[i].ID, profiles[i].Name
	}, arg)
	if err != nil {
		return nil, err
	}
	matches := make([]gophish.SMTP, len(idx))
	for i, j := range idx {
		matches[i] = profiles[j]
	}
	return matches, nil
}
