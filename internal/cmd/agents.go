package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

func agentsCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "agents",
		Short: "List monitoring agents",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			var raw []struct {
				ID      string `json:"id"`
				Name    string `json:"name"`
				Hostname string `json:"hostname"`
				Status  string `json:"status"`
				Version string `json:"version"`
			}
			if err := cl.Get("/agents", &raw); err != nil {
				return err
			}
			if outputJSON {
				b, _ := json.MarshalIndent(raw, "", "  ")
				fmt.Println(string(b))
				return nil
			}
			if len(raw) == 0 {
				fmt.Println("No agents.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tHOSTNAME\tSTATUS\tVERSION")
			for _, a := range raw {
				id := a.ID
				if len(id) > 8 {
					id = id[:8]
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", id, a.Name, a.Hostname, a.Status, a.Version)
			}
			return w.Flush()
		},
	}
	var etName string
	enroll := &cobra.Command{
		Use:   "enroll-token",
		Short: "Create an agent enrollment token",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			var resp any
			if err := cl.Post("/agents/enroll-tokens", map[string]any{"name": etName}, &resp); err != nil {
				return err
			}
			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	enroll.Flags().StringVar(&etName, "name", "", "token name")
	c.AddCommand(enroll)
	return c
}

func watchCmd() *cobra.Command {
	var wPage string
	c := &cobra.Command{
		Use:   "watch",
		Short: "Live monitor status (refreshes every 10s, Ctrl+C to exit)",
		Example: "  uptriva watch\n  uptriva watch --status-page uptriva",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			pageID := ""
			pageName := ""
			if wPage != "" {
				var pages []map[string]any
				if err := cl.Get("/status-pages", &pages); err != nil {
					return err
				}
				for _, p := range pages {
					pid, _ := p["id"].(string)
					slug, _ := p["slug"].(string)
					name, _ := p["name"].(string)
					if strings.HasPrefix(pid, wPage) || slug == wPage {
						pageID, pageName = pid, name
						break
					}
				}
				if pageID == "" {
					return fmt.Errorf("no status page matching %q", wPage)
				}
			}
			for {
				var raw []monitor
				path := "/monitors"
				if pageID != "" {
					path += "?status_page_id=" + pageID
				}
				if err := cl.Get(path, &raw); err != nil {
					return err
				}
				fmt.Print("\033[H\033[J")
				up, down, paused := 0, 0, 0
				var rows []monitor
				for _, m := range raw {
					switch {
					case strings.EqualFold(m.Status, "up"):
						up++
					case strings.EqualFold(m.Status, "down"):
						down++
						rows = append([]monitor{m}, rows...)
						continue
					default:
						paused++
					}
					rows = append(rows, m)
				}
				title := "uptriva watch"
				if pageName != "" {
					title += " — " + pageName
				}
				fmt.Printf("%s — %s (refresh 10s, ctrl+c quit)\n\n", title, time.Now().Format("15:04:05"))
				w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
				fmt.Fprintln(w, "\033[2mID\tSTATUS\tLATENCY\tNAME\tTARGET\033[0m")
				for _, m := range rows {
					id := m.ID
					if len(id) > 8 {
						id = id[:8]
					}
					lat := "-"
					l := m.Latency
					if l == 0 {
						l = m.Avg
					}
					if l > 0 {
						lat = fmt.Sprintf("%.0fms", l)
					}
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", id, colorStatus(m.Status), lat, m.Name, m.Target)
				}
				w.Flush()
				fmt.Printf("\n\033[32m%d up\033[0m, \033[31m%d down\033[0m, \033[90m%d paused\033[0m", up, down, paused)
				time.Sleep(10 * time.Second)
			}
		},
	}
	return c
}
