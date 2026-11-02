package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uptriva/cli/internal/api"
	"github.com/uptriva/cli/internal/config"
)

var outputJSON bool

func NewRoot(version string) *cobra.Command {
	root := &cobra.Command{
		Use:     "uptriva",
		Short:   "Uptriva CLI — manage monitors, incidents and status pages",
		Version: version,
	}
	root.PersistentFlags().BoolVarP(&outputJSON, "json", "o", false, "output JSON instead of a table")
	root.AddCommand(authCmd(), monitorsCmd(), incidentsCmd(), statusPagesCmd(), agentsCmd(), watchCmd(), contactsCmd(), maintenanceCmd(), groupsCmd(), webhooksCmd(), slaCmd(), heartbeatCmd(), overviewCmd(), agentEventsCmd())
	return root
}

func client() (*api.Client, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	if cfg.Token == "" {
		return nil, fmt.Errorf("not logged in — run `uptriva auth login --token <token>` or set UPTRIVA_TOKEN")
	}
	return api.New(cfg.BaseURL, cfg.Token), nil
}
