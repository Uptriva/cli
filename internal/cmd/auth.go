package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/uptriva/cli/internal/config"
	"golang.org/x/term"
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
					if term.IsTerminal(int(os.Stdin.Fd())) {
						b, err := term.ReadPassword(int(os.Stdin.Fd()))
						fmt.Fprintln(os.Stderr)
						if err != nil || strings.TrimSpace(string(b)) == "" {
							return fmt.Errorf("no token provided")
						}
						token = strings.TrimSpace(string(b))
					} else {
						sc := bufio.NewScanner(os.Stdin)
						if !sc.Scan() || strings.TrimSpace(sc.Text()) == "" {
							return fmt.Errorf("no token provided")
						}
						token = strings.TrimSpace(sc.Text())
					}
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
			masked := fmt.Sprintf("<set, %d chars>", len(cfg.Token))
			if len(cfg.Token) > 12 {
				masked = cfg.Token[:4] + "…" + cfg.Token[len(cfg.Token)-2:]
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
