package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var githubToken = os.Getenv("GITHUB_TOKEN")

const (
	sdkRepo    = "gmcorenet/sdk"
	cliVersion = "0.1.0"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "release":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: gmcore release <minor|major|bugfix|v1.0.0>")
			os.Exit(1)
		}
		bumpType := os.Args[2]
		if err := release(bumpType); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

	case "build-framework":
		version := "dev"
		if len(os.Args) >= 3 {
			version = os.Args[2]
		}
		if err := buildFramework(version); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

	case "version", "--version", "-v":
		fmt.Printf("gmcore workspace tool %s\n", cliVersion)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("gmcore - GMCore Workspace Management Tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  gmcore release <type>      Create and push SDK release tag")
	fmt.Println("    type: minor (1.0.0 → 1.1.0)")
	fmt.Println("          major (1.0.0 → 2.0.0)")
	fmt.Println("          bugfix (1.0.0 → 1.0.1)")
	fmt.Println("          or explicit version (e.g., v1.2.3)")
	fmt.Println("  gmcore build-framework [ver]  Build framework tarball locally")
	fmt.Println("  gmcore version                Show version")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  gmcore release minor")
	fmt.Println("  gmcore release v1.0.0")
}

func release(bumpOrVersion string) error {
	var newVersion string

	if strings.HasPrefix(bumpOrVersion, "v") || strings.Contains(bumpOrVersion, ".") {
		if !strings.HasPrefix(bumpOrVersion, "v") {
			bumpOrVersion = "v" + bumpOrVersion
		}
		newVersion = bumpOrVersion
		fmt.Printf("Using explicit version: %s\n", newVersion)
	} else {
		current, err := getLatestVersion()
		if err != nil {
			return fmt.Errorf("failed to get latest version: %w", err)
		}

		switch bumpOrVersion {
		case "minor":
			newVersion = incrementMinor(current)
		case "major":
			newVersion = incrementMajor(current)
		case "bugfix":
			newVersion = incrementBugfix(current)
		default:
			return fmt.Errorf("unknown release type: %s (use: minor, major, bugfix, or v1.0.0)", bumpOrVersion)
		}
		fmt.Printf("Current: %s → New: %s\n", current, newVersion)
	}

	sdkPath, err := getSdkPath()
	if err != nil {
		return err
	}

	fmt.Printf("Creating release %s...\n", newVersion)

	tagCmd := exec.Command("git", "tag", newVersion)
	tagCmd.Dir = sdkPath
	if output, err := tagCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create tag: %s", string(output))
	}

	pushURL := fmt.Sprintf("https://%s@github.com/%s.git", githubToken, sdkRepo)
	pushCmd := exec.Command("git", "push", pushURL, newVersion)
	pushCmd.Dir = sdkPath
	if output, err := pushCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to push tag: %s", string(output))
	}

	fmt.Printf("Release %s pushed! GitHub Actions will build and publish.\n", newVersion)
	fmt.Printf("Watch progress at: https://github.com/%s/actions\n", sdkRepo)

	return nil
}

func getLatestVersion() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", sdkRepo)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return "v0.0.0", nil
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &release); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	return release.TagName, nil
}

func incrementMinor(current string) string {
	return bumpVersion(current, "minor")
}

func incrementMajor(current string) string {
	return bumpVersion(current, "major")
}

func incrementBugfix(current string) string {
	return bumpVersion(current, "bugfix")
}

func bumpVersion(current, bumpType string) string {
	current = strings.TrimPrefix(current, "v")
	parts := strings.Split(current, ".")
	if len(parts) < 3 {
		for len(parts) < 3 {
			parts = append(parts, "0")
		}
	}

	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	patch, _ := strconv.Atoi(parts[2])

	switch bumpType {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "bugfix":
		patch++
	}

	return fmt.Sprintf("v%d.%d.%d", major, minor, patch)
}

func getSdkPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	sdkPath := filepath.Join(wd, "..", "sdk")
	if _, err := os.Stat(sdkPath); err != nil {
		return "", fmt.Errorf("sdk path not found. Expected at: %s", sdkPath)
	}

	if githubToken != "" {
		remoteURL := fmt.Sprintf("https://%s@github.com/%s.git", githubToken, sdkRepo)
		cmd := exec.Command("git", "remote", "set-url", "origin", remoteURL)
		cmd.Dir = sdkPath
		cmd.Run()
	}

	return sdkPath, nil
}

