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
			fmt.Fprintln(w, "ID\tURL\tACTIVE")
			for _, wh := range raw {
				id, _ := wh["id"].(string)
				if len(id) > 8 {
					id = id[:8]
				}
				fmt.Fprintf(w, "%s\t%v\t%v\n", id, wh["url"], wh["active"])
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

func membersCmd() *cobra.Command {
	c := &cobra.Command{Use: "members", Short: "Manage status page members"}
	var pageID string
	list := &cobra.Command{
		Use:   "list",
		Short: "List members",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			if pageID == "" {
				return fmt.Errorf("--status-page is required")
			}
			var raw any
			if err := cl.Get("/status-pages/"+pageID+"/members", &raw); err != nil {
				return err
			}
			b, _ := json.MarshalIndent(raw, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	var mEmail, mRole string
	invite := &cobra.Command{
		Use:   "invite",
		Short: "Invite a member",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			if pageID == "" {
				return fmt.Errorf("--status-page is required")
			}
			var resp any
			if err := cl.Post("/status-pages/"+pageID+"/members", map[string]any{"email": mEmail, "role": mRole}, &resp); err != nil {
				return err
			}
			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	invite.Flags().StringVar(&mEmail, "email", "", "member email")
	invite.Flags().StringVar(&mRole, "role", "viewer", "member role")
	c.PersistentFlags().StringVar(&pageID, "status-page", "", "status page id")
	c.AddCommand(list, invite)
	return c
}
