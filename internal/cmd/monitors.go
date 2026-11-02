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
	var mType, mTarget, mName string
	var mInterval int
	create := &cobra.Command{
		Use:   "create",
		Short: "Create a monitor",
		RunE: func(c *cobra.Command, args []string) error {
			if pageID == "" {
				return fmt.Errorf("--status-page is required")
			}
			if mType == "" || mTarget == "" {
				return fmt.Errorf("--type and --target are required")
			}
			cl, err := client()
			if err != nil {
				return err
			}
			body := map[string]any{
				"status_page_id": pageID,
				"type":           mType,
				"target":         mTarget,
				"name":           mName,
			}
			if mInterval > 0 {
				body["interval"] = mInterval
			}
			var resp any
			if err := cl.Post("/monitors", body, &resp); err != nil {
				return err
			}
			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	create.Flags().StringVar(&mType, "type", "", "monitor type (http, tcp, ping, heartbeat, ...)")
	create.Flags().StringVar(&mTarget, "target", "", "monitor target (url, host:port, ...)")
	create.Flags().StringVar(&mName, "name", "", "monitor name")
	create.Flags().IntVar(&mInterval, "interval", 0, "check interval in seconds")
	create.Flags().StringVar(&pageID, "status-page", "", "status page id")
	action := func(verb, method, suffix string) *cobra.Command {
		return &cobra.Command{
			Use:   verb + " <id>",
			Short: verb + " a monitor",
			Args:  cobra.ExactArgs(1),
			RunE: func(c *cobra.Command, args []string) error {
				cl, err := client()
				if err != nil {
					return err
				}
				if method == "DELETE" {
					return cl.Delete("/monitors/" + args[0] + suffix)
				}
				return cl.Post("/monitors/"+args[0]+suffix, nil, nil)
			},
		}
	}
	c.AddCommand(list, create, action("pause", "POST", "/pause"), action("resume", "POST", "/resume"), action("delete", "DELETE", ""))
	return c
}
