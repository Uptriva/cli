package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/uptriva/cli/internal/api"
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

func findMonitor(cl *api.Client, arg string) (*monitor, error) {
	var all []monitor
	if err := cl.Get("/monitors", &all); err != nil {
		return nil, err
	}
	low := strings.ToLower(arg)
	var match *monitor
	for i, m := range all {
		if strings.HasPrefix(m.ID, arg) || strings.EqualFold(m.Name, arg) {
			return &all[i], nil
		}
		if strings.Contains(low, strings.ToLower(m.Name)) || strings.Contains(strings.ToLower(m.Name), low) {
			match = &all[i]
		}
	}
	if match == nil {
		return nil, fmt.Errorf("no monitor matching %q", arg)
	}
	return match, nil
}

type monitor struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Status       string   `json:"status"`
	Target       string   `json:"target"`
	StatusPageID string   `json:"status_page_id"`
	Regions      []string `json:"regions"`
	Latency      float64  `json:"avg_latency_24h"`
	Avg          float64  `json:"avg_latency"`
}

func pageNames(cl *api.Client) map[string]string {
	var pages []map[string]any
	if err := cl.Get("/status-pages", &pages); err != nil {
		return nil
	}
	out := map[string]string{}
	for _, p := range pages {
		id, _ := p["id"].(string)
		name, _ := p["name"].(string)
		out[id] = name
	}
	return out
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
			pages := pageNames(cl)
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "\033[2mID\tNAME\tTYPE\tSTATUS\tREGIONS\tPAGE\tTARGET\033[0m")
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
				regions := "-"
				if len(m.Regions) > 0 {
					regions = strings.Join(m.Regions, ",")
				}
				page := "-"
				if n, ok := pages[m.StatusPageID]; ok {
					page = n
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", id, m.Name, m.Type, colorStatus(m.Status), regions, page, m.Target)
			}
			return w.Flush()
		},
	}
	list.Flags().StringVar(&pageID, "status-page", "", "filter by status page id")
	list.Flags().StringVar(&fType, "type", "", "filter by monitor type")
	list.Flags().StringVar(&fStatus, "status", "", "filter by status (up, down, degraded, paused)")
	var mType, mTarget, mName, mAgent, mSettings, mGameType string
	var mInterval int
	create := &cobra.Command{
		Use:     "create",
		Short:   "Create a monitor",
		Example: "  uptriva monitors create --type http --target https://example.com --status-page <id> --name Example",
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
			if mAgent != "" {
				body["agent_id"] = mAgent
			}
			if mSettings != "" {
				var s any
				if err := json.Unmarshal([]byte(mSettings), &s); err != nil {
					return fmt.Errorf("--settings must be valid JSON: %w", err)
				}
				body["settings"] = s
			}
			if mGameType != "" {
				body["game_type"] = mGameType
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
	create.Flags().StringVar(&mAgent, "agent", "", "agent id for agent-bound check types")
	create.Flags().StringVar(&mSettings, "settings", "", "monitor settings as JSON (e.g. '{\"tls_mode\":\"tls\"}')")
	create.Flags().StringVar(&mGameType, "game-type", "", "game type for game_server monitors (fivem, minecraft, rust, cs2, ark, valheim, custom)")
	create.Flags().StringVar(&pageID, "status-page", "", "status page id")
	past := map[string]string{"pause": "Paused", "resume": "Resumed", "delete": "Deleted"}
	action := func(verb, method, suffix string) *cobra.Command {
		return &cobra.Command{
			Use:     verb + " <id|name>",
			Short:   past[verb] + " a monitor",
			Args:    cobra.ExactArgs(1),
			Example: "  uptriva monitors " + verb + " Website",
			RunE: func(c *cobra.Command, args []string) error {
				cl, err := client()
				if err != nil {
					return err
				}
				m, err := findMonitor(cl, args[0])
				if err != nil {
					return err
				}
				if method == "DELETE" {
					if err := cl.Delete("/monitors/" + m.ID + suffix); err != nil {
						return err
					}
					fmt.Printf("%s monitor %q\n", past[verb], m.Name)
					return nil
				}
				status := "pending"
				if verb == "pause" {
					status = "paused"
				}
				var resp any
				if err := cl.Put("/monitors/"+m.ID, map[string]string{"status": status}, &resp); err != nil {
					return err
				}
				fmt.Printf("%s monitor %q\n", past[verb], m.Name)
				return nil
			},
		}
	}
	show := &cobra.Command{
		Use:     "show <id|name>",
		Short:   "Show monitor details",
		Args:    cobra.ExactArgs(1),
		Example: "  uptriva monitors show Website\n  uptriva monitors show 8b289bd7",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			var all []monitor
			if err := cl.Get("/monitors", &all); err != nil {
				return err
			}
			match, err := findMonitor(cl, args[0])
			if err != nil {
				return err
			}
			var detail map[string]any
			if err := cl.Get("/monitors/"+match.ID, &detail); err != nil {
				return err
			}
			if outputJSON {
				b, _ := json.MarshalIndent(detail, "", "  ")
				fmt.Println(string(b))
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			row := func(k string, v any) {
				if v != nil && fmt.Sprint(v) != "" {
					fmt.Fprintf(w, "%s\t%v\n", "\033[2m"+k+"\033[0m", v)
				}
			}
			row("id", detail["id"])
			row("name", detail["name"])
			row("type", detail["type"])
			row("status", colorStatus(fmt.Sprint(detail["status"])))
			row("target", detail["target"])
			row("interval", fmt.Sprintf("%vs", detail["interval"]))
			row("regions", detail["regions"])
			row("uptime 30d", detail["uptime_percent"])
			row("avg latency", detail["avg_latency_24h"])
			row("depends on", detail["depends_on"])
			row("agent", detail["agent_id"])
			w.Flush()
			if checks, ok := detail["checks"].([]any); ok && len(checks) > 0 {
				fmt.Println("\nrecent checks")
				w = tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
				fmt.Fprintln(w, "\033[2mTIME\tSTATUS\tLATENCY\tMESSAGE\033[0m")
				n := len(checks)
				if n > 10 {
					checks = checks[:10]
					_ = n
				}
				for _, ch := range checks {
					cm := ch.(map[string]any)
					t, _ := cm["created_at"].(string)
					if len(t) >= 19 {
						t = t[11:19]
					}
					fmt.Fprintf(w, "%s\t%s\t%vms\t%v\n", t, colorStatus(fmt.Sprint(cm["status"])), cm["latency"], cm["message"])
				}
				w.Flush()
			}
			return nil
		},
	}
	var uName, uTarget string
	var uInterval int
	update := &cobra.Command{
		Use:     "update <id|name>",
		Short:   "Update a monitor",
		Args:    cobra.ExactArgs(1),
		Example: "  uptriva monitors update Website --interval 60",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			m, err := findMonitor(cl, args[0])
			if err != nil {
				return err
			}
			body := map[string]any{}
			if uName != "" {
				body["name"] = uName
			}
			if uTarget != "" {
				body["target"] = uTarget
			}
			if uInterval > 0 {
				body["interval"] = uInterval
			}
			if len(body) == 0 {
				return fmt.Errorf("nothing to update — pass --name, --target or --interval")
			}
			var resp any
			if err := cl.Put("/monitors/"+m.ID, body, &resp); err != nil {
				return err
			}
			fmt.Printf("Updated monitor %q\n", m.Name)
			return nil
		},
	}
	update.Flags().StringVar(&uName, "name", "", "new name")
	update.Flags().StringVar(&uTarget, "target", "", "new target")
	update.Flags().IntVar(&uInterval, "interval", 0, "new interval seconds")
	c.AddCommand(list, create, show, update, action("pause", "POST", "/pause"), action("resume", "POST", "/resume"), action("delete", "DELETE", ""))
	return c
}
