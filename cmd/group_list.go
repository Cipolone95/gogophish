package cmd

import (
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var groupListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all target groups",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		groups, err := client.GetGroups()
		if err != nil {
			return err
		}

		if len(groups) == 0 {
			fmt.Println("No groups found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tTARGETS\tMODIFIED")
		for _, g := range groups {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n",
				g.ID,
				g.Name,
				strconv.Itoa(len(g.Targets)),
				formatDate(g.ModifiedDate),
			)
		}
		return w.Flush()
	},
}

func init() {
	groupCmd.AddCommand(groupListCmd)
}
