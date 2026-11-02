package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"encoding/json"
	"github.com/spf13/cobra"
)

func incidentsCmd() *cobra.Command {
	c := &cobra.Command{Use: "incidents", Short: "Manage incidents"}
	var pageID string
	list := &cobra.Command{
		Use:   "list",
		Short: "List incidents",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			path := "/incidents"
			if pageID != "" {
				path += "?status_page_id=" + pageID
			}
			var resp []struct {
				ID       string `json:"id"`
				Title    string `json:"title"`
				Severity string `json:"severity"`
				Status   string `json:"status"`
			}
			if err := cl.Get(path, &resp); err != nil {
				return err
			}
			if outputJSON {
				b, _ := json.MarshalIndent(resp, "", "  ")
				fmt.Println(string(b))
				return nil
			}
			if len(resp) == 0 {
				fmt.Println("No incidents.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tTITLE\tSEVERITY\tSTATUS")
			for _, i := range resp {
				id := i.ID
				if len(id) > 8 {
					id = id[:8]
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", id, i.Title, i.Severity, i.Status)
			}
			return w.Flush()
		},
	}
	list.Flags().StringVar(&pageID, "status-page", "", "filter by status page id")
	act := func(verb string) *cobra.Command {
		return &cobra.Command{
			Use:   verb + " <id>",
			Short: verb + " an incident",
			Args:  cobra.ExactArgs(1),
			RunE: func(c *cobra.Command, args []string) error {
				cl, err := client()
				if err != nil {
					return err
				}
				return cl.Post("/incidents/"+args[0]+"/"+verb, nil, nil)
			},
		}
	}
	c.AddCommand(list, act("ack"), act("resolve"))
	return c
}
