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
	var iTitle, iSeverity, iPage string
	create := &cobra.Command{
		Use:     "create",
		Short:   "Create an incident",
		Example: "  uptriva incidents create --title 'DB failover' --severity major --status-page <id>",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			if iPage == "" {
				return fmt.Errorf("--status-page is required")
			}
			if iTitle == "" {
				return fmt.Errorf("--title is required")
			}
			var resp any
			if err := cl.Post("/status-pages/"+iPage+"/incidents", map[string]any{"title": iTitle, "severity": iSeverity}, &resp); err != nil {
				return err
			}
			fmt.Printf("Created incident %q\n", iTitle)
			_ = resp
			return nil
		},
	}
	create.Flags().StringVar(&iTitle, "title", "", "incident title")
	create.Flags().StringVar(&iSeverity, "severity", "minor", "severity (minor, major, critical)")
	create.Flags().StringVar(&iPage, "status-page", "", "status page id")
	var uMsg string
	post := &cobra.Command{
		Use:   "update <id>",
		Short: "Post an update to an incident",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			if uMsg == "" {
				return fmt.Errorf("--message is required")
			}
			var resp any
			if err := cl.Post("/incidents/"+args[0]+"/updates", map[string]any{"message": uMsg}, &resp); err != nil {
				return err
			}
			fmt.Println("Update posted.")
			return nil
		},
	}
	post.Flags().StringVar(&uMsg, "message", "", "update message")
	c.AddCommand(list, create, post, act("ack"), act("resolve"), act("publish"), act("unpublish"))
	return c
}
