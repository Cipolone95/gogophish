package cmd

import (
	"fmt"
	"os"

	"github.com/cipolone95/gogophish/internal/gophish"
	"github.com/spf13/cobra"
)

var (
	templateSubject  string
	templateHTMLFile string
	templateTextFile string
)

var templateCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create an email template",
	Example: `  # HTML-only template
  gogophish template create "Q1 Phish" --subject "Action Required" --html-file email.html

  # HTML + plain-text fallback
  gogophish template create "Q1 Phish" --subject "Action Required" --html-file email.html --text-file email.txt`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if templateSubject == "" {
			return fmt.Errorf("--subject is required")
		}
		if templateHTMLFile == "" && templateTextFile == "" {
			return fmt.Errorf("at least one of --html-file or --text-file is required")
		}

		t := gophish.Template{
			Name:    args[0],
			Subject: templateSubject,
		}

		if templateHTMLFile != "" {
			b, err := os.ReadFile(templateHTMLFile)
			if err != nil {
				return fmt.Errorf("reading --html-file: %w", err)
			}
			t.HTML = string(b)
		}

		if templateTextFile != "" {
			b, err := os.ReadFile(templateTextFile)
			if err != nil {
				return fmt.Errorf("reading --text-file: %w", err)
			}
			t.Text = string(b)
		}

		client, err := newClient()
		if err != nil {
			return err
		}

		created, err := client.CreateTemplate(t)
		if err != nil {
			return err
		}

		fmt.Printf("Created template %q (ID: %d)\n", created.Name, created.ID)
		return nil
	},
}

func init() {
	templateCmd.AddCommand(templateCreateCmd)
	templateCreateCmd.Flags().StringVar(&templateSubject, "subject", "", "email subject line (required)")
	templateCreateCmd.Flags().StringVar(&templateHTMLFile, "html-file", "", "path to HTML body file")
	templateCreateCmd.Flags().StringVar(&templateTextFile, "text-file", "", "path to plain-text body file")
}
