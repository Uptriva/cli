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
			var resp any
			if err := cl.Get(path, &resp); err != nil {
				return err
			}
			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			_ = os.Stdout
			_ = tabwriter.NewWriter
			return nil
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
