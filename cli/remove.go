package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"yuki_zpm.org/logger"
	"yuki_zpm.org/manifest"
	"yuki_zpm.org/internal/vendor"
)

func RemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <package>",
		Short: "Remove a dependency",
		Long:  "Remove a dependency from the project manifest and update build files",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			skipBuildUpdate, _ := cmd.Flags().GetBool("skip-build-update")
			return runRemove(args[0], skipBuildUpdate)
		},
	}
	
	cmd.Flags().Bool("skip-build-update", false, "Skip updating build.zig after removing dependency")
	
	return cmd
}

func runRemove(packageName string, skipBuildUpdate bool) error {
	cwd := "."
	
	m, err := manifest.Load(cwd)
	if err != nil {
		logger.Error("No yuki.toml found. Run 'yuki init' first.")
		return fmt.Errorf("manifest not found: %w", err)
	}

	logger.Info("Removing dependency '%s'", packageName)

	removed := false
	
	if _, exists := m.Dependencies[packageName]; exists {
		delete(m.Dependencies, packageName)
		removed = true
		logger.Info("Removed '%s' from dependencies", packageName)
	}
	
	if _, exists := m.DevDeps[packageName]; exists {
		delete(m.DevDeps, packageName)
		removed = true
		logger.Info("Removed '%s' from dev-dependencies", packageName)
	}
	
	if _, exists := m.BuildDeps[packageName]; exists {
		delete(m.BuildDeps, packageName)
		removed = true
		logger.Info("Removed '%s' from build-dependencies", packageName)
	}

	if !removed {
		logger.Error("Dependency '%s' not found", packageName)
		return fmt.Errorf("dependency not found")
	}

	if err := m.Save(cwd); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}

	lockFile, err := manifest.LoadLockFile(cwd)
	if err != nil {
		return fmt.Errorf("failed to load lock file: %w", err)
	}

	var updatedPackages []manifest.LockedPackage
	for _, pkg := range lockFile.Package {
		if pkg.Name != packageName {
			updatedPackages = append(updatedPackages, pkg)
		}
	}
	lockFile.Package = updatedPackages

	if err := lockFile.Save(cwd); err != nil {
		return fmt.Errorf("failed to save lock file: %w", err)
	}

	vendorer := vendor.New()
	
	if err := vendorer.RemovePackageFiles(cwd, packageName); err != nil {
		logger.Warn("Failed to remove package files: %v", err)
	} else {
		logger.Info("Removed package files for '%s'", packageName)
	}

	if err := vendorer.GenerateYukiZig(cwd, lockFile, m); err != nil {
		logger.Warn("Failed to regenerate yuki.zig: %v", err)
	}
	
	if !skipBuildUpdate {
		logger.Info("Updating build.zig...")
		if err := vendorer.UpdateBuildZig(cwd, lockFile, m); err != nil {
			logger.Warn("Failed to update build.zig: %v", err)
			logger.Info("You may need to manually remove the dependency from your build.zig")
		} else {
			logger.Success("Successfully updated build.zig")
		}
	}

	logger.Success("Successfully removed dependency '%s'", packageName)
	
	if !skipBuildUpdate {
		logger.Info("build.zig has been updated to remove the dependency")
	} else {
		logger.Info("Run 'yuki install' to update the dependency tree and build files")
	}

	return nil
}