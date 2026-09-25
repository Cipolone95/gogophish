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

		merged, added, skipped := mergeTargets(group.Targets, imported)
		group.Targets = merged

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

// mergeTargets appends any target from incoming whose email isn't already
// present in existing, so re-importing the same CSV (or migrating a group
// that already exists on the destination) doesn't create duplicate targets.
// It returns the merged slice and counts of added vs. skipped duplicates.
func mergeTargets(existing, incoming []gophish.Target) (merged []gophish.Target, added, skipped int) {
	merged = existing
	for _, t := range incoming {
		duplicate := false
		for _, e := range merged {
			if e.Email == t.Email {
				duplicate = true
				break
			}
		}
		if duplicate {
			skipped++
			continue
		}
		merged = append(merged, t)
		added++
	}
	return merged, added, skipped
}
