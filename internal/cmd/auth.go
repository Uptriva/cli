package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uptriva/cli/internal/config"
)

func authCmd() *cobra.Command {
	c := &cobra.Command{Use: "auth", Short: "Manage authentication"}
	var token string
	login := &cobra.Command{
		Use:   "login",
		Short: "Save an API token",
		RunE: func(c *cobra.Command, args []string) error {
			if token == "" {
				return fmt.Errorf("pass --token <uk_... or scoped token>")
			}
			cfg, _ := config.Load()
			cfg.Token = token
			if err := cfg.Save(); err != nil {
				return err
			}
			fmt.Println("Token saved to", mustPath())
			return nil
		},
	}
	login.Flags().StringVar(&token, "token", "", "API token to store")
	logout := &cobra.Command{
		Use:   "logout",
		Short: "Remove the stored token",
		RunE: func(c *cobra.Command, args []string) error {
			cfg, _ := config.Load()
			cfg.Token = ""
			return cfg.Save()
		},
	}
	c.AddCommand(login, logout)
	return c
}

func mustPath() string {
	p, _ := config.Path()
	return p
}
