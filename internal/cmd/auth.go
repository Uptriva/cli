package cmd

import (
	"bufio"
	"os"
	"strings"
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
				if t := os.Getenv("UPTRIVA_TOKEN"); t != "" {
					token = t
				} else {
					fmt.Fprint(os.Stderr, "API token: ")
					sc := bufio.NewScanner(os.Stdin)
					if !sc.Scan() || strings.TrimSpace(sc.Text()) == "" {
						return fmt.Errorf("no token provided")
					}
					token = strings.TrimSpace(sc.Text())
				}
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
	whoami := &cobra.Command{
		Use:   "whoami",
		Short: "Show current auth config",
		RunE: func(c *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			if cfg.Token == "" {
				fmt.Println("not authenticated")
				return nil
			}
			masked := cfg.Token
			if len(masked) > 12 {
				masked = masked[:8] + "..." + masked[len(masked)-4:]
			}
			fmt.Printf("endpoint: %s\ntoken: %s\n", cfg.BaseURL, masked)
			return nil
		},
	}
	c.AddCommand(login, logout, whoami)
	return c
}

func mustPath() string {
	p, _ := config.Path()
	return p
}