func buildFramework(version string) error {
	fmt.Printf("Building framework %s locally...\n", version)

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get wd: %w", err)
	}

	sdkPath := filepath.Join(wd, "..", "sdk")
	if _, err := os.Stat(sdkPath); os.IsNotExist(err) {
		return fmt.Errorf("sdk path not found at: %s", sdkPath)
	}

	tmpDir, err := os.MkdirTemp("", "gmcore-build-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	frameworkDir := filepath.Join(tmpDir, "gmcore-framework-"+version)
	if err := os.MkdirAll(frameworkDir, 0755); err != nil {
		return fmt.Errorf("failed to create framework dir: %w", err)
	}

	dirs := []string{
		"bin", "cmd/app", "config", "public",
		"var/cache", "var/log", "var/tmp",
		"src", "tests", "migrations", "templates",
		"vendor/gmcore",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(frameworkDir, dir), 0755); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", dir, err)
		}
	}

	sdkEntries, err := os.ReadDir(sdkPath)
	if err != nil {
		return fmt.Errorf("failed to read sdk dir: %w", err)
	}

	for _, entry := range sdkEntries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "gmcore-") {
			continue
		}

		src := filepath.Join(sdkPath, entry.Name())
		dest := filepath.Join(frameworkDir, "vendor/gmcore", entry.Name())

		if err := copyDir(src, dest); err != nil {
			return fmt.Errorf("failed to copy %s: %w", entry.Name(), err)
		}
		fmt.Printf("  Copied %s\n", entry.Name())
	}

	cleanVersion := strings.TrimPrefix(version, "v")
	goModContent := fmt.Sprintf(`module github.com/gmcore/app

go 1.21

require (
	gmcore.io/gmcore-config v%s
	gmcore.io/gmcore-crud v%s
	gmcore.io/gmcore-encryption v%s
	gmcore.io/gmcore-events v%s
	gmcore.io/gmcore-router v%s
	gmcore.io/gmcore-settings v%s
	gmcore.io/gmcore-store v%s
	gmcore.io/gmcore-uuid v%s
)

replace (
	gmcore.io/gmcore-config => ./vendor/gmcore/gmcore-config
	gmcore.io/gmcore-crud => ./vendor/gmcore/gmcore-crud
	gmcore.io/gmcore-encryption => ./vendor/gmcore/gmcore-encryption
	gmcore.io/gmcore-events => ./vendor/gmcore/gmcore-events
	gmcore.io/gmcore-router => ./vendor/gmcore/gmcore-router
	gmcore.io/gmcore-settings => ./vendor/gmcore/gmcore-settings
	gmcore.io/gmcore-store => ./vendor/gmcore/gmcore-store
	gmcore.io/gmcore-uuid => ./vendor/gmcore/gmcore-uuid
)
`, cleanVersion, cleanVersion, cleanVersion, cleanVersion, cleanVersion, cleanVersion, cleanVersion, cleanVersion)

	if err := os.WriteFile(filepath.Join(frameworkDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("failed to write go.mod: %w", err)
	}

	tarballPath := filepath.Join(tmpDir, "gmcore-framework-"+version+".tar.gz")
	if err := createTarball(frameworkDir, tarballPath); err != nil {
		return fmt.Errorf("failed to create tarball: %w", err)
	}

	outputDir := filepath.Join(wd, "..", "dist")
	os.MkdirAll(outputDir, 0755)
	finalPath := filepath.Join(outputDir, "gmcore-framework-"+version+".tar.gz")

	if err := os.Rename(tarballPath, finalPath); err != nil {
		return fmt.Errorf("failed to move tarball: %w", err)
	}

	fmt.Printf("Framework built: %s\n", finalPath)

	return nil
}

func copyDir(src, dest string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		if rel == "." {
			return nil
		}

		destPath := filepath.Join(dest, rel)

		if info.IsDir() {
			return os.MkdirAll(destPath, info.Mode())
		}

		return copyFile(path, destPath)
	})
}

func copyFile(src, dest string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	buf := make([]byte, 32*1024)
	for {
		n, err := srcFile.Read(buf)
		if n > 0 {
			if _, werr := destFile.Write(buf[:n]); werr != nil {
				return werr
			}
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}
	}

	return nil
}

func createTarball(dir, tarballPath string) error {
	cmd := exec.Command("tar", "-czf", tarballPath, "-C", dir, ".")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("tar failed: %s", string(output))
	}
	return nil
}
