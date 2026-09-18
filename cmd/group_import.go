package cmd

import (
	"fmt"

	"github.com/cipolone95/gogophish/internal/gophish"
	"github.com/spf13/cobra"
)

var groupImportFile string

var groupImportCmd = &cobra.Command{
	Use:   "import <group-name>",
	Short: "Bulk-import targets into a group from a CSV file",
	Example: `  gogophish group import-sample --output targets.csv
  # generate a starter CSV in the format GoPhish expects

  gogophish group import "Sales Team" --file targets.csv
  # creates "Sales Team" if it doesn't exist yet, or adds new targets to it`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if groupImportFile == "" {
			return fmt.Errorf("--file is required")
		}
		name := args[0]

		client, err := newClient()
		if err != nil {
			return err
		}

		imported, err := client.ImportGroupCSV(groupImportFile)
		if err != nil {
			return err
		}
		if len(imported) == 0 {
			return fmt.Errorf("no targets parsed from %s", groupImportFile)
		}

		groups, err := client.GetGroups()
		if err != nil {
			return err
		}

		var group *gophish.Group
		for i := range groups {
			if groups[i].Name == name {
				group = &groups[i]
				break
			}
		}
		isNew := group == nil
		if isNew {
			group = &gophish.Group{Name: name}
		}

		added, skipped := 0, 0
		for _, t := range imported {
			duplicate := false
			for _, existing := range group.Targets {
				if existing.Email == t.Email {
					duplicate = true
					break
				}
			}
			if duplicate {
				skipped++
				continue
			}
			group.Targets = append(group.Targets, t)
			added++
		}

		var result *gophish.Group
		if isNew {
			result, err = client.CreateGroup(group.Name, group.Targets)
		} else {
			result, err = client.UpdateGroup(*group)
		}
		if err != nil {
			return err
		}

		fmt.Printf("Imported %d target(s) from %s into %q (%d new, %d duplicate(s) skipped, %d total)\n",
			len(imported), groupImportFile, result.Name, added, skipped, len(result.Targets))
		return nil
	},
}

func init() {
	groupCmd.AddCommand(groupImportCmd)
	groupImportCmd.Flags().StringVar(&groupImportFile, "file", "", "path to CSV file of targets (required)")
}
