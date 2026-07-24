package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

var campaignListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all campaigns",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		campaigns, err := client.GetCampaigns()
		if err != nil {
			return err
		}

		if len(campaigns) == 0 {
			fmt.Println("No campaigns found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSTATUS\tCREATED\tLAUNCH DATE")
		for _, c := range campaigns {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
				c.ID,
				c.Name,
				c.Status,
				formatDate(c.CreatedDate),
				formatDate(c.LaunchDate),
			)
		}
		return w.Flush()
	},
}

func init() {
	campaignCmd.AddCommand(campaignListCmd)
}

func formatDate(s string) string {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil || t.Year() <= 1 {
		return "-"
	}
	return t.Format("2006-01-02 15:04")
}
