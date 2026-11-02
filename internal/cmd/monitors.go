package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

type monitor struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Target string `json:"target"`
}

func monitorsCmd() *cobra.Command {
	c := &cobra.Command{Use: "monitors", Short: "Manage monitors"}
	var pageID string
	list := &cobra.Command{
		Use:   "list",
		Short: "List monitors",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			path := "/monitors"
			if pageID != "" {
				path += "?status_page_id=" + pageID
			}
			var resp struct {
				Monitors []monitor `json:"monitors"`
			}
			if err := cl.Get(path, &resp); err != nil {
				return err
			}
			if outputJSON {
				b, _ := json.MarshalIndent(resp.Monitors, "", "  ")
				fmt.Println(string(b))
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tTYPE\tSTATUS\tTARGET")
			for _, m := range resp.Monitors {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", m.ID[:8], m.Name, m.Type, m.Status, m.Target)
			}
			return w.Flush()
		},
	}
	list.Flags().StringVar(&pageID, "status-page", "", "filter by status page id")
	c.AddCommand(list)
	return c
}
