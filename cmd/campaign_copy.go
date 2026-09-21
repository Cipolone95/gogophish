package cmd

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/cipolone95/gogophish/internal/gophish"
	"github.com/spf13/cobra"
)

var (
	copyLaunchDate string
	copySendByDate string
	copyGroups     []string
)

var campaignCopyCmd = &cobra.Command{
	Use:   "copy <name>",
	Short: "Copy a campaign, incrementing the trailing number in its name",
	Example: `  gogophish campaign copy example-email1 --group "Sales Team"
  # creates example-email2 targeting the same template, page, smtp

  gogophish campaign copy example-email1 --group "Sales Team" --group "IT Staff"
  # target multiple groups

  gogophish campaign copy example-email9 --group "Sales Team"
  # creates example-email10

  gogophish campaign copy example-email1 --group "Sales Team" \
    --launch-date 2026-10-01T09:00:00-04:00 --send-by-date 2026-10-01T17:00:00-04:00
  # schedule the launch and set a cutoff for when GoPhish should stop sending`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		src, err := client.GetCampaignByName(args[0])
		if err != nil {
			return err
		}

		// GoPhish does not return groups in the campaign list response, so
		// prefer --group flags; fall back to whatever the API did return.
		var groups []gophish.NameRef
		switch {
		case len(copyGroups) > 0:
			groups = make([]gophish.NameRef, len(copyGroups))
			for i, g := range copyGroups {
				groups[i] = gophish.NameRef{Name: g}
			}
		case len(src.Groups) > 0:
			groups = make([]gophish.NameRef, len(src.Groups))
			for i, g := range src.Groups {
				groups[i] = gophish.NameRef{Name: g.Name}
			}
		default:
			return fmt.Errorf("no groups found on source campaign — specify one with --group \"Group Name\"")
		}

		newName := incrementName(src.Name)

		launchDate := time.Now().UTC().Format(time.RFC3339)
		if copyLaunchDate != "" {
			launchDate = copyLaunchDate
		}

		req := gophish.CreateCampaignRequest{
			Name:       newName,
			Template:   gophish.NameRef{Name: src.Template.Name},
			URL:        src.URL,
			Page:       &gophish.NameRef{Name: src.Page.Name},
			SMTP:       gophish.NameRef{Name: src.SMTP.Name},
			LaunchDate: launchDate,
			Groups:     groups,
		}
		if copySendByDate != "" {
			req.SendByDate = &copySendByDate
		}

		created, err := client.CreateCampaign(req)
		if err != nil {
			return err
		}

		fmt.Printf("Created %q (ID: %d) — launching %s\n",
			created.Name, created.ID, formatDate(created.LaunchDate))
		return nil
	},
}

func init() {
	campaignCmd.AddCommand(campaignCopyCmd)
	campaignCopyCmd.Flags().StringVar(&copyLaunchDate, "launch-date", "", "schedule launch in RFC3339 format (default: now)")
	campaignCopyCmd.Flags().StringVar(&copySendByDate, "send-by-date", "", "target completion time in RFC3339 format (optional)")
	campaignCopyCmd.Flags().StringArrayVar(&copyGroups, "group", nil, "target group name (repeatable for multiple groups)")
}

var trailingDigits = regexp.MustCompile(`^(.*?)(\d+)$`)

func incrementName(name string) string {
	m := trailingDigits.FindStringSubmatch(name)
	if m == nil {
		return name + "2"
	}
	n, _ := strconv.Atoi(m[2])
	return m[1] + strconv.Itoa(n+1)
}
