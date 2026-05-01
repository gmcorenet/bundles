package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	sdkPath    = "/home/dev/workspace/sdk"
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
			fmt.Fprintln(os.Stderr, "Usage: gmcore release <version>")
			os.Exit(1)
		}
		version := os.Args[2]
		if err := release(version); err != nil {
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
	fmt.Println("  gmcore release <version>        Create and push SDK release tag")
	fmt.Println("  gmcore build-framework [ver]    Build framework tarball locally")
	fmt.Println("  gmcore version                  Show version")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  gmcore release v1.0.0")
	fmt.Println("  gmcore build-framework v1.0.0")
}

func release(version string) error {
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}

	fmt.Printf("Creating release %s...\n", version)

	tagCmd := exec.Command("git", "tag", version)
	tagCmd.Dir = sdkPath
	if output, err := tagCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create tag: %s", string(output))
	}

	pushCmd := exec.Command("git", "push", "origin", version)
	pushCmd.Dir = sdkPath
	if output, err := pushCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to push tag: %s", string(output))
	}

	fmt.Printf("Release %s pushed! GitHub Actions will build and publish.\n", version)
	fmt.Printf("Watch progress at: https://github.com/gmcorenet/sdk/actions\n")

	return nil
}

func buildFramework(version string) error {
	fmt.Printf("Building framework %s locally...\n", version)

	absSdkPath, err := filepath.Abs(sdkPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
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

	sdkEntries, err := os.ReadDir(absSdkPath)
	if err != nil {
		return fmt.Errorf("failed to read sdk dir: %w", err)
	}

	for _, entry := range sdkEntries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "gmcore-") {
			continue
		}

		src := filepath.Join(absSdkPath, entry.Name())
		dest := filepath.Join(frameworkDir, "vendor/gmcore", entry.Name())

		if err := copyDir(src, dest); err != nil {
			return fmt.Errorf("failed to copy %s: %w", entry.Name(), err)
		}
		fmt.Printf("  Copied %s\n", entry.Name())
	}

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
`, version, version, version, version, version, version, version, version)

	if err := os.WriteFile(filepath.Join(frameworkDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("failed to write go.mod: %w", err)
	}

	fmt.Println("Running go mod tidy...")
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = frameworkDir
	if output, err := tidyCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("go mod tidy failed: %s", string(output))
	}

	tarballPath := filepath.Join(tmpDir, "gmcore-framework-"+version+".tar.gz")
	if err := createTarball(frameworkDir, tarballPath); err != nil {
		return fmt.Errorf("failed to create tarball: %w", err)
	}

	outputDir := "/home/dev/workspace/dist"
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
