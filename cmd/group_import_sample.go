package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// sampleGroupCSV mirrors the format expected by GoPhish's bulk import
// endpoint (POST /api/import/group). Columns are matched case-insensitively
// by GoPhish and can appear in any order; this order matches the sample
// shown in GoPhish's own user guide.
const sampleGroupCSV = `First Name,Last Name,Position,Email
Richard,Bourne,CEO,rbourne@example.com
Boyd,Jenius,Systems Administrator,bjenius@example.com
Haiti,Moreo,Sales & Marketing,hmoreo@example.com
`

var groupImportSampleOutput string

var groupImportSampleCmd = &cobra.Command{
	Use:   "import-sample",
	Short: "Print a sample CSV file for bulk-importing group targets",
	Long: `Prints a sample CSV in the format GoPhish expects for bulk-importing
targets into a group (the same format used by "Bulk Import Users" in the
GoPhish UI, and the POST /api/import/group API endpoint).

Columns are matched case-insensitively and may appear in any order:
First Name, Last Name, Email, Position. Email is the only field GoPhish
requires for each target.`,
	Example: `  gogophish group import-sample
  gogophish group import-sample --output targets.csv`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if groupImportSampleOutput == "" {
			fmt.Print(sampleGroupCSV)
			return nil
		}

		if err := os.WriteFile(groupImportSampleOutput, []byte(sampleGroupCSV), 0o644); err != nil {
			return err
		}
		fmt.Printf("Wrote sample CSV to %s\n", groupImportSampleOutput)
		return nil
	},
}

func init() {
	groupCmd.AddCommand(groupImportSampleCmd)
	groupImportSampleCmd.Flags().StringVarP(&groupImportSampleOutput, "output", "o", "", "write sample CSV to this path instead of stdout")
}
