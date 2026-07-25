package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const sampleConfig = `# gogophish configuration file
# Save this as .gogophish.yaml in the current directory, or pass it via --config <path>.

# GoPhish server URL (e.g. https://localhost:3333)
url: https://localhost:3333

# GoPhish API key, found under Settings in the GoPhish admin UI
api_key: your-api-key-here

# Skip TLS certificate verification (useful for self-signed certs)
insecure: false
`

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage the gogophish CLI configuration file",
}

func init() {
	rootCmd.AddCommand(configCmd)
}

// defaultConfigPath returns the default config file location: .gogophish.yaml
// in the current working directory.
func defaultConfigPath() string {
	wd, err := os.Getwd()
	if err != nil {
		return ".gogophish.yaml"
	}
	return filepath.Join(wd, ".gogophish.yaml")
}
