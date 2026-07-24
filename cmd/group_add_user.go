package cmd

import (
	"fmt"

	"github.com/cipolone95/gogophish/internal/gophish"
	"github.com/spf13/cobra"
)

var (
	addUserEmail     string
	addUserFirstName string
	addUserLastName  string
	addUserPosition  string
)

var groupAddUserCmd = &cobra.Command{
	Use:   "add-user <group-name>",
	Short: "Add a single user to a target group",
	Example: `  gogophish group add-user "Sales Team" --email jdoe@example.com --first-name John --last-name Doe
  gogophish group add-user "Sales Team" --email jdoe@example.com --position "Sales Manager"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if addUserEmail == "" {
			return fmt.Errorf("--email is required")
		}

		client, err := newClient()
		if err != nil {
			return err
		}

		group, err := client.GetGroupByName(args[0])
		if err != nil {
			return err
		}

		// check for duplicate email
		for _, t := range group.Targets {
			if t.Email == addUserEmail {
				return fmt.Errorf("%s is already in group %q", addUserEmail, group.Name)
			}
		}

		group.Targets = append(group.Targets, gophish.Target{
			Email:     addUserEmail,
			FirstName: addUserFirstName,
			LastName:  addUserLastName,
			Position:  addUserPosition,
		})

		updated, err := client.UpdateGroup(*group)
		if err != nil {
			return err
		}

		fmt.Printf("Added %s to %q (%d targets total)\n", addUserEmail, updated.Name, len(updated.Targets))
		return nil
	},
}

func init() {
	groupCmd.AddCommand(groupAddUserCmd)
	groupAddUserCmd.Flags().StringVar(&addUserEmail, "email", "", "target email address (required)")
	groupAddUserCmd.Flags().StringVar(&addUserFirstName, "first-name", "", "target first name")
	groupAddUserCmd.Flags().StringVar(&addUserLastName, "last-name", "", "target last name")
	groupAddUserCmd.Flags().StringVar(&addUserPosition, "position", "", "target job position")
}
