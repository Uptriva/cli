package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func colorStatus(s string) string {
	switch strings.ToLower(s) {
	case "up":
		return "\033[32mup\033[0m"
	case "down":
		return "\033[31mdown\033[0m"
	case "degraded":
		return "\033[33mdegraded\033[0m"
	case "paused":
		return "\033[90mpaused\033[0m"
	default:
		return s
	}
}

type monitor struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	Status  string `json:"status"`
	Target  string `json:"target"`
	Latency float64 `json:"avg_latency_24h"`
	Avg     float64 `json:"avg_latency"`
}

func monitorsCmd() *cobra.Command {
	c := &cobra.Command{Use: "monitors", Short: "Manage monitors"}
	var pageID, fType, fStatus string
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
			var raw []monitor
			if err := cl.Get(path, &raw); err != nil {
				return err
			}
			resp.Monitors = raw
			if outputJSON {
				b, _ := json.MarshalIndent(resp.Monitors, "", "  ")
				fmt.Println(string(b))
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "\033[2mID\tNAME\tTYPE\tSTATUS\tTARGET\033[0m")
			for _, m := range resp.Monitors {
				if fType != "" && m.Type != fType {
					continue
				}
				if fStatus != "" && m.Status != fStatus {
					continue
				}
				id := m.ID
				if len(id) > 8 {
					id = id[:8]
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", id, m.Name, m.Type, colorStatus(m.Status), m.Target)
			}
			return w.Flush()
		},
	}
	list.Flags().StringVar(&pageID, "status-page", "", "filter by status page id")
	list.Flags().StringVar(&fType, "type", "", "filter by monitor type")
	list.Flags().StringVar(&fStatus, "status", "", "filter by status (up, down, degraded, paused)")
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
	show := &cobra.Command{
		Use:   "show <id|name>",
		Short: "Show monitor details",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			var all []monitor
			if err := cl.Get("/monitors", &all); err != nil {
				return err
			}
			arg := strings.ToLower(args[0])
			var match *monitor
			for i, m := range all {
				if strings.HasPrefix(m.ID, args[0]) || strings.EqualFold(m.Name, args[0]) {
					match = &all[i]
					break
				}
				if strings.Contains(strings.ToLower(m.Name), arg) {
					match = &all[i]
				}
			}
			if match == nil {
				return fmt.Errorf("no monitor matching %q", args[0])
			}
			var detail any
			if err := cl.Get("/monitors/"+match.ID, &detail); err != nil {
				detail = map[string]string{"error": err.Error()}
			}
			b, _ := json.MarshalIndent(detail, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	c.AddCommand(list, create, show, action("pause", "POST", "/pause"), action("resume", "POST", "/resume"), action("delete", "DELETE", ""))
	return c
}
