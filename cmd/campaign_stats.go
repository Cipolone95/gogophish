package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var campaignStatsJSON bool

// campaignStat is the curated id/email/status view of a campaign result —
// GoPhish's own Result also carries name, position, ip, lat/long, send_date,
// reported, and modified_date, but only these three were asked for.
type campaignStat struct {
	ID     int64  `json:"id"`
	Email  string `json:"email"`
	Status string `json:"status"`
}

var campaignStatsCmd = &cobra.Command{
	Use:   "stats <name|id>",
	Short: "Show per-target results (ID, email, status) for a campaign",
	Example: `  gogophish campaign stats example-email1
  gogophish campaign stats 42
  gogophish campaign stats example-email1 --json`,
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
		if len(matches) > 1 {
			return fmt.Errorf("%q matches %d campaigns; use the exact name or ID", args[0], len(matches))
		}

		results, err := client.GetCampaignResults(matches[0].ID)
		if err != nil {
			return err
		}

		stats := make([]campaignStat, len(results.Results))
		for i, r := range results.Results {
			stats[i] = campaignStat{ID: r.ID, Email: r.Email, Status: r.Status}
		}

		if campaignStatsJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(stats)
		}

		if len(stats) == 0 {
			fmt.Println("No results found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tEMAIL\tSTATUS")
		for _, s := range stats {
			fmt.Fprintf(w, "%d\t%s\t%s\n", s.ID, s.Email, s.Status)
		}
		return w.Flush()
	},
}

func init() {
	campaignCmd.AddCommand(campaignStatsCmd)
	campaignStatsCmd.Flags().BoolVar(&campaignStatsJSON, "json", false, "output as JSON instead of a table")
}
