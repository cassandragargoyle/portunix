/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/google/go-github/v56/github"
	"github.com/spf13/cobra"

	"portunix.ai/portunix/src/helpers/ptx-github/internal/download"
)

var (
	downloadFlagAsset     string
	downloadFlagOutput    string
	downloadFlagPlatform  string
	downloadFlagResume    bool
	downloadFlagSHA256    string
	downloadFlagQuiet     bool
	downloadFlagAll       bool
	downloadFlagOverwrite bool
)

func newDownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download <owner/repo|url> [tag]",
		Short: "Download a release asset",
		Long: `Download a release asset from GitHub.

Tag defaults to the latest release. Use --asset to pick a specific asset
file by name, or --platform to auto-pick a binary for the current OS/arch.
Use --all to download every asset in the release.

Examples:
  portunix github download cassandragargoyle/portunix
  portunix github download cassandragargoyle/portunix v1.5.1 \
      --asset portunix-linux-amd64.tar.gz
  portunix github download cassandragargoyle/portunix v1.5.1 \
      --platform linux-amd64 -o ./dist
  portunix github download cassandragargoyle/portunix v1.5.1 --all -o ./dist`,
		Args: cobra.RangeArgs(1, 2),
		RunE: runDownload,
	}
	cmd.Flags().StringVar(&downloadFlagAsset, "asset", "", "exact asset name to download")
	cmd.Flags().StringVarP(&downloadFlagOutput, "output", "o", "", "output file or directory")
	cmd.Flags().StringVar(&downloadFlagPlatform, "platform", "",
		"select asset by platform (e.g., linux-amd64, windows-amd64, auto)")
	cmd.Flags().BoolVar(&downloadFlagResume, "resume", false, "resume a partially downloaded file")
	cmd.Flags().StringVar(&downloadFlagSHA256, "sha256", "", "verify downloaded file against this SHA-256 (hex)")
	cmd.Flags().BoolVarP(&downloadFlagQuiet, "quiet", "q", false, "suppress progress output")
	cmd.Flags().BoolVar(&downloadFlagAll, "all", false, "download all assets in the release")
	cmd.Flags().BoolVar(&downloadFlagOverwrite, "overwrite", false, "overwrite existing files (without resume)")
	return cmd
}

func runDownload(cmd *cobra.Command, args []string) error {
	owner, repo, err := parseOwnerRepoFlexible(args[0])
	if err != nil {
		return err
	}
	c, err := newClient(context.Background())
	if err != nil {
		return err
	}

	var rel *github.RepositoryRelease
	if len(args) == 2 {
		rel, err = c.ReleaseByTag(owner, repo, args[1])
	} else {
		rel, err = c.LatestRelease(owner, repo)
	}
	if err != nil {
		return err
	}
	if len(rel.Assets) == 0 {
		return fmt.Errorf("release %s has no assets", rel.GetTagName())
	}

	outBase := downloadFlagOutput
	if outBase == "" {
		outBase = "."
	}

	switch {
	case downloadFlagAll:
		return downloadAll(c.Token(), rel.Assets, outBase)
	case downloadFlagAsset != "":
		asset := findAssetByName(rel.Assets, downloadFlagAsset)
		if asset == nil {
			return fmt.Errorf("asset %q not found in release %s", downloadFlagAsset, rel.GetTagName())
		}
		return downloadOne(c.Token(), asset, outBase)
	case downloadFlagPlatform != "":
		platform := downloadFlagPlatform
		if platform == "auto" {
			platform = autoPlatform()
		}
		asset := findAssetByPlatform(rel.Assets, platform)
		if asset == nil {
			return fmt.Errorf("no asset matching platform %q in release %s",
				platform, rel.GetTagName())
		}
		return downloadOne(c.Token(), asset, outBase)
	default:
		// No specific selector — print available assets and exit.
		fmt.Printf("Release %s has %d assets. Use --asset, --platform, or --all:\n",
			rel.GetTagName(), len(rel.Assets))
		for _, a := range rel.Assets {
			fmt.Printf("  %s  (%s)\n", a.GetName(), formatBytes(int64(a.GetSize())))
		}
		return fmt.Errorf("no asset selector provided")
	}
}

