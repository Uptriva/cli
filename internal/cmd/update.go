package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const repo = "Uptriva/cli"

func updateCmd() *cobra.Command {
	var check bool
	c := &cobra.Command{
		Use:   "update",
		Short: "Update uptriva to the latest release",
		RunE: func(c *cobra.Command, args []string) error {
			rel, err := latestRelease()
			if err != nil {
				return err
			}
			current := appVersion
			latest := strings.TrimPrefix(rel.Tag, "v")
			if strings.TrimPrefix(current, "v") == latest {
				fmt.Printf("already up to date (%s)\n", current)
				return nil
			}
			fmt.Printf("current %s — latest %s\n", current, rel.Tag)
			if check {
				fmt.Println("update available")
				return nil
			}
			asset := assetName(rel.Tag)
			var url string
			for _, a := range rel.Assets {
				if a.Name == asset {
					url = a.URL
				}
			}
			if url == "" {
				return fmt.Errorf("no release asset for %s/%s", runtime.GOOS, runtime.GOARCH)
			}
			fmt.Printf("downloading %s\n", asset)
			bin, err := download(url)
			if err != nil {
				return err
			}
			exe, err := os.Executable()
			if err != nil {
				return err
			}
			exe, _ = filepath.EvalSymlinks(exe)
			tmp := exe + ".new"
			if err := os.WriteFile(tmp, bin, 0o755); err != nil {
				return err
			}
			if err := os.Rename(tmp, exe); err != nil {
				os.Remove(tmp)
				return err
			}
			fmt.Printf("updated to %s\n", rel.Tag)
			return nil
		},
	}
	c.Flags().BoolVar(&check, "check", false, "only check, don't install")
	return c
}

type release struct {
	Tag    string `json:"tag_name"`
	Assets []struct {
		Name string `json:"name"`
		URL  string `json:"browser_download_url"`
	} `json:"assets"`
}

func latestRelease() (*release, error) {
	resp, err := http.Get("https://api.github.com/repos/" + repo + "/releases/latest")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var r release
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

func assetName(tag string) string {
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("uptriva_%s_%s_%s.%s", strings.TrimPrefix(tag, "v"), runtime.GOOS, runtime.GOARCH, ext)
}

func download(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if runtime.GOOS == "windows" {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		zr, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
		if err != nil {
			return nil, err
		}
		for _, f := range zr.File {
			if strings.HasSuffix(f.Name, "uptriva.exe") {
				rc, _ := f.Open()
				defer rc.Close()
				return io.ReadAll(rc)
			}
		}
		return nil, fmt.Errorf("binary not found in archive")
	}
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err != nil {
			return nil, fmt.Errorf("binary not found in archive")
		}
		if strings.HasSuffix(h.Name, "uptriva") {
			return io.ReadAll(tr)
		}
	}
}
