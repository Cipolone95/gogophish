package cmd

import (
	"fmt"
	"strings"

	"github.com/cipolone95/gogophish/internal/gophish"
	"github.com/spf13/cobra"
)

var (
	smtpCreateHost             string
	smtpCreateFromAddress      string
	smtpCreateUsername         string
	smtpCreatePassword         string
	smtpCreateIgnoreCertErrors bool
	smtpCreateHeaders          []string
)

var smtpCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a sending profile",
	Example: `  gogophish smtp create "Internal SMTP" --host smtp.example.com:587 --from-address phisher@example.com

  gogophish smtp create "Internal SMTP" --host smtp.example.com:587 --from-address phisher@example.com \
    --username relay --password s3cret --ignore-cert-errors

  gogophish smtp create "Internal SMTP" --host smtp.example.com:587 --from-address phisher@example.com \
    --header "X-Mailer: GoPhish" --header "X-Priority: 1"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if smtpCreateHost == "" {
			return fmt.Errorf("--host is required")
		}
		if smtpCreateFromAddress == "" {
			return fmt.Errorf("--from-address is required")
		}

		headers := make([]gophish.Header, 0, len(smtpCreateHeaders))
		for _, h := range smtpCreateHeaders {
			key, value, ok := strings.Cut(h, ":")
			if !ok {
				return fmt.Errorf("invalid --header %q — expected \"Key: Value\"", h)
			}
			headers = append(headers, gophish.Header{Key: strings.TrimSpace(key), Value: strings.TrimSpace(value)})
		}

		s := gophish.SMTP{
			Name:             args[0],
			Host:             smtpCreateHost,
			FromAddress:      smtpCreateFromAddress,
			Username:         smtpCreateUsername,
			Password:         smtpCreatePassword,
			IgnoreCertErrors: smtpCreateIgnoreCertErrors,
			Headers:          headers,
		}

		client, err := newClient()
		if err != nil {
			return err
		}

		created, err := client.CreateSMTPProfile(s)
		if err != nil {
			return err
		}

		fmt.Printf("Created sending profile %q (ID: %d)\n", created.Name, created.ID)
		return nil
	},
}

func init() {
	smtpCmd.AddCommand(smtpCreateCmd)
	smtpCreateCmd.Flags().StringVar(&smtpCreateHost, "host", "", "SMTP host:port (required)")
	smtpCreateCmd.Flags().StringVar(&smtpCreateFromAddress, "from-address", "", "from address used in sent emails (required)")
	smtpCreateCmd.Flags().StringVar(&smtpCreateUsername, "username", "", "SMTP auth username")
	smtpCreateCmd.Flags().StringVar(&smtpCreatePassword, "password", "", "SMTP auth password")
	smtpCreateCmd.Flags().BoolVar(&smtpCreateIgnoreCertErrors, "ignore-cert-errors", false, "skip TLS certificate verification for this profile")
	smtpCreateCmd.Flags().StringArrayVar(&smtpCreateHeaders, "header", nil, `custom email header as "Key: Value" (repeatable)`)
}
