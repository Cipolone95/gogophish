package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all email templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		templates, err := client.GetTemplates()
		if err != nil {
			return err
		}

		if len(templates) == 0 {
			fmt.Println("No templates found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSUBJECT\tHTML\tTEXT")
		for _, t := range templates {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
				t.ID,
				t.Name,
				t.Subject,
				yesNo(t.HTML != ""),
				yesNo(t.Text != ""),
			)
		}
		return w.Flush()
	},
}

func init() {
	templateCmd.AddCommand(templateListCmd)
}

func yesNo(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
