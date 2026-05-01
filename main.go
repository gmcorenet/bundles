package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

const (
	sdkRepo    = "gmcorenet/sdk"
	cliVersion = "0.1.0"
)

func main() {
	token, err := getGitHubToken()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

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
		if err := release(bumpType, token); err != nil {
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

	case "self-update":
		targetVersion := ""
		if len(os.Args) >= 3 {
			targetVersion = os.Args[2]
		}
		if err := selfUpdate(targetVersion, token); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

	case "version", "--version", "-v":
		fmt.Printf("gmcore workspace tool %s\n", cliVersion)

	case "uninstall":
		if err := uninstall(); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

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
	fmt.Println("  gmcore self-update [version]  Update tool to latest or specific version")
	fmt.Println("  gmcore version                Show version")
	fmt.Println("  gmcore uninstall              Uninstall workspace tool")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  gmcore release minor")
	fmt.Println("  gmcore release v1.0.0")
	fmt.Println("  gmcore self-update")
	fmt.Println("  gmcore self-update 0.4.0")
}

func uninstall() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to find current executable: %w", err)
	}

	var targetPath string
	switch runtime.GOOS {
	case "linux", "darwin":
		targetPath = "/usr/local/bin/gmcore"
	case "windows":
		targetPath = "C:\\Program Files\\gmcore\\gmcore.exe"
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	if exePath != targetPath {
		return fmt.Errorf("uninstall only works when running the installed binary")
	}

	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove binary: %w", err)
	}

	fmt.Printf("Uninstalled gmcore from %s\n", targetPath)
	return nil
}

func getGitHubToken() (string, error) {
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("GITHUB_TOKEN not set and could not determine home directory")
	}

	configPath := filepath.Join(usr.HomeDir, ".gmcore", "config")

	file, err := os.Open(configPath)
	if err != nil {
		return "", fmt.Errorf(`GITHUB_TOKEN not configured

Create config file at: %s

With content:
  token = YOUR_GITHUB_TOKEN

To create a token:
1. Go to https://github.com/settings/tokens
2. Generate new token (classic)
3. Select "repo" scope
4. Copy the token and add it to ~/.gmcore/config`, configPath)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "token") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1]), nil
			}
		}
	}

	return "", fmt.Errorf(`token not found in %s

Add to config:
  token = YOUR_GITHUB_TOKEN`, configPath)
}

