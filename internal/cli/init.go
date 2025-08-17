package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
	"yuki/internal/logger"
	"yuki/internal/manifest"
)

func InitCmd() *cobra.Command {
	var autoYes bool
	
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Yuki project",
		Long:  "Create a new Yuki project with manifest file, build script, and source files",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, args, autoYes)
		},
	}

	cmd.Flags().BoolVarP(&autoYes, "yes", "y", false, "Skip prompts and use default values")
	
	return cmd
}

func runInit(cmd *cobra.Command, args []string, autoYes bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if manifest.Exists(cwd) {
		logger.Error("Project already initialized (yuki.toml exists)")
		return fmt.Errorf("project already initialized")
	}

	logger.Info("Initializing new Yuki project...")

	projectInfo, err := promptProjectInfo(cwd, autoYes)
	if err != nil {
		return err
	}

	scripts := map[string]string{
		"test":   fmt.Sprintf("zig test %s", projectInfo.RootFile),
		"format": "zig fmt src/",
	}

	m := &manifest.Manifest{
		Package: *projectInfo,
		Features: map[string][]string{
			"default": {},
		},
		Scripts: scripts,
	}

	if err := m.Save(cwd); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	if err := createProjectStructure(cwd, projectInfo.Name, projectInfo.RootFile); err != nil {
		return fmt.Errorf("failed to create project structure: %w", err)
	}

	logger.Success("Successfully initialized project '%s'", projectInfo.Name)
	logger.Info("Next steps:")
	logger.Info("  - Edit %s to implement your project", projectInfo.RootFile)
	logger.Info("  - Add dependencies with: yuki add <package>@<version>")
	logger.Info("  - Build your project with: yuki build")
	
	return nil
}

