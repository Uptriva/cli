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
		b, _ := json.MarshalIndent(resp, "", "  ")
		fmt.Println(string(b))
		_ = os.Stdout
		_ = tabwriter.NewWriter
		return nil
	}
	c.Flags().StringVar(&showID, "show", "", "show a single status page")
	return c
}
