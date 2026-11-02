package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"encoding/json"
	"github.com/spf13/cobra"
)

func statusPagesCmd() *cobra.Command {
	c := &cobra.Command{Use: "status-pages", Short: "List status pages"}
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