func promptProjectInfo(cwd string, autoYes bool) (*manifest.PackageInfo, error) {
	reader := bufio.NewReader(os.Stdin)
	defaultName := filepath.Base(cwd)
	
	detectedZigVersion := detectZigVersion()
	defaultZigVersion := "0.12.0"
	if detectedZigVersion != "" {
		defaultZigVersion = detectedZigVersion
	}

	var name, version, description, author, license, username, zigVersion, rootFile string

	if autoYes {
		name = defaultName
		version = "0.1.0"
		description = ""
		author = ""
		license = "MIT"
		username = ""
		zigVersion = defaultZigVersion
		rootFile = "src/main.zig"
		
		logger.Info("Using auto-init mode with defaults:")
		logger.Info("  Package name: %s", name)
		logger.Info("  Version: %s", version)
		logger.Info("  License: %s", license)
		logger.Info("  Zig version: %s", zigVersion)
		logger.Info("  Root file: %s", rootFile)
	} else {
		fmt.Printf("Package name (%s): ", defaultName)
		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		name = strings.TrimSpace(input)
		if name == "" {
			name = defaultName
		}

		fmt.Print("Version (0.1.0): ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		version = strings.TrimSpace(input)
		if version == "" {
			version = "0.1.0"
		}

		fmt.Print("Description: ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		description = strings.TrimSpace(input)

		fmt.Print("Author (Your name <youremail@example.com>): ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		author = strings.TrimSpace(input)

		fmt.Print("License (MIT): ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		license = strings.TrimSpace(input)
		if license == "" {
			license = "MIT"
		}

		fmt.Print("GitHub username: ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		username = strings.TrimSpace(input)

		if detectedZigVersion != "" {
			fmt.Printf("Minimum Zig version (detected: %s): ", detectedZigVersion)
		} else {
			fmt.Printf("Minimum Zig version (%s): ", defaultZigVersion)
		}
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		zigVersion = strings.TrimSpace(input)
		if zigVersion == "" {
			zigVersion = defaultZigVersion
		}

		fmt.Print("Root source file (src/main.zig): ")
		input, err = reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		rootFile = strings.TrimSpace(input)
		if rootFile == "" {
			rootFile = "src/main.zig"
		}

		rootFile, err = validateAndSanitizeRootFile(rootFile)
		if err != nil {
			return nil, fmt.Errorf("invalid root file: %w", err)
		}
	}

	var authors []string
	if author != "" {
		authors = []string{author}
	}

	var homepage, repository string
	if username != "" {
		homepage = fmt.Sprintf("https://github.com/%s/%s", username, name)
		repository = fmt.Sprintf("https://github.com/%s/%s", username, name)
	}

	return &manifest.PackageInfo{
		Name:        name,
		Version:     version,
		Authors:     authors,
		License:     license,
		Description: description,
		Homepage:    homepage,
		Repository:  repository,
		ZigVersion:  zigVersion,
		RootFile:    rootFile,
	}, nil
}

func detectZigVersion() string {
	cmd := exec.Command("zig", "version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	version := strings.TrimSpace(string(output))
	
	// Handle different Zig version formats
	// Examples: "0.12.0", "0.13.0-dev.123+abc123", "0.11.0"
	
	// Extract base version (major.minor.patch)
	re := regexp.MustCompile(`^(\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(version)
	if len(matches) > 1 {
		return matches[1]
	}

	devRe := regexp.MustCompile(`^(\d+)\.(\d+)\.\d+-dev`)
	devMatches := devRe.FindStringSubmatch(version)
	if len(devMatches) > 2 {
		return fmt.Sprintf("%s.%s.0", devMatches[1], devMatches[2])
	}
	return version
}

func validateAndSanitizeRootFile(rootFile string) (string, error) {
	cleaned := filepath.Clean(rootFile)

	if strings.Contains(cleaned, "..") {
		return "", fmt.Errorf("root file cannot reference parent directories")
	}
	
	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("root file must be a relative path")
	}

	if !strings.HasSuffix(cleaned, ".zig") {
		return "", fmt.Errorf("root file must be a .zig file")
	}

	cleaned = strings.ReplaceAll(cleaned, "\\", "/")
	
	return cleaned, nil
}

func createProjectStructure(projectRoot, projectName, rootFile string) error {
	rootFileDir := filepath.Dir(rootFile)
	if rootFileDir != "." {
		fullDir := filepath.Join(projectRoot, rootFileDir)
		if err := os.MkdirAll(fullDir, 0755); err != nil {
			return err
		}
	}

	mainZigContent := fmt.Sprintf(`const std = @import("std");

pub fn main() !void {
    std.log.info("Hello from %s!", .{"%s"});
}

test "basic test" {
    try std.testing.expectEqual(2 + 2, 4);
}
`, projectName, projectName)

	mainZigPath := filepath.Join(projectRoot, rootFile)
	if err := os.WriteFile(mainZigPath, []byte(mainZigContent), 0644); err != nil {
		return err
	}

	buildZigContent := fmt.Sprintf(`const std = @import("std");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const exe = b.addExecutable(.{
        .name = "%s",
        .root_source_file = b.path("%s"),
        .target = target,
        .optimize = optimize,
    });

    b.installArtifact(exe);

    const run_cmd = b.addRunArtifact(exe);
    run_cmd.step.dependOn(b.getInstallStep());

    if (b.args) |args| {
        run_cmd.addArgs(args);
    }

    const run_step = b.step("run", "Run the app");
    run_step.dependOn(&run_cmd.step);

    const unit_tests = b.addTest(.{
        .root_source_file = b.path("%s"),
        .target = target,
        .optimize = optimize,
    });

    const run_unit_tests = b.addRunArtifact(unit_tests);
    const test_step = b.step("test", "Run unit tests");
    test_step.dependOn(&run_unit_tests.step);
}
`, projectName, rootFile, rootFile)

	buildZigPath := filepath.Join(projectRoot, "build.zig")
	if err := os.WriteFile(buildZigPath, []byte(buildZigContent), 0644); err != nil {
		return err
	}

	yukiZigContent := `// Auto-generated file by Yuki package manager
// Do not edit this file directly

// No dependencies
`

	yukiZigPath := filepath.Join(projectRoot, "yuki.zig")
	if err := os.WriteFile(yukiZigPath, []byte(yukiZigContent), 0644); err != nil {
		return err
	}

	return nil
}