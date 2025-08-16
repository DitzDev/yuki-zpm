package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"yuki/internal/logger"
	"yuki/internal/manifest"
	"yuki/internal/resolver"
)

func CheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Verify dependencies and manifest integrity",
		Long:  "Check that all dependencies can be resolved and manifest is valid",
		RunE:  runCheck,
	}
}

func runCheck(cmd *cobra.Command, args []string) error {
	cwd := "."
	
	logger.Info("Checking manifest and dependencies...")

	m, err := manifest.Load(cwd)
	if err != nil {
		logger.Error("No yuki.toml found. Run 'yuki init' first.")
		return fmt.Errorf("manifest not found: %w", err)
	}

	if err := m.Validate(); err != nil {
		logger.Error("Manifest validation failed: %v", err)
		return err
	}
	logger.Success("Manifest is valid")

	resolver := resolver.New()
	if err := resolver.ValidateDependencies(m); err != nil {
		logger.Error("Dependency validation failed: %v", err)
		return err
	}
	logger.Success("All dependencies are accessible")

	resolution, err := resolver.Resolve(m)
	if err != nil {
		logger.Error("Dependency resolution failed: %v", err)
		return err
	}
	logger.Success("All dependencies can be resolved")

	logger.Info("Found %d dependencies:", len(resolution.Dependencies))
	for _, dep := range resolution.Dependencies {
		logger.Info("  - %s@%s", dep.Name, dep.Version)
	}

	logger.Success("All checks passed")
	return nil
}
