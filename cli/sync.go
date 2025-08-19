package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"yuki_zpm.org/logger"
	"yuki_zpm.org/manifest"
	"yuki_zpm.org/resolver"
)

func SyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Synchronize dependencies",
		Long:  "Ensure dependencies are consistent and correct",
		RunE:  runSync,
	}
}

func runSync(cmd *cobra.Command, args []string) error {
	cwd := "."
	
	logger.Info("Synchronizing dependencies...")

	
	m, err := manifest.Load(cwd)
	if err != nil {
		logger.Error("No yuki.toml found. Run 'yuki init' first.")
		return fmt.Errorf("manifest not found: %w", err)
	}

	
	lockFile, err := manifest.LoadLockFile(cwd)
	if err != nil {
		return fmt.Errorf("failed to load lock file: %w", err)
	}

	
	manifestDeps := m.GetAllDependencies()
	
	
	inconsistencies := []string{}
	
	
	lockPkgMap := make(map[string]bool)
	for _, pkg := range lockFile.Package {
		lockPkgMap[pkg.Name] = true
	}
	
	for name := range manifestDeps {
		if !lockPkgMap[name] {
			inconsistencies = append(inconsistencies, fmt.Sprintf("'%s' is in manifest but not in lock file", name))
		}
	}
	
	
	for _, pkg := range lockFile.Package {
		if _, exists := manifestDeps[pkg.Name]; !exists {
			inconsistencies = append(inconsistencies, fmt.Sprintf("'%s' is in lock file but not in manifest", pkg.Name))
		}
	}

	if len(inconsistencies) > 0 {
		logger.Warn("Found inconsistencies:")
		for _, issue := range inconsistencies {
			logger.Warn("  - %s", issue)
		}
		logger.Info("Run 'yuki install' to resolve inconsistencies")
		return fmt.Errorf("dependencies are not synchronized")
	}

	
	resolver := resolver.New()
	if err := resolver.ValidateDependencies(m); err != nil {
		logger.Error("Dependency validation failed: %v", err)
		return err
	}

	logger.Success("Dependencies are synchronized and valid")
	return nil
}
