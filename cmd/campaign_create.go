package cmd

import (
	"fmt"
	"time"

	"github.com/cipolone95/gogophish/internal/gophish"
	"github.com/spf13/cobra"
)

var (
	campaignCreateTemplate   string
	campaignCreatePage       string
	campaignCreateSMTP       string
	campaignCreatePhishURL   string
	campaignCreateGroups     []string
	campaignCreateLaunchDate string
	campaignCreateSendByDate string
)

var campaignCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new campaign",
	Example: `  gogophish campaign create example-email1 \
    --template "Q1 Phish" --page "Login Page" --smtp "Internal SMTP" \
    --phish-url https://phish.example.com --group "Sales Team"

  gogophish campaign create example-email1 --group "Sales Team" --group "IT Staff" \
    --template "Q1 Phish" --page "Login Page" --smtp "Internal SMTP" --phish-url https://phish.example.com

  # schedule the launch and set a cutoff for when GoPhish should stop sending
  gogophish campaign create example-email1 \
    --template "Q1 Phish" --page "Login Page" --smtp "Internal SMTP" \
    --phish-url https://phish.example.com --group "Sales Team" \
    --launch-date 2026-10-01T09:00:00-04:00 --send-by-date 2026-10-01T17:00:00-04:00`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if campaignCreateTemplate == "" {
			return fmt.Errorf("--template is required")
		}
		if campaignCreatePage == "" {
			return fmt.Errorf("--page is required")
		}
		if campaignCreateSMTP == "" {
			return fmt.Errorf("--smtp is required")
		}
		if campaignCreatePhishURL == "" {
			return fmt.Errorf("--phish-url is required")
		}
		if len(campaignCreateGroups) == 0 {
			return fmt.Errorf("at least one --group is required")
		}

		groups := make([]gophish.NameRef, len(campaignCreateGroups))
		for i, g := range campaignCreateGroups {
			groups[i] = gophish.NameRef{Name: g}
		}

		launchDate := time.Now().UTC().Format(time.RFC3339)
		if campaignCreateLaunchDate != "" {
			launchDate = campaignCreateLaunchDate
		}

		req := gophish.CreateCampaignRequest{
			Name:       args[0],
			Template:   gophish.NameRef{Name: campaignCreateTemplate},
			URL:        campaignCreatePhishURL,
			Page:       gophish.NameRef{Name: campaignCreatePage},
			SMTP:       gophish.NameRef{Name: campaignCreateSMTP},
			LaunchDate: launchDate,
			Groups:     groups,
		}
		if campaignCreateSendByDate != "" {
			req.SendByDate = &campaignCreateSendByDate
		}

		client, err := newClient()
		if err != nil {
			return err
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
	campaignCmd.AddCommand(campaignCreateCmd)
	campaignCreateCmd.Flags().StringVar(&campaignCreateTemplate, "template", "", "email template name (required)")
	campaignCreateCmd.Flags().StringVar(&campaignCreatePage, "page", "", "landing page name (required)")
	campaignCreateCmd.Flags().StringVar(&campaignCreateSMTP, "smtp", "", "sending profile name (required)")
	campaignCreateCmd.Flags().StringVar(&campaignCreatePhishURL, "phish-url", "", "phishing URL for tracking links & the landing page (required) — distinct from the CLI's --url connection flag")
	campaignCreateCmd.Flags().StringArrayVar(&campaignCreateGroups, "group", nil, "target group name (repeatable for multiple groups; at least one required)")
	campaignCreateCmd.Flags().StringVar(&campaignCreateLaunchDate, "launch-date", "", "schedule launch in RFC3339 format (default: now)")
	campaignCreateCmd.Flags().StringVar(&campaignCreateSendByDate, "send-by-date", "", "target completion time in RFC3339 format (optional)")
}
