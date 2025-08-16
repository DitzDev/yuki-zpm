package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"yuki/internal/logger"
	"yuki/internal/manifest"
	"yuki/internal/resolver"
)

func OutdatedCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "outdated",
		Short: "Show outdated dependencies",
		Long:  "List dependencies that have newer versions available",
		RunE:  runOutdated,
	}
}

func runOutdated(cmd *cobra.Command, args []string) error {
	cwd := "."
	
	
	lockFile, err := manifest.LoadLockFile(cwd)
	if err != nil {
		return fmt.Errorf("failed to load lock file: %w", err)
	}

	if len(lockFile.Package) == 0 {
		logger.Info("No dependencies installed")
		return nil
	}

	logger.Info("Checking for outdated dependencies...")

	resolver := resolver.New()
	updates, err := resolver.CheckForUpdates(lockFile)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if len(updates) == 0 {
		logger.Success("All dependencies are up to date!")
		return nil
	}

	logger.Info("Outdated dependencies:")
	for _, update := range updates {
		logger.Warn("  %-20s %s → %s", update.Name, update.CurrentVersion, update.LatestVersion)
	}

	logger.Info("\nRun 'yuki update' to update all dependencies")
	logger.Info("Run 'yuki add <package>@<version>' to update specific packages")

	return nil
}
