package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/statikowsky/mar/internal/version"
)

const latestReleaseURL = "https://api.github.com/repos/statikowsky/mar/releases/latest"

var updateClient = &http.Client{Timeout: 5 * time.Second}

type updateResult struct {
	Current         string `json:"current"`
	Latest          string `json:"latest"`
	UpdateAvailable *bool  `json:"update_available"`
	URL             string `json:"url"`
}

func newUpdateCmd() *cobra.Command {
	var asJSON bool
	c := &cobra.Command{
		Use:   "update",
		Short: "Check for a newer mar release",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			release, err := fetchLatestRelease(cmd.Context())
			if err != nil {
				return err
			}
			available, comparable := newerRelease(version.Version, release.TagName)
			result := updateResult{Current: version.Version, Latest: release.TagName, URL: release.HTMLURL}
			if comparable {
				result.UpdateAvailable = &available
			}
			if asJSON {
				return printJSON(cmd, result)
			}
			if !comparable {
				fmt.Fprintf(cmd.OutOrStdout(), "Latest release: %s\nCurrent build: %s (cannot compare development builds)\nRelease: %s\n", result.Latest, result.Current, result.URL)
				return nil
			}
			if !available {
				fmt.Fprintf(cmd.OutOrStdout(), "mar %s is up to date.\n", result.Current)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Update available: %s → %s\nRelease: %s\n\nHomebrew: brew upgrade mar\nGo:       go install github.com/statikowsky/mar@latest\nManual:   download the release above\n", result.Current, result.Latest, result.URL)
			return nil
		},
	}
	c.Flags().BoolVar(&asJSON, "json", false, "output JSON")
	return c
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

func fetchLatestRelease(ctx context.Context) (githubRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, latestReleaseURL, nil)
	if err != nil {
		return githubRelease{}, fmt.Errorf("check for updates: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "mar/"+version.Version)
	resp, err := updateClient.Do(req)
	if err != nil {
		return githubRelease{}, fmt.Errorf("check for updates: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return githubRelease{}, fmt.Errorf("check for updates: GitHub returned %s", resp.Status)
	}
	var release githubRelease
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&release); err != nil {
		return githubRelease{}, fmt.Errorf("check for updates: decode GitHub response: %w", err)
	}
	if _, ok := releaseParts(release.TagName); !ok || release.HTMLURL == "" {
		return githubRelease{}, fmt.Errorf("check for updates: invalid GitHub release")
	}
	return release, nil
}

func newerRelease(current, latest string) (bool, bool) {
	a, ok := releaseParts(current)
	if !ok {
		return false, false
	}
	b, ok := releaseParts(latest)
	if !ok {
		return false, false
	}
	for i := range a {
		if a[i] != b[i] {
			return a[i] < b[i], true
		}
	}
	return false, true
}

func releaseParts(v string) ([3]uint64, bool) {
	var result [3]uint64
	parts := strings.Split(strings.TrimPrefix(v, "v"), ".")
	if !strings.HasPrefix(v, "v") || len(parts) != len(result) {
		return result, false
	}
	for i, part := range parts {
		n, err := strconv.ParseUint(part, 10, 64)
		if err != nil || strconv.FormatUint(n, 10) != part {
			return result, false
		}
		result[i] = n
	}
	return result, true
}
