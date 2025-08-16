package build

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"yuki/internal/logger"
	"yuki/internal/manifest"
)

type Builder struct{}

func New() *Builder {
	return &Builder{}
}

func (b *Builder) Build(projectRoot string, release bool) error {
	if err := b.ensureBuildZig(projectRoot); err != nil {
		return err
	}
	
	logger.Info("Building project...")
	
	args := []string{"build"}
	if release {
		args = append(args, "--release")
	}
	
	cmd := exec.Command("zig", args...)
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	
	logger.Success("Build completed successfully")
	return nil
}

func (b *Builder) Test(projectRoot string) error {
	if err := b.ensureBuildZig(projectRoot); err != nil {
		return err
	}
	
	logger.Info("Running tests...")
	
	cmd := exec.Command("zig", "build", "test")
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}
	
	logger.Success("All tests passed")
	return nil
}

func (b *Builder) Run(projectRoot string, args []string) error {
	if err := b.ensureBuildZig(projectRoot); err != nil {
		return err
	}
	
	logger.Info("Building and running project...")
	
	cmdArgs := []string{"build", "run"}
	if len(args) > 0 {
		cmdArgs = append(cmdArgs, "--")
		cmdArgs = append(cmdArgs, args...)
	}
	
	cmd := exec.Command("zig", cmdArgs...)
	cmd.Dir = projectRoot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("run failed: %w", err)
	}
	
	return nil
}

func (b *Builder) Clean(projectRoot string) error {
	logger.Info("Cleaning build artifacts...")

	buildDirs := []string{
		"zig-cache",
		"zig-out",
		".zig-cache",
	}
	
	for _, dir := range buildDirs {
		dirPath := filepath.Join(projectRoot, dir)
		if err := os.RemoveAll(dirPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove %s: %w", dir, err)
		}
	}
	
	logger.Success("Clean completed")
	return nil
}

func (b *Builder) CheckZigInstalled() error {
	_, err := exec.LookPath("zig")
	if err != nil {
		return fmt.Errorf("zig is not installed or not available in PATH")
	}

	cmd := exec.Command("zig", "version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get zig version: %w", err)
	}
	
	version := strings.TrimSpace(string(output))
	logger.Debug("Found Zig version: %s", version)
	
	return nil
}

func (b *Builder) ensureBuildZig(projectRoot string) error {
	buildZigPath := filepath.Join(projectRoot, "build.zig")

	if _, err := os.Stat(buildZigPath); err == nil {
		return nil
	}
	
	m, err := manifest.Load(projectRoot)
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}
	
	buildZigContent := b.generateBuildZig(m)
	
	if err := os.WriteFile(buildZigPath, []byte(buildZigContent), 0644); err != nil {
		return fmt.Errorf("failed to create build.zig: %w", err)
	}
	
	logger.Debug("Generated build.zig")
	return nil
}

func (b *Builder) generateBuildZig(m *manifest.Manifest) string {
	return fmt.Sprintf(`const std = @import("std");

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
`, m.Package.Name)
}
