package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"yuki_zpm.org/logger"
	"yuki_zpm.org/manifest"
	"yuki_zpm.org/resolver"
)

func UpdateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "update [package]",
		Short: "Update dependencies",
		Long:  "Update dependencies to their latest compatible versions",
		RunE:  runUpdate,
	}
}

func runUpdate(cmd *cobra.Command, args []string) error {
	cwd := "."

	
	_, err := manifest.Load(cwd)
	if err != nil {
		logger.Error("No yuki.toml found. Run 'yuki init' first.")
		return fmt.Errorf("manifest not found: %w", err)
	}

	
	lockFile, err := manifest.LoadLockFile(cwd)
	if err != nil {
		return fmt.Errorf("failed to load lock file: %w", err)
	}

	logger.Info("Checking for dependency updates...")

	res := resolver.New()
	updates, err := res.CheckForUpdates(lockFile)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if len(updates) == 0 {
		logger.Success("All dependencies are up to date")
		return nil
	}

	
	if len(args) > 0 {
		packageName := args[0]
		var filteredUpdates []resolver.UpdateInfo
		for _, update := range updates {
			if update.Name == packageName {
				filteredUpdates = append(filteredUpdates, update)
			}
		}
		if len(filteredUpdates) == 0 {
			logger.Info("Package '%s' is already up to date", packageName)
			return nil
		}
		updates = filteredUpdates
	}

	logger.Info("Available updates:")
	for _, update := range updates {
		logger.Info("  %s: %s → %s", update.Name, update.CurrentVersion, update.LatestVersion)
	}

	
	
	logger.Info("Run 'yuki add %s@%s' to update specific packages", updates[0].Name, updates[0].LatestVersion)

	return nil
}