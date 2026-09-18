package cmd

import (
	"fmt"
	"os"

	"github.com/cipolone95/gogophish/internal/gophish"
	"github.com/spf13/cobra"
)

var (
	templateUpdateSubject         string
	templateUpdateHTMLFile        string
	templateUpdateTextFile        string
	templateUpdateAttachments     []string
	templateUpdateClearAttachment bool
)

var templateUpdateCmd = &cobra.Command{
	Use:   "update <name|id>",
	Short: "Update an existing email template",
	Example: `  # add an attachment to an existing template, keeping everything else
  gogophish template update "Q1 Phish" --attachment invoice.pdf

  # replace an attachment (matched by file name) with a new version
  gogophish template update "Q1 Phish" --attachment invoice.pdf

  # drop all attachments and add a new one
  gogophish template update "Q1 Phish" --clear-attachments --attachment memo.docx

  # update the subject and HTML body only
  gogophish template update "Q1 Phish" --subject "Updated: Action Required" --html-file email-v2.html`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if templateUpdateSubject == "" && templateUpdateHTMLFile == "" && templateUpdateTextFile == "" &&
			len(templateUpdateAttachments) == 0 && !templateUpdateClearAttachment {
			return fmt.Errorf("no changes specified — use --subject, --html-file, --text-file, --attachment, and/or --clear-attachments")
		}

		client, err := newClient()
		if err != nil {
			return err
		}

		all, err := client.GetTemplates()
		if err != nil {
			return err
		}

		matches, err := matchTemplates(all, args[0])
		if err != nil {
			return err
		}
		if len(matches) == 0 {
			return fmt.Errorf("no template found matching %q", args[0])
		}
		if len(matches) > 1 {
			return fmt.Errorf("%q matches %d templates; use the exact name or ID", args[0], len(matches))
		}
		t := matches[0]

		if templateUpdateSubject != "" {
			t.Subject = templateUpdateSubject
		}

		if templateUpdateHTMLFile != "" {
			b, err := os.ReadFile(templateUpdateHTMLFile)
			if err != nil {
				return fmt.Errorf("reading --html-file: %w", err)
			}
			t.HTML = string(b)
		}

		if templateUpdateTextFile != "" {
			b, err := os.ReadFile(templateUpdateTextFile)
			if err != nil {
				return fmt.Errorf("reading --text-file: %w", err)
			}
			t.Text = string(b)
		}

		if templateUpdateClearAttachment {
			t.Attachments = nil
		}
		for _, path := range templateUpdateAttachments {
			attachment, err := loadAttachment(path)
			if err != nil {
				return fmt.Errorf("reading --attachment %s: %w", path, err)
			}
			t.Attachments = setAttachment(t.Attachments, attachment)
		}

		updated, err := client.UpdateTemplate(t)
		if err != nil {
			return err
		}

		fmt.Printf("Updated template %q (ID: %d, %d attachment(s))\n", updated.Name, updated.ID, len(updated.Attachments))
		return nil
	},
}

func init() {
	templateCmd.AddCommand(templateUpdateCmd)
	templateUpdateCmd.Flags().StringVar(&templateUpdateSubject, "subject", "", "new email subject line")
	templateUpdateCmd.Flags().StringVar(&templateUpdateHTMLFile, "html-file", "", "path to new HTML body file")
	templateUpdateCmd.Flags().StringVar(&templateUpdateTextFile, "text-file", "", "path to new plain-text body file")
	templateUpdateCmd.Flags().StringArrayVar(&templateUpdateAttachments, "attachment", nil, "path to a file to attach (repeatable; replaces any existing attachment with the same file name)")
	templateUpdateCmd.Flags().BoolVar(&templateUpdateClearAttachment, "clear-attachments", false, "remove all existing attachments before applying --attachment flags")
}

// setAttachment adds a to existing, replacing any attachment that already
// has the same Name so re-running --attachment with an updated file
// overwrites rather than duplicates it.
func setAttachment(existing []gophish.Attachment, a gophish.Attachment) []gophish.Attachment {
	for i, e := range existing {
		if e.Name == a.Name {
			existing[i] = a
			return existing
		}
	}
	return append(existing, a)
}
