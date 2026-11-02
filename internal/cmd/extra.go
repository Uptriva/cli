package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func webhooksCmd() *cobra.Command {
	c := &cobra.Command{Use: "webhooks", Short: "Manage outgoing webhooks"}
	list := &cobra.Command{
		Use:   "list",
		Short: "List webhooks",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			var raw []map[string]any
			if err := cl.Get("/webhooks", &raw); err != nil {
				return err
			}
			if outputJSON {
				b, _ := json.MarshalIndent(raw, "", "  ")
				fmt.Println(string(b))
				return nil
			}
			if len(raw) == 0 {
				fmt.Println("No webhooks.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tPROVIDER\tURL\tACTIVE")
			for _, wh := range raw {
				id, _ := wh["id"].(string)
				if len(id) > 8 {
					id = id[:8]
				}
				active := "no"
				if b, _ := wh["is_active"].(bool); b {
					active = "yes"
				}
				fmt.Fprintf(w, "%s\t%v\t%v\t%v\t%s\n", id, wh["name"], wh["provider"], wh["url"], active)
			}
			return w.Flush()
		},
	}
	var wURL, wEvents string
	var wPage string
	create := &cobra.Command{
		Use:   "create",
		Short: "Create a webhook",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			body := map[string]any{"url": wURL}
			if wPage != "" {
				body["status_page_id"] = wPage
			}
			if wEvents != "" {
				body["events"] = wEvents
			}
			var resp any
			if err := cl.Post("/webhooks", body, &resp); err != nil {
				return err
			}
			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	create.Flags().StringVar(&wURL, "url", "", "webhook url")
	create.Flags().StringVar(&wPage, "status-page", "", "status page id")
	create.Flags().StringVar(&wEvents, "events", "", "events filter")
	del := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a webhook",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			return cl.Delete("/webhooks/" + args[0])
		},
	}
	test := &cobra.Command{
		Use:   "test <id>",
		Short: "Send a test webhook",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			var resp any
			if err := cl.Post("/webhooks/"+args[0]+"/test", nil, &resp); err != nil {
				return err
			}
			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	c.AddCommand(list, create, del, test)
	return c
}

