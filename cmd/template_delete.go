package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var templateDeleteYes bool

var templateDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete an email template by name",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}

		tmpl, err := client.GetTemplateByName(args[0])
		if err != nil {
			return err
		}

		if !templateDeleteYes {
			fmt.Printf("Delete template %q (ID: %d)? [y/N]: ", tmpl.Name, tmpl.ID)
			r := bufio.NewReader(os.Stdin)
			answer, _ := r.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(answer)) != "y" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		if err := client.DeleteTemplate(tmpl.ID); err != nil {
			return err
		}

		fmt.Printf("Deleted template %q (ID: %d)\n", tmpl.Name, tmpl.ID)
		return nil
	},
}

func init() {
	templateCmd.AddCommand(templateDeleteCmd)
	templateDeleteCmd.Flags().BoolVarP(&templateDeleteYes, "yes", "y", false, "skip confirmation prompt")
}
