package cmd

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

const repo = "Uptriva/cli"

var updateHTTPClient = &http.Client{Timeout: 60 * time.Second}

const maxAssetBytes = 256 << 20

func updateCmd() *cobra.Command {
	var check, force bool
	c := &cobra.Command{
		Use:   "update",
		Short: "Update uptriva to the latest release",
		RunE: func(c *cobra.Command, args []string) error {
			rel, err := latestRelease()
			if err != nil {
				return err
			}
			current := "v" + strings.TrimPrefix(appVersion, "v")
			if current == rel.Tag {
				fmt.Printf("already up to date (%s)\n", appVersion)
				return nil
			}
			if semver.IsValid(current) && semver.IsValid(rel.Tag) && semver.Compare(rel.Tag, current) <= 0 && !force {
				return fmt.Errorf("latest release %s is not newer than current %s (use --force to downgrade)", rel.Tag, appVersion)
			}
			fmt.Printf("current %s — latest %s\n", appVersion, rel.Tag)
			if check {
				fmt.Println("update available")
				return nil
			}
			asset := assetName(rel.Tag)
			var url, checksumsURL string
			for _, a := range rel.Assets {
				if a.Name == asset {
					url = a.URL
				}
				if strings.HasSuffix(a.Name, "checksums.txt") {
					checksumsURL = a.URL
				}
			}
			if url == "" {
				return fmt.Errorf("no release asset for %s/%s", runtime.GOOS, runtime.GOARCH)
			}
			if checksumsURL == "" {
				return fmt.Errorf("no checksums.txt in release %s — refusing to install unverified binary", rel.Tag)
			}
			expected, err := expectedChecksum(checksumsURL, asset)
			if err != nil {
				return err
			}
			fmt.Printf("downloading %s\n", asset)
			archive, err := download(url)
			if err != nil {
				return err
			}
			sum := sha256.Sum256(archive)
			if got := hex.EncodeToString(sum[:]); !strings.EqualFold(got, expected) {
				return fmt.Errorf("checksum mismatch for %s (expected %s, got %s) — refusing to install", asset, expected, got)
			}
			bin, err := extractBinary(archive)
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
	c.Flags().BoolVar(&force, "force", false, "allow installing an older or equal version")
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
	resp, err := updateHTTPClient.Get("https://api.github.com/repos/" + repo + "/releases/latest")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned %d", resp.StatusCode)
	}
	var r release
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

func expectedChecksum(checksumsURL, asset string) (string, error) {
	resp, err := updateHTTPClient.Get(checksumsURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksums.txt download returned %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(b), "\n") {
		f := strings.Fields(line)
		if len(f) == 2 && f[1] == asset {
			return f[0], nil
		}
	}
	return "", fmt.Errorf("no checksum entry for %s in checksums.txt", asset)
}

func assetName(tag string) string {
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("uptriva_%s_%s_%s.%s", strings.TrimPrefix(tag, "v"), runtime.GOOS, runtime.GOARCH, ext)
}

func download(url string) ([]byte, error) {
	resp, err := updateHTTPClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, maxAssetBytes))
}

func extractBinary(archive []byte) ([]byte, error) {
	want := "uptriva"
	if runtime.GOOS == "windows" {
		want += ".exe"
		zr, err := zip.NewReader(bytes.NewReader(archive), int64(len(archive)))
		if err != nil {
			return nil, err
		}
		for _, f := range zr.File {
			if filepath.Base(f.Name) == want {
				rc, err := f.Open()
				if err != nil {
					return nil, err
				}
				defer rc.Close()
				return io.ReadAll(io.LimitReader(rc, maxAssetBytes))
			}
		}
		return nil, fmt.Errorf("binary not found in archive")
	}
	gz, err := gzip.NewReader(bytes.NewReader(archive))
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(gz)
	for {
		h, err := tr.Next()
		if err != nil {
			return nil, fmt.Errorf("binary not found in archive")
		}
		if filepath.Base(h.Name) == want {
			return io.ReadAll(io.LimitReader(tr, maxAssetBytes))
		}
	}
}