func downloadAll(token string, assets []*github.ReleaseAsset, outDir string) error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	for _, a := range assets {
		if err := downloadOne(token, a, outDir); err != nil {
			return err
		}
	}
	return nil
}

func downloadOne(token string, asset *github.ReleaseAsset, outBase string) error {
	dest, err := resolveDest(outBase, asset.GetName())
	if err != nil {
		return err
	}

	if !downloadFlagResume && !downloadFlagOverwrite {
		if _, err := os.Stat(dest); err == nil {
			return fmt.Errorf("file %q already exists (use --resume or --overwrite)", dest)
		}
	}

	var progress download.ProgressCallback
	if !downloadFlagQuiet {
		progress = download.SimpleProgressBar
	}

	fmt.Fprintf(os.Stderr, "Downloading %s -> %s\n", asset.GetName(), dest)
	res, err := download.Download(download.Options{
		URL:         asset.GetBrowserDownloadURL(),
		Dest:        dest,
		Resume:      downloadFlagResume,
		Token:       token,
		ExpectedSHA: downloadFlagSHA256,
		Progress:    progress,
	})
	if err != nil {
		return err
	}
	fmt.Printf("Downloaded %s (%s)\n", res.Path, formatBytes(res.Size))
	if res.SHA256 != "" {
		fmt.Printf("SHA-256: %s\n", res.SHA256)
	}
	return nil
}

// resolveDest returns the output path. If outBase is a directory or doesn't
// exist (and ends with a slash) it is treated as a directory.
func resolveDest(outBase, assetName string) (string, error) {
	if outBase == "" {
		return assetName, nil
	}
	info, err := os.Stat(outBase)
	if err == nil && info.IsDir() {
		return filepath.Join(outBase, assetName), nil
	}
	if os.IsNotExist(err) && (strings.HasSuffix(outBase, string(os.PathSeparator)) ||
		strings.HasSuffix(outBase, "/")) {
		if err := os.MkdirAll(outBase, 0o755); err != nil {
			return "", err
		}
		return filepath.Join(outBase, assetName), nil
	}
	if err == nil {
		// File exists — caller decides whether to overwrite/resume.
		return outBase, nil
	}
	if os.IsNotExist(err) {
		// Treat as a file path.
		if dir := filepath.Dir(outBase); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return "", err
			}
		}
		return outBase, nil
	}
	return "", err
}

func findAssetByName(assets []*github.ReleaseAsset, name string) *github.ReleaseAsset {
	for _, a := range assets {
		if a.GetName() == name {
			return a
		}
	}
	return nil
}

// findAssetByPlatform looks for an asset whose name contains every token of the
// platform spec (e.g., "linux-amd64" -> contains "linux" AND ("amd64" OR "x86_64")).
func findAssetByPlatform(assets []*github.ReleaseAsset, platform string) *github.ReleaseAsset {
	tokens := splitPlatform(platform)
	if len(tokens) == 0 {
		return nil
	}
	var best *github.ReleaseAsset
	bestScore := 0
	for _, a := range assets {
		name := strings.ToLower(a.GetName())
		score := 0
		ok := true
		for _, group := range tokens {
			matched := false
			for _, t := range group {
				if strings.Contains(name, t) {
					matched = true
					break
				}
			}
			if !matched {
				ok = false
				break
			}
			score++
		}
		if ok && score > bestScore {
			best = a
			bestScore = score
		}
	}
	return best
}

// splitPlatform breaks a platform spec into ordered token groups.
// "linux-amd64" -> [["linux"], ["amd64", "x86_64", "x64"]]
func splitPlatform(spec string) [][]string {
	parts := strings.Split(strings.ToLower(spec), "-")
	out := make([][]string, 0, len(parts))
	for _, p := range parts {
		switch p {
		case "amd64", "x86_64", "x64":
			out = append(out, []string{"amd64", "x86_64", "x64"})
		case "arm64", "aarch64":
			out = append(out, []string{"arm64", "aarch64"})
		case "darwin", "macos", "mac":
			out = append(out, []string{"darwin", "macos", "mac"})
		case "windows", "win":
			out = append(out, []string{"windows", "win"})
		default:
			out = append(out, []string{p})
		}
	}
	return out
}

func autoPlatform() string {
	return fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
}

func formatBytes(b int64) string {
	return download.FormatBytes(b)
}