func selfUpdate(targetVersion, token string) error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to find executable: %w", err)
	}

	platform := getPlatform()
	arch := getArch()
	binaryName := fmt.Sprintf("gmcore-%s-%s", platform, arch)

	if targetVersion == "" {
		fmt.Println("Checking for latest version...")
		latest, err := getLatestVersion(token)
		if err != nil {
			return fmt.Errorf("failed to get latest version: %w", err)
		}
		targetVersion = latest
		fmt.Printf("Latest version: %s\n", targetVersion)
	}

	currentVersion := cliVersion
	if targetVersion == currentVersion {
		fmt.Printf("Already at version %s\n", currentVersion)
		return nil
	}

	fmt.Printf("Updating from %s to %s...\n", currentVersion, targetVersion)

	downloadURL := fmt.Sprintf("https://github.com/gmcorenet/workspace/releases/download/%s/%s", targetVersion, binaryName)

	req, err := http.NewRequest("HEAD", downloadURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "token "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to check version: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode == 404 {
		return fmt.Errorf("version %s not found. Run 'gmcore release --help' to see available versions.", targetVersion)
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	tmpDir, err := os.MkdirTemp("", "gmcore-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	req, _ = http.NewRequest("GET", downloadURL, nil)
	req.Header.Set("Authorization", "token "+token)

	resp, err = client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download: status %d", resp.StatusCode)
	}

	tmpBinary := filepath.Join(tmpDir, "gmcore")
	out, err := os.Create(tmpBinary)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		return fmt.Errorf("failed to write: %w", err)
	}
	out.Close()

	if err := os.Chmod(tmpBinary, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	if err := os.Rename(tmpBinary, exePath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	fmt.Printf("Updated to %s successfully\n", targetVersion)
	return nil
}

func getLatestVersion(token string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/gmcorenet/workspace/releases/latest")

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "token "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := decodeJSON(resp.Body, &release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

func decodeJSON(r io.Reader, v interface{}) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	s := string(data)
	if idx := strings.Index(s, "\"tag_name\""); idx != -1 {
		start := strings.Index(s[idx:], "\"") + idx + 1
		end := start + strings.Index(s[start:], "\"")
		if tag := strings.TrimSpace(s[start:end]); tag != "" {
			if m, ok := v.(*struct{ TagName string }); ok {
				m.TagName = tag
			}
		}
	}
	return nil
}

func getPlatform() string {
	switch strings.ToLower(os.Getenv("GOOS")) {
	case "windows":
		return "windows"
	case "darwin":
		return "darwin"
	default:
		return "linux"
	}
}

func getArch() string {
	switch strings.ToLower(os.Getenv("GOARCH")) {
	case "arm64":
		return "arm64"
	default:
		return "amd64"
	}
}

func release(bumpOrVersion, token string) error {
	var newVersion string

	if strings.HasPrefix(bumpOrVersion, "v") || strings.Contains(bumpOrVersion, ".") {
		if !strings.HasPrefix(bumpOrVersion, "v") {
			bumpOrVersion = "v" + bumpOrVersion
		}
		newVersion = bumpOrVersion
		fmt.Printf("Using explicit version: %s\n", newVersion)
	} else {
		current, err := getLatestVersionFromTags(token)
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

	sdkPath, err := getSdkPath(token)
	if err != nil {
		return err
	}

	fmt.Printf("Creating release %s...\n", newVersion)

	tagCmd := exec.Command("git", "tag", newVersion)
	tagCmd.Dir = sdkPath
	if output, err := tagCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create tag: %s", string(output))
	}

	pushURL := fmt.Sprintf("https://%s@github.com/%s.git", token, sdkRepo)
	pushCmd := exec.Command("git", "push", pushURL, newVersion)
	pushCmd.Dir = sdkPath
	if output, err := pushCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to push tag: %s", string(output))
	}

	fmt.Printf("Release %s pushed! GitHub Actions will build and publish.\n", newVersion)
	fmt.Printf("Watch progress at: https://github.com/%s/actions\n", sdkRepo)

	return nil
}

func getLatestVersionFromTags(token string) (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", sdkRepo)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "token "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
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

	if resp.StatusCode == 404 {
		tagsURL := fmt.Sprintf("https://api.github.com/repos/%s/tags", sdkRepo)
		req, _ := http.NewRequest("GET", tagsURL, nil)
		req.Header.Set("Authorization", "token "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "v0.0.0", nil
		}
		defer resp.Body.Close()

		if resp.StatusCode == 200 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return "v0.0.0", nil
			}

			var tags []struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal(body, &tags); err != nil || len(tags) == 0 {
				return "v0.0.0", nil
			}
			return tags[0].Name, nil
		}
	}

	return "v0.0.0", nil
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

func getSdkPath(token string) (string, error) {
	if sdkPath := os.Getenv("GMCORE_SDK_PATH"); sdkPath != "" {
		if _, err := os.Stat(sdkPath); err != nil {
			return "", fmt.Errorf("sdk path from GMCORE_SDK_PATH not found: %s", sdkPath)
		}
		configureRemote(sdkPath, token)
		return sdkPath, nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	sdkPath := filepath.Join(wd, "sdk")
	if _, err := os.Stat(sdkPath); err == nil {
		configureRemote(sdkPath, token)
		return sdkPath, nil
	}

	sdkPath = filepath.Join(wd, "..", "sdk")
	if _, err := os.Stat(sdkPath); err != nil {
		return "", fmt.Errorf("sdk path not found. Set GMCORE_SDK_PATH or run from workspace with sibling sdk directory")
	}
	configureRemote(sdkPath, token)
	return sdkPath, nil
}

func configureRemote(sdkPath, token string) {
	remoteURL := fmt.Sprintf("https://%s@github.com/%s.git", token, sdkRepo)
	cmd := exec.Command("git", "remote", "set-url", "origin", remoteURL)
	cmd.Dir = sdkPath
	cmd.Run()
}

func buildFramework(version string) error {
	fmt.Printf("Building framework %s locally...\n", version)

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get wd: %w", err)
	}

	sdkPath := filepath.Join(wd, "sdk")
	if _, err := os.Stat(sdkPath); os.IsNotExist(err) {
		sdkPath = filepath.Join(wd, "..", "sdk")
		if _, err := os.Stat(sdkPath); os.IsNotExist(err) {
			return fmt.Errorf("sdk path not found at: %s or %s", filepath.Join(wd, "sdk"), sdkPath)
		}
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
	gmcore.net/gmcore-config v%s
	gmcore.net/gmcore-crud v%s
	gmcore.net/gmcore-encryption v%s
	gmcore.net/gmcore-events v%s
	gmcore.net/gmcore-router v%s
	gmcore.net/gmcore-settings v%s
	gmcore.net/gmcore-store v%s
	gmcore.net/gmcore-uuid v%s
)

replace (
	gmcore.net/gmcore-config => ./vendor/gmcore/gmcore-config
	gmcore.net/gmcore-crud => ./vendor/gmcore/gmcore-crud
	gmcore.net/gmcore-encryption => ./vendor/gmcore/gmcore-encryption
	gmcore.net/gmcore-events => ./vendor/gmcore/gmcore-events
	gmcore.net/gmcore-router => ./vendor/gmcore/gmcore-router
	gmcore.net/gmcore-settings => ./vendor/gmcore/gmcore-settings
	gmcore.net/gmcore-store => ./vendor/gmcore/gmcore-store
	gmcore.net/gmcore-uuid => ./vendor/gmcore/gmcore-uuid
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
