package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func printJSON(v any) error {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
	return nil
}

func slaCmd() *cobra.Command {
	c := &cobra.Command{Use: "sla", Short: "View SLA stats and targets"}
	var pageID string
	get := &cobra.Command{
		Use:   "get",
		Short: "Get SLA stats for a status page",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			if pageID == "" {
				return fmt.Errorf("--status-page is required")
			}
			var resp map[string]any
			if err := cl.Get("/status-pages/"+pageID+"/sla", &resp); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(resp)
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			for _, k := range []string{"uptime_percent", "period", "monitors_total", "monitors_meeting", "monitors_failing"} {
				if v, ok := resp[k]; ok {
					fmt.Fprintf(w, "%s\t%v\n", "\033[2m"+k+"\033[0m", v)
				}
			}
			w.Flush()
			for k, v := range resp {
				if _, isMap := v.(map[string]any); isMap {
					fmt.Fprintf(w, "%s\t%v\n", "\033[2m"+k+"\033[0m", v)
				}
			}
			return nil
		},
	}
	targets := &cobra.Command{
		Use:   "targets",
		Short: "Get SLA targets",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			if pageID == "" {
				return fmt.Errorf("--status-page is required")
			}
			var resp map[string]any
			if err := cl.Get("/status-pages/"+pageID+"/sla-targets", &resp); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(resp)
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			for k, v := range resp {
				fmt.Fprintf(w, "%s\t%v\n", "\033[2m"+k+"\033[0m", v)
			}
			return w.Flush()
		},
	}
	c.PersistentFlags().StringVar(&pageID, "status-page", "", "status page id")
	c.AddCommand(get, targets)
	return c
}

func heartbeatCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "heartbeat <id|name>",
		Short:   "Push a heartbeat ping to a heartbeat monitor",
		Args:    cobra.ExactArgs(1),
		Example: "  uptriva heartbeat \"Pinger Falkenstein\"",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			m, err := findMonitor(cl, args[0])
			if err != nil {
				return err
			}
			var detail map[string]any
			if err := cl.Get("/monitors/"+m.ID, &detail); err != nil {
				return err
			}
			var token string
			if settings, ok := detail["settings"].(string); ok {
				var sm map[string]any
				if json.Unmarshal([]byte(settings), &sm) == nil {
					token, _ = sm["heartbeat_token"].(string)
				}
			}
			if tok, ok := detail["settings"].(map[string]any); ok {
				token, _ = tok["heartbeat_token"].(string)
			}
			if token == "" {
				return fmt.Errorf("monitor %q has no heartbeat token — is it a heartbeat monitor?", m.Name)
			}
			var resp any
			if err := cl.PostPublic("/heartbeat/"+m.ID, "Bearer "+token, &resp); err != nil {
				return err
			}
			fmt.Printf("Heartbeat sent for %q\n", m.Name)
			return nil
		},
	}
}

func overviewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "overview",
		Short: "Show account uptime overview",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			var resp struct {
				Tier            string  `json:"tier"`
				GlobalUptime    float64 `json:"global_uptime"`
				AvgResponse     float64 `json:"avg_response_ms"`
				AgentsOnline    int     `json:"agents_online"`
				ActiveIncidents int     `json:"active_incidents"`
				TotalMonitors   int     `json:"total_monitors"`
				TotalPages      int     `json:"total_status_pages"`
				Monitors        []struct {
					Name    string  `json:"name"`
					Status  string  `json:"status"`
					Uptime  float64 `json:"uptime"`
					Latency int     `json:"latency_ms"`
					Page    string  `json:"status_page_name"`
				} `json:"monitors"`
				TopAgents []struct {
					Name   string  `json:"name"`
					CPU    float64 `json:"cpu_usage"`
					RAM    float64 `json:"ram_usage"`
					Disk   float64 `json:"disk_usage"`
					Region string  `json:"region"`
				} `json:"top_agents"`
			}
			if err := cl.Get("/overview", &resp); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(resp)
			}
			fmt.Printf("%.2f%% uptime · %dms avg · %d monitors · %d agents online", resp.GlobalUptime, int(resp.AvgResponse), resp.TotalMonitors, resp.AgentsOnline)
			if resp.ActiveIncidents > 0 {
				fmt.Printf(" · \033[31m%d active incident(s)\033[0m", resp.ActiveIncidents)
			}
			fmt.Println()
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "\033[2mMONITOR\tSTATUS\tUPTIME\tLATENCY\tPAGE\033[0m")
			for _, m := range resp.Monitors {
				lat := "-"
				if m.Latency > 0 {
					lat = fmt.Sprintf("%dms", m.Latency)
				}
				fmt.Fprintf(w, "%s\t%s\t%.2f%%\t%s\t%s\n", m.Name, colorStatus(m.Status), m.Uptime, lat, m.Page)
			}
			w.Flush()
			if len(resp.TopAgents) > 0 {
				fmt.Println("\ntop agents")
				w = tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
				fmt.Fprintln(w, "\033[2mAGENT\tCPU\tRAM\tDISK\tREGION\033[0m")
				for _, a := range resp.TopAgents {
					fmt.Fprintf(w, "%s\t%.0f%%\t%.0f%%\t%.0f%%\t%s\n", a.Name, a.CPU, a.RAM, a.Disk, a.Region)
				}
				w.Flush()
			}
			return nil
		},
	}
}

func agentEventsCmd() *cobra.Command {
	c := &cobra.Command{Use: "agent-events", Short: "Manage agent events"}
	act := func(verb string) *cobra.Command {
		return &cobra.Command{
			Use:   verb + " <id>",
			Short: verb + " an agent event",
			Args:  cobra.ExactArgs(1),
			RunE: func(c *cobra.Command, args []string) error {
				cl, err := client()
				if err != nil {
					return err
				}
				var resp any
				if err := cl.Post("/agent-events/"+args[0]+"/"+verb, nil, &resp); err != nil {
					return err
				}
				fmt.Printf("%sd event %s\n", verb, args[0])
				return nil
			},
		}
	}
	c.AddCommand(act("acknowledge"), act("resolve"), act("reopen"))
	return c
}
