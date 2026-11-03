package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type applyDoc struct {
	StatusPage string `yaml:"status_page"`
	Monitors   []struct {
		Name     string `yaml:"name"`
		Type     string `yaml:"type"`
		Target   string `yaml:"target"`
		Interval int    `yaml:"interval"`
	} `yaml:"monitors"`
	Contacts []struct {
		Name  string `yaml:"name"`
		Email string `yaml:"email"`
	} `yaml:"contacts"`
	Groups []struct {
		Name string `yaml:"name"`
	} `yaml:"groups"`
}

func applyCmd() *cobra.Command {
	var file string
	c := &cobra.Command{
		Use:   "apply -f file.yaml",
		Short: "Apply a declarative config",
		Example: `  uptriva apply -f monitors.yaml

  # monitors.yaml
  status_page: uptriva
  monitors:
    - name: Website
      type: http
      target: https://example.com
      interval: 60`,
		RunE: func(c *cobra.Command, args []string) error {
			if file == "" {
				return fmt.Errorf("-f <file> is required")
			}
			b, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			var doc applyDoc
			if err := yaml.Unmarshal(b, &doc); err != nil {
				return fmt.Errorf("invalid yaml: %w", err)
			}
			cl, err := client()
			if err != nil {
				return err
			}
			pageID := ""
			if doc.StatusPage != "" {
				var pages []map[string]any
				if err := cl.Get("/status-pages", &pages); err != nil {
					return err
				}
				for _, p := range pages {
					pid, _ := p["id"].(string)
					slug, _ := p["slug"].(string)
					if slug == doc.StatusPage || pid == doc.StatusPage {
						pageID = pid
						break
					}
				}
				if pageID == "" {
					return fmt.Errorf("status page %q not found", doc.StatusPage)
				}
			}
			var existing []monitor
			if err := cl.Get("/monitors", &existing); err != nil {
				return err
			}
			byName := map[string]monitor{}
			for _, m := range existing {
				if pageID == "" || m.StatusPageID == pageID {
					byName[m.Name] = m
				}
			}
			created, updated, failed := 0, 0, 0
			for _, m := range doc.Monitors {
				body := map[string]any{"name": m.Name, "type": m.Type, "target": m.Target}
				if m.Interval > 0 {
					body["interval"] = m.Interval
				}
				if ex, ok := byName[m.Name]; ok {
					var resp any
					if err := cl.Put("/monitors/"+ex.ID, body, &resp); err != nil {
						fmt.Printf("  x %-20s %s\n", m.Name, err)
						failed++
						continue
					}
					fmt.Printf("  ~ %-20s updated\n", m.Name)
					updated++
					continue
				}
				if pageID != "" {
					body["status_page_id"] = pageID
				}
				var resp any
				if err := cl.Post("/monitors", body, &resp); err != nil {
					fmt.Printf("  x %-20s %s\n", m.Name, err)
					failed++
					continue
				}
				fmt.Printf("  + %-20s created\n", m.Name)
				created++
			}
			for _, ct := range doc.Contacts {
				body := map[string]any{"name": ct.Name, "status_page_id": pageID}
				if ct.Email != "" {
					body["email"] = ct.Email
				}
				var resp any
				if err := cl.Post("/notify-contacts", body, &resp); err != nil {
					fmt.Printf("  x contact %-14s %s\n", ct.Name, err)
					failed++
					continue
				}
				fmt.Printf("  + contact %-14s created\n", ct.Name)
				created++
			}
			for _, g := range doc.Groups {
				var resp any
				if err := cl.Post("/groups", map[string]any{"name": g.Name, "status_page_id": pageID}, &resp); err != nil {
					fmt.Printf("  x group %-16s %s\n", g.Name, err)
					failed++
					continue
				}
				fmt.Printf("  + group %-16s created\n", g.Name)
				created++
			}
			fmt.Printf("\n%d created, %d updated, %d failed\n", created, updated, failed)
			return nil
		},
	}
	c.Flags().StringVarP(&file, "file", "f", "", "yaml config file")
	return c
}
