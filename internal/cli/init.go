package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"yuki/internal/logger"
	"yuki/internal/manifest"
)

func InitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Yuki project",
		Long:  "Create a new Yuki project with manifest file, build script, and source files",
		RunE:  runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	if manifest.Exists(cwd) {
		logger.Error("Project already initialized (yuki.toml exists)")
		return fmt.Errorf("project already initialized")
	}

	logger.Info("Initializing new Yuki project...")

	projectInfo, err := promptProjectInfo(cwd)
	if err != nil {
		return err
	}

	m := &manifest.Manifest{
		Package: *projectInfo,
		Features: map[string][]string{
			"default": {},
		},
		Scripts: map[string]string{
			"test":   "zig test src/main.zig",
			"format": "zig fmt src/",
		},
	}

	if err := m.Save(cwd); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	if err := createProjectStructure(cwd, projectInfo.Name); err != nil {
		return fmt.Errorf("failed to create project structure: %w", err)
	}

	logger.Success("Successfully initialized project '%s'", projectInfo.Name)
	logger.Info("Next steps:")
	logger.Info("  - Edit src/main.zig to implement your project")
	logger.Info("  - Add dependencies with: yuki add <package>@<version>")
	logger.Info("  - Build your project with: yuki build")
	
	return nil
}

func promptProjectInfo(cwd string) (*manifest.PackageInfo, error) {
	reader := bufio.NewReader(os.Stdin)

	defaultName := filepath.Base(cwd)
	
	fmt.Printf("Package name (%s): ", defaultName)
	name, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = defaultName
	}

	fmt.Print("Version (0.1.0): ")
	version, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	version = strings.TrimSpace(version)
	if version == "" {
		version = "0.1.0"
	}

	fmt.Print("Description: ")
	description, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	description = strings.TrimSpace(description)

	fmt.Print("Author email: ")
	author, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	author = strings.TrimSpace(author)

	fmt.Print("License (MIT): ")
	license, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	license = strings.TrimSpace(license)
	if license == "" {
		license = "MIT"
	}

	fmt.Print("GitHub username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	username = strings.TrimSpace(username)

	fmt.Print("Minimum Zig version (0.12.0): ")
	zigVersion, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	zigVersion = strings.TrimSpace(zigVersion)
	if zigVersion == "" {
		zigVersion = "0.12.0"
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
		Edition:     "2024",
		Authors:     authors,
		License:     license,
		Description: description,
		Homepage:    homepage,
		Repository:  repository,
		ZigVersion:  zigVersion,
	}, nil
}

func createProjectStructure(projectRoot, projectName string) error {
	// Create src directory
	srcDir := filepath.Join(projectRoot, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return err
	}

	// Create main.zig
	mainZigContent := fmt.Sprintf(`const std = @import("std");

pub fn main() !void {
    std.log.info("Hello from %s!", .{});
}

test "basic test" {
    try std.testing.expectEqual(2 + 2, 4);
}
`, projectName)

	mainZigPath := filepath.Join(srcDir, "main.zig")
	if err := os.WriteFile(mainZigPath, []byte(mainZigContent), 0644); err != nil {
		return err
	}

	buildZigContent := fmt.Sprintf(`const std = @import("std");

pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});

    const exe = b.addExecutable(.{
        .name = "%s",
        .root_source_file = b.path("src/main.zig"),
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
        .root_source_file = b.path("src/main.zig"),
        .target = target,
        .optimize = optimize,
    });

    const run_unit_tests = b.addRunArtifact(unit_tests);
    const test_step = b.step("test", "Run unit tests");
    test_step.dependOn(&run_unit_tests.step);
}
`, projectName)

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
