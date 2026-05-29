/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	appversion "portunix.ai/app/version"
)

const githubReleasesAPI = "https://api.github.com/repos/cassandragargoyle/portunix/releases/latest"

var (
	versionCheckCurrent bool
	versionCheckGitHub  bool
	versionCheckSuggest bool
)

var versionCheckCmd = &cobra.Command{
	Use:   "check",
	Short: "Show current version, last GitHub release, and suggest next +dev.N",
	Long: `Check version state per ADR-036 versioning strategy.

By default, prints all three values:
  - Current local version (embedded in this binary)
  - Last GitHub stable release (queried via GitHub API)
  - Suggested next internal development version (+dev.N)

Use flags to limit output for scripting.`,
	Example: `  portunix version check
  portunix version check --suggest
  portunix version check --github`,
	Run: runVersionCheck,
}

func init() {
	versionCheckCmd.Flags().BoolVar(&versionCheckCurrent, "current", false, "Show only current local version")
	versionCheckCmd.Flags().BoolVar(&versionCheckGitHub, "github", false, "Show only last GitHub release tag")
	versionCheckCmd.Flags().BoolVar(&versionCheckSuggest, "suggest", false, "Show only suggested next +dev.N version")
	versionCmd.AddCommand(versionCheckCmd)
}

func runVersionCheck(cmd *cobra.Command, args []string) {
	// "all" mode when no specific flag is set
	all := !versionCheckCurrent && !versionCheckGitHub && !versionCheckSuggest

	current := appversion.ProductVersion

	if all || versionCheckCurrent {
		fmt.Printf("Current local version : %s\n", current)
	}

	githubTag, githubErr := fetchLatestGitHubTag()
	if all || versionCheckGitHub {
		if githubErr != nil {
			fmt.Printf("Last GitHub release   : (unavailable: %v)\n", githubErr)
		} else {
			fmt.Printf("Last GitHub release   : %s\n", githubTag)
		}
	}

	if all || versionCheckSuggest {
		if githubErr != nil {
			if !all && versionCheckSuggest {
				// Single-purpose call failed
				fmt.Println("(unavailable: could not determine base GitHub version)")
				return
			}
			fmt.Printf("Suggested next dev    : (unavailable: could not determine base GitHub version)\n")
		} else {
			suggested := suggestNextDevVersion(githubTag)
			fmt.Printf("Suggested next dev    : %s\n", suggested)
		}
	}
}

// fetchLatestGitHubTag queries the GitHub REST API for the latest release tag.
// Returns the tag name (e.g. "v1.9.2") on success.
func fetchLatestGitHubTag() (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest(http.MethodGet, githubReleasesAPI, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "portunix-version-check")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var payload struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", fmt.Errorf("failed to decode GitHub response: %w", err)
	}
	if payload.TagName == "" {
		return "", fmt.Errorf("GitHub response did not include a tag_name")
	}
	return payload.TagName, nil
}

// suggestNextDevVersion returns the next +dev.N tag for the given GitHub base.
// It lists local git tags matching "{base}+dev.*" and picks max(N)+1.
// Falls back to "{base}+dev.1" if no matching tags exist or git is unavailable.
func suggestNextDevVersion(base string) string {
	pattern := base + "+dev.*"
	out, err := exec.Command("git", "tag", "-l", pattern).Output()
	if err != nil {
		// Not a git repo or git not available — start at dev.1
		return base + "+dev.1"
	}

	re := regexp.MustCompile(`^` + regexp.QuoteMeta(base) + `\+dev\.(\d+)$`)
	var nums []int
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		m := re.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		nums = append(nums, n)
	}

	if len(nums) == 0 {
		return base + "+dev.1"
	}
	sort.Ints(nums)
	return fmt.Sprintf("%s+dev.%d", base, nums[len(nums)-1]+1)
}
