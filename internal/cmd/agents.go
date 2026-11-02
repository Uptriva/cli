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
	return c
}

func watchCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "watch",
		Short: "Live monitor status (refreshes every 10s, Ctrl+C to exit)",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			for {
				var raw []monitor
				if err := cl.Get("/monitors", &raw); err != nil {
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
				fmt.Printf("uptriva watch — %s (refresh 10s, ctrl+c quit)\n\n", time.Now().Format("15:04:05"))
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
