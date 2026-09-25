package cmd

import (
	"fmt"

	"github.com/cipolone95/gogophish/internal/gophish"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	migrateSourceConfig string
	migrateDestConfig   string
	migrateAll          bool
	migrateTemplates    []string
	migrateGroups       []string
	migrateSMTP         []string
	migrateDryRun       bool
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate templates, groups, and sending profiles between GoPhish instances",
	Long: `Copies email templates, target groups, and sending profiles from one
GoPhish instance to another.

Campaigns are intentionally NOT migrated. GoPhish has no API to import a
campaign as historical data — recreating one on the destination would launch
a brand new, live campaign and re-send real email to its targets. If you want
to relaunch a campaign on the destination, do it deliberately with
"gogophish campaign create".

Source and destination are each a config file in the same format produced by
"gogophish config create" (url / api_key / insecure).

--template/--group/--smtp each accept a name, a numeric ID, or a wildcard
pattern (the same matching "gogophish <resource> delete" uses), and are
repeatable. To migrate every item of just ONE resource type — e.g. all
templates but no groups or sending profiles — pass that flag with the
wildcard "*" instead of using --all, which migrates all three types together.
--all cannot be combined with --template/--group/--smtp.`,
	Example: `  # generate the two config files first
  gogophish config create --url https://old.example.com --api-key OLDKEY -o old.yaml
  gogophish config create --url https://new.example.com --api-key NEWKEY -o new.yaml

  # migrate everything (all templates, all groups, all sending profiles)
  gogophish migrate --source-config old.yaml --dest-config new.yaml --all

  # migrate ALL templates, but no groups or sending profiles
  gogophish migrate --source-config old.yaml --dest-config new.yaml --template "*"

  # migrate specific items only
  gogophish migrate --source-config old.yaml --dest-config new.yaml \
    --template "Q1 Phish" --group "Sales Team" --smtp "Internal SMTP"

  # preview what would happen without changing the destination
  gogophish migrate --source-config old.yaml --dest-config new.yaml --all --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if migrateSourceConfig == "" {
			return fmt.Errorf("--source-config is required")
		}
		if migrateDestConfig == "" {
			return fmt.Errorf("--dest-config is required")
		}

		selective := len(migrateTemplates) > 0 || len(migrateGroups) > 0 || len(migrateSMTP) > 0
		if migrateAll && selective {
			return fmt.Errorf("--all cannot be combined with --template/--group/--smtp")
		}
		if !migrateAll && !selective {
			return fmt.Errorf("nothing to migrate — pass --all or one or more of --template/--group/--smtp")
		}

		src, err := clientFromConfigFile(migrateSourceConfig)
		if err != nil {
			return fmt.Errorf("loading --source-config: %w", err)
		}
		dest, err := clientFromConfigFile(migrateDestConfig)
		if err != nil {
			return fmt.Errorf("loading --dest-config: %w", err)
		}

		if migrateDryRun {
			fmt.Println("Dry run — no changes will be made on the destination.")
		}

		if err := migrateGroupsData(src, dest); err != nil {
			return err
		}
		if err := migrateTemplatesData(src, dest); err != nil {
			return err
		}
		if err := migrateSMTPData(src, dest); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().StringVar(&migrateSourceConfig, "source-config", "", "config file for the source GoPhish instance (required)")
	migrateCmd.Flags().StringVar(&migrateDestConfig, "dest-config", "", "config file for the destination GoPhish instance (required)")
	migrateCmd.Flags().BoolVar(&migrateAll, "all", false, "migrate all templates, groups, and sending profiles")
	migrateCmd.Flags().StringArrayVar(&migrateTemplates, "template", nil, `migrate the template matching this name, ID, or wildcard pattern (repeatable; use "*" for all templates)`)
	migrateCmd.Flags().StringArrayVar(&migrateGroups, "group", nil, `migrate the group matching this name, ID, or wildcard pattern (repeatable; use "*" for all groups)`)
	migrateCmd.Flags().StringArrayVar(&migrateSMTP, "smtp", nil, `migrate the sending profile matching this name, ID, or wildcard pattern (repeatable; use "*" for all sending profiles)`)
	migrateCmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "show what would be migrated without changing the destination")
}

// clientFromConfigFile builds a GoPhish client from a standalone config file
// in the "gogophish config create" format. It uses its own viper instance
// rather than the package-level one, since migrate needs two independent
// instances (source and destination) at once.
func clientFromConfigFile(path string) (*gophish.Client, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	url := v.GetString("url")
	apiKey := v.GetString("api_key")
	if url == "" {
		return nil, fmt.Errorf("%s: url not set", path)
	}
	if apiKey == "" {
		return nil, fmt.Errorf("%s: api_key not set", path)
	}

	return gophish.NewClient(url, apiKey, v.GetBool("insecure")), nil
}

func migrateGroupsData(src, dest *gophish.Client) error {
	if !migrateAll && len(migrateGroups) == 0 {
		return nil
	}

	srcGroups, err := src.GetGroups()
	if err != nil {
		return fmt.Errorf("listing source groups: %w", err)
	}

	selected, err := selectGroups(srcGroups, migrateGroups)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		return nil
	}

	destGroups, err := dest.GetGroups()
	if err != nil {
		return fmt.Errorf("listing destination groups: %w", err)
	}
	destByName := make(map[string]gophish.Group, len(destGroups))
	for _, g := range destGroups {
		destByName[g.Name] = g
	}

	fmt.Printf("Groups (%d selected):\n", len(selected))
	for _, g := range selected {
		if existing, ok := destByName[g.Name]; ok {
			merged, added, skipped := mergeTargets(existing.Targets, g.Targets)
			if migrateDryRun {
				fmt.Printf("  %q exists on destination — would merge (%d new, %d duplicate)\n", g.Name, added, skipped)
				continue
			}
			existing.Targets = merged
			updated, err := dest.UpdateGroup(existing)
			if err != nil {
				fmt.Printf("  error merging %q: %v\n", g.Name, err)
				continue
			}
			fmt.Printf("  merged %q (%d new, %d duplicate, %d total)\n", updated.Name, added, skipped, len(updated.Targets))
			continue
		}

		if migrateDryRun {
			fmt.Printf("  would create %q (%d targets)\n", g.Name, len(g.Targets))
			continue
		}
		created, err := dest.CreateGroup(g.Name, g.Targets)
		if err != nil {
			fmt.Printf("  error creating %q: %v\n", g.Name, err)
			continue
		}
		fmt.Printf("  created %q (ID: %d, %d targets)\n", created.Name, created.ID, len(created.Targets))
	}
	return nil
}

func migrateTemplatesData(src, dest *gophish.Client) error {
	if !migrateAll && len(migrateTemplates) == 0 {
		return nil
	}

	srcTemplates, err := src.GetTemplates()
	if err != nil {
		return fmt.Errorf("listing source templates: %w", err)
	}

	selected, err := selectTemplates(srcTemplates, migrateTemplates)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		return nil
	}

	destTemplates, err := dest.GetTemplates()
	if err != nil {
		return fmt.Errorf("listing destination templates: %w", err)
	}
	destNames := make(map[string]bool, len(destTemplates))
	for _, t := range destTemplates {
		destNames[t.Name] = true
	}

	fmt.Printf("Templates (%d selected):\n", len(selected))
	for _, t := range selected {
		if destNames[t.Name] {
			fmt.Printf("  %q already exists on destination, skipping\n", t.Name)
			continue
		}
		if migrateDryRun {
			fmt.Printf("  would create %q\n", t.Name)
			continue
		}
		t.ID = 0
		created, err := dest.CreateTemplate(t)
		if err != nil {
			fmt.Printf("  error creating %q: %v\n", t.Name, err)
			continue
		}
		fmt.Printf("  created %q (ID: %d)\n", created.Name, created.ID)
	}
	return nil
}

func migrateSMTPData(src, dest *gophish.Client) error {
	if !migrateAll && len(migrateSMTP) == 0 {
		return nil
	}

	srcProfiles, err := src.GetSMTPProfiles()
	if err != nil {
		return fmt.Errorf("listing source sending profiles: %w", err)
	}

	selected, err := selectSMTPProfiles(srcProfiles, migrateSMTP)
	if err != nil {
		return err
	}
	if len(selected) == 0 {
		return nil
	}

	destProfiles, err := dest.GetSMTPProfiles()
	if err != nil {
		return fmt.Errorf("listing destination sending profiles: %w", err)
	}
	destNames := make(map[string]bool, len(destProfiles))
	for _, s := range destProfiles {
		destNames[s.Name] = true
	}

	fmt.Printf("Sending profiles (%d selected):\n", len(selected))
	for _, s := range selected {
		if destNames[s.Name] {
			fmt.Printf("  %q already exists on destination, skipping\n", s.Name)
			continue
		}
		if s.Password == "" {
			fmt.Printf("  warning: %q has no password in the source API response — you may need to set one manually after migrating\n", s.Name)
		}
		if migrateDryRun {
			fmt.Printf("  would create %q\n", s.Name)
			continue
		}
		s.ID = 0
		created, err := dest.CreateSMTPProfile(s)
		if err != nil {
			fmt.Printf("  error creating %q: %v\n", s.Name, err)
			continue
		}
		fmt.Printf("  created %q (ID: %d)\n", created.Name, created.ID)
	}
	return nil
}

// selectGroups resolves --all or a set of --group name/ID/wildcard values
// against the source group list, deduplicating overlapping matches.
func selectGroups(all []gophish.Group, patterns []string) ([]gophish.Group, error) {
	if migrateAll {
		return all, nil
	}
	seen := make(map[int64]bool)
	var selected []gophish.Group
	for _, pattern := range patterns {
		matches, err := matchGroups(all, pattern)
		if err != nil {
			return nil, err
		}
		if len(matches) == 0 {
			fmt.Printf("warning: no source group matches %q, skipping\n", pattern)
			continue
		}
		for _, g := range matches {
			if seen[g.ID] {
				continue
			}
			seen[g.ID] = true
			selected = append(selected, g)
		}
	}
	return selected, nil
}

// selectTemplates resolves --all or a set of --template name/ID/wildcard
// values against the source template list, deduplicating overlapping matches.
func selectTemplates(all []gophish.Template, patterns []string) ([]gophish.Template, error) {
	if migrateAll {
		return all, nil
	}
	seen := make(map[int64]bool)
	var selected []gophish.Template
	for _, pattern := range patterns {
		matches, err := matchTemplates(all, pattern)
		if err != nil {
			return nil, err
		}
		if len(matches) == 0 {
			fmt.Printf("warning: no source template matches %q, skipping\n", pattern)
			continue
		}
		for _, t := range matches {
			if seen[t.ID] {
				continue
			}
			seen[t.ID] = true
			selected = append(selected, t)
		}
	}
	return selected, nil
}

// selectSMTPProfiles resolves --all or a set of --smtp name/ID/wildcard
// values against the source sending profile list, deduplicating overlapping
// matches.
func selectSMTPProfiles(all []gophish.SMTP, patterns []string) ([]gophish.SMTP, error) {
	if migrateAll {
		return all, nil
	}
	seen := make(map[int64]bool)
	var selected []gophish.SMTP
	for _, pattern := range patterns {
		matches, err := matchSMTPProfiles(all, pattern)
		if err != nil {
			return nil, err
		}
		if len(matches) == 0 {
			fmt.Printf("warning: no source sending profile matches %q, skipping\n", pattern)
			continue
		}
		for _, s := range matches {
			if seen[s.ID] {
				continue
			}
			seen[s.ID] = true
			selected = append(selected, s)
		}
	}
	return selected, nil
}
