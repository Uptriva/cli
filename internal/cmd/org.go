package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

func contactsCmd() *cobra.Command {
	c := &cobra.Command{Use: "contacts", Short: "Manage notification contacts"}
	var listPage string
	list := &cobra.Command{
		Use:   "list",
		Short: "List notification contacts",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			if listPage == "" {
				return fmt.Errorf("--status-page is required")
			}
			var raw []map[string]any
			if err := cl.Get("/notify-contacts?status_page_id="+listPage, &raw); err != nil {
				return err
			}
			if outputJSON {
				b, _ := json.MarshalIndent(raw, "", "  ")
				fmt.Println(string(b))
				return nil
			}
			if len(raw) == 0 {
				fmt.Println("No contacts.")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tTYPE\tTARGET")
			for _, ct := range raw {
				id, _ := ct["id"].(string)
				if len(id) > 8 {
					id = id[:8]
				}
				fmt.Fprintf(w, "%s\t%v\t%v\t%v\n", id, ct["name"], ct["type"], ct["target"])
			}
			return w.Flush()
		},
	}
	list.Flags().StringVar(&listPage, "status-page", "", "status page id")
	var cName, cPage, cEmail, cSlack, cTelegram, cTelegramChat, cSMS, cPhone string
	create := &cobra.Command{
		Use:   "create",
		Short: "Create a notification contact",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			if cPage == "" {
				return fmt.Errorf("--status-page is required")
			}
			body := map[string]any{"name": cName, "status_page_id": cPage}
			if cEmail != "" {
				body["email"] = cEmail
			}
			if cSlack != "" {
				body["slack_webhook_url"] = cSlack
			}
			if cTelegram != "" {
				body["telegram_bot_url"] = cTelegram
			}
			if cTelegramChat != "" {
				body["telegram_chat_id"] = cTelegramChat
			}
			if cSMS != "" {
				body["sms"] = cSMS
			}
			if cPhone != "" {
				body["phone"] = cPhone
			}
			var resp any
			if err := cl.Post("/notify-contacts", body, &resp); err != nil {
				return err
			}
			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	create.Flags().StringVar(&cName, "name", "", "contact name")
	create.Flags().StringVar(&cPage, "status-page", "", "status page id")
	create.Flags().StringVar(&cEmail, "email", "", "email address")
	create.Flags().StringVar(&cSlack, "slack", "", "slack webhook url")
	create.Flags().StringVar(&cTelegram, "telegram", "", "telegram bot url")
	create.Flags().StringVar(&cTelegramChat, "telegram-chat", "", "telegram chat id")
	create.Flags().StringVar(&cSMS, "sms", "", "sms number (+49...)")
	create.Flags().StringVar(&cPhone, "phone", "", "voice call number (+49...)")
	del := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a contact",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			return cl.Delete("/notify-contacts/" + args[0])
		},
	}
	c.AddCommand(list, create, del)
	return c
}

func maintenanceCmd() *cobra.Command {
	c := &cobra.Command{Use: "maintenance", Short: "Manage maintenance windows"}
	var pageID string
	list := &cobra.Command{
		Use:   "list",
		Short: "List maintenance windows",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			path := "/status-pages/" + pageID + "/maintenance"
			var raw any
			if err := cl.Get(path, &raw); err != nil {
				return err
			}
			b, _ := json.MarshalIndent(raw, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	list.Flags().StringVar(&pageID, "status-page", "", "status page id")
	var mTitle, mStart, mEnd string
	create := &cobra.Command{
		Use:   "create",
		Short: "Create a maintenance window",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			if pageID == "" {
				return fmt.Errorf("--status-page is required")
			}
			var resp any
			if err := cl.Post("/status-pages/"+pageID+"/maintenance", map[string]any{"title": mTitle, "start_time": mStart, "end_time": mEnd}, &resp); err != nil {
				return err
			}
			fmt.Printf("Created maintenance %q\n", mTitle)
			_ = resp
			return nil
		},
	}
	create.Flags().StringVar(&mTitle, "title", "", "maintenance title")
	create.Flags().StringVar(&mStart, "start", "", "start time (RFC3339)")
	create.Flags().StringVar(&mEnd, "end", "", "end time (RFC3339)")
	create.Flags().StringVar(&pageID, "status-page", "", "status page id")
	delM := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a maintenance window",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			return cl.Delete("/maintenance/" + args[0])
		},
	}
	c.AddCommand(list, create, delM)
	return c
}

func groupsCmd() *cobra.Command {
	c := &cobra.Command{Use: "groups", Short: "Manage monitor groups"}
	var pageID string
	list := &cobra.Command{
		Use:   "list",
		Short: "List monitor groups",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			var raw any
			if err := cl.Get("/groups", &raw); err != nil {
				return err
			}
			b, _ := json.MarshalIndent(raw, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	var gName string
	create := &cobra.Command{
		Use:   "create",
		Short: "Create a monitor group",
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			var resp any
			if err := cl.Post("/groups", map[string]any{"name": gName, "status_page_id": pageID}, &resp); err != nil {
				return err
			}
			b, _ := json.MarshalIndent(resp, "", "  ")
			fmt.Println(string(b))
			return nil
		},
	}
	create.Flags().StringVar(&gName, "name", "", "group name")
	create.Flags().StringVar(&pageID, "status-page", "", "status page id")
	delG := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a group",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			cl, err := client()
			if err != nil {
				return err
			}
			return cl.Delete("/groups/" + args[0])
		},
	}
	c.AddCommand(list, create, delG)
	return c
}

