package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

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
			var resp any
			if err := cl.Get("/status-pages/"+pageID+"/sla", &resp); err != nil {
				return err
			}
			return printJSON(resp)
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
			var resp any
			if err := cl.Get("/status-pages/"+pageID+"/sla-targets", &resp); err != nil {
				return err
			}
			return printJSON(resp)
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
			var resp any
			if err := cl.Get("/overview", &resp); err != nil {
				return err
			}
			return printJSON(resp)
		},
	}
}

func tierinfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tier-info",
		Short: "Show your plan tier and limits",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			var resp struct {
				Tier   string                 `json:"tier"`
				Usage  map[string]float64     `json:"usage"`
				Limits map[string]interface{} `json:"limits"`
			}
			if err := cl.Get("/tier-info", &resp); err != nil {
				return err
			}
			if outputJSON {
				return printJSON(resp)
			}
			fmt.Printf("plan\t%s\n", resp.Tier)
			pairs := [][2]string{
				{"monitors", "max_monitors"},
				{"agents", "max_agents"},
				{"status_pages", "max_status_pages"},
				{"login_seats", "max_login_seats"},
				{"notify_seats", "max_notify_seats"},
			}
			for _, p := range pairs {
				used, uok := resp.Usage[p[0]]
				max, mok := resp.Limits[p[1]].(float64)
				if uok && mok && max > 0 {
					fmt.Printf("%s\t%d / %d\n", strings.ReplaceAll(p[0], "_", " "), int(used), int(max))
				} else if uok {
					fmt.Printf("%s\t%d\n", strings.ReplaceAll(p[0], "_", " "), int(used))
				}
			}
			if g, ok := resp.Usage["groups"]; ok {
				fmt.Printf("groups\t%d\n", int(g))
			}
			var feats []string
			for k, v := range resp.Limits {
				if b, ok := v.(bool); ok && b {
					feats = append(feats, strings.ReplaceAll(k, "_", " "))
				}
			}
			sort.Strings(feats)
			if len(feats) > 0 {
				fmt.Printf("features\t%s\n", strings.Join(feats, ", "))
			}
			if regions, ok := resp.Limits["allowed_regions"].([]interface{}); ok {
				var rs []string
				for _, r := range regions {
					rs = append(rs, fmt.Sprint(r))
				}
				fmt.Printf("regions\t%s\n", strings.Join(rs, ", "))
			}
			return nil
		},
	}
}

func subscribersCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "subscribers",
		Short: "List status page subscribers",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			var resp any
			if err := cl.Get("/subscribers", &resp); err != nil {
				return err
			}
			return printJSON(resp)
		},
	}
}

func auditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "audit-logs",
		Short: "Show audit log entries",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			var resp any
			if err := cl.Get("/audit-logs", &resp); err != nil {
				return err
			}
			return printJSON(resp)
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
