package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"encoding/json"
	"github.com/spf13/cobra"
)

func statusPagesCmd() *cobra.Command {
	c := &cobra.Command{Use: "status-pages", Short: "Manage status pages"}
	var spName, spSlug string
	create := &cobra.Command{
		Use:   "create",
		Short: "Create a status page",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			var resp any
			if err := cl.Post("/status-pages", map[string]any{"name": spName, "slug": spSlug}, &resp); err != nil {
				return err
			}
			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	create.Flags().StringVar(&spName, "name", "", "status page name")
	create.Flags().StringVar(&spSlug, "slug", "", "status page slug")
	show := &cobra.Command{
		Use:   "show <id|slug>",
		Short: "Show a status page",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			var pages []map[string]any
			if err := cl.Get("/status-pages", &pages); err != nil {
				return err
			}
			var id string
			for _, p := range pages {
				pid, _ := p["id"].(string)
				slug, _ := p["slug"].(string)
				if strings.HasPrefix(pid, args[0]) || slug == args[0] {
					id = pid
					break
				}
			}
			if id == "" {
				return fmt.Errorf("no status page matching %q", args[0])
			}
			var resp any
			if err := cl.Get("/status-pages/"+id, &resp); err != nil {
				return err
			}
			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	del := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a status page",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			return cl.Delete("/status-pages/" + args[0])
		},
	}
	c.AddCommand(create, del, show)
	var showID string
	c.RunE = func(c *cobra.Command, args []string) error {
		cl, err := client()
		if err != nil {
			return err
		}
		path := "/status-pages"
		if showID != "" {
			path += "/" + showID
		}
		var resp any
		if err := cl.Get(path, &resp); err != nil {
			return err
		}
		if outputJSON {
			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		pages, ok := resp.([]any)
		if !ok {
			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			return nil
		}
		if len(pages) == 0 {
			fmt.Println("No status pages.")
			return nil
		}
		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSLUG")
		for _, p := range pages {
			pg := p.(map[string]any)
			id, _ := pg["id"].(string)
			if len(id) > 8 {
				id = id[:8]
			}
			fmt.Fprintf(w, "%s\t%v\t%v\n", id, pg["name"], pg["slug"])
		}
		return w.Flush()
	}
	c.Flags().StringVar(&showID, "show", "", "show a single status page")
	return c
}
