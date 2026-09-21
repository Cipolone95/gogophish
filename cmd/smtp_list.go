package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var smtpListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sending profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		profiles, err := client.GetSMTPProfiles()
		if err != nil {
			return err
		}

		if len(profiles) == 0 {
			fmt.Println("No sending profiles found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tHOST\tFROM ADDRESS\tMODIFIED")
		for _, s := range profiles {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\n",
				s.ID,
				s.Name,
				s.Host,
				s.FromAddress,
				formatDate(s.ModifiedDate),
			)
		}
		return w.Flush()
	},
}

func init() {
	smtpCmd.AddCommand(smtpListCmd)
}
