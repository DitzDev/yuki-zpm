package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"yuki_zpm.org/logger"
	"yuki_zpm.org/manifest"
)

func WhyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "why <package>",
		Short: "Explain why a package is needed",
		Long:  "Show the dependency chain that requires a specific package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWhy(args[0])
		},
	}
}

func runWhy(packageName string) error {
	cwd := "."
	
	
	m, err := manifest.Load(cwd)
	if err != nil {
		logger.Error("No yuki.toml found. Run 'yuki init' first.")
		return fmt.Errorf("manifest not found: %w", err)
	}

	logger.Info("Analyzing dependency tree for '%s'...", packageName)

	
	var dependencyType string
	var found bool

	if _, exists := m.Dependencies[packageName]; exists {
		dependencyType = "runtime dependency"
		found = true
	} else if _, exists := m.DevDeps[packageName]; exists {
		dependencyType = "development dependency"
		found = true
	} else if _, exists := m.BuildDeps[packageName]; exists {
		dependencyType = "build dependency"
		found = true
	}

	if found {
		logger.Success("Package '%s' is a direct %s of '%s'", packageName, dependencyType, m.Package.Name)
		return nil
	}

	
	
	logger.Warn("Package '%s' is not found in the dependency tree", packageName)
	logger.Info("This could mean:")
	logger.Info("  - The package is not used by this project")
	logger.Info("  - It's a transitive dependency (not yet fully implemented)")
	logger.Info("  - There's a typo in the package name")

	
	allDeps := m.GetAllDependencies()
	if len(allDeps) > 0 {
		logger.Info("\nAvailable dependencies:")
		for name := range allDeps {
			logger.Info("  - %s", name)
		}
	}

	return fmt.Errorf("package not found in dependency tree")
}
