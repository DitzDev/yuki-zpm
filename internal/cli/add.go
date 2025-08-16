package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"yuki/internal/fetch"
	"yuki/internal/logger"
	"yuki/internal/manifest"
	"yuki/internal/resolver"
)

func AddCmd() *cobra.Command {
	var dev bool
	var build bool
	var alias string
	var rootFile string
	var branch string

	cmd := &cobra.Command{
		Use:   "add <package>[@version]",
		Short: "Add a dependency to the project",
		Long:  "Add a dependency to the project manifest after validation. Use 'yuki install' to install the dependency.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdd(args[0], dev, build, alias, rootFile, branch)
		},
	}

	cmd.Flags().BoolVar(&dev, "dev", false, "Add as development dependency")
	cmd.Flags().BoolVar(&build, "build", false, "Add as build dependency")
	cmd.Flags().StringVar(&alias, "as", "", "Alias name for the dependency")
	cmd.Flags().StringVar(&rootFile, "root_file", "", "Root file path for the dependency (e.g. clap.zig, src/clap.zig)")
	cmd.Flags().StringVar(&branch, "branch", "", "Specific branch to use for the dependency")

	return cmd
}

func runAdd(packageSpec string, dev, build bool, alias, rootFile, branch string) error {
	cwd := "."

	m, err := manifest.Load(cwd)
	if err != nil {
		logger.Error("No yuki.toml found. Run 'yuki init' first.")
		return fmt.Errorf("manifest not found: %w", err)
	}

	packageName, version, gitURL, err := parsePackageSpec(packageSpec)
	if err != nil {
		return fmt.Errorf("invalid package specification: %w", err)
	}

	dependencyName := packageName
	if alias != "" {
		dependencyName = alias
		logger.Info("Using alias '%s' for package '%s'", alias, packageName)
	}

	logger.Info("Validating dependency '%s'...", dependencyName)

	dep := manifest.Dependency{
		Git:     gitURL,
		Version: version,
	}

	// Set branch if specified
	if branch != "" {
		dep.Branch = branch
		logger.Info("Using branch: %s", branch)
		// Clear version when branch is specified to avoid conflicts
		if version != "" {
			logger.Warn("Version '%s' specified with --branch flag. Branch takes precedence.", version)
			dep.Version = ""
		}
	}

	if rootFile != "" {
		dep.RootFile = rootFile
		logger.Info("Using root file: %s", rootFile)
	} else {
		if m.Package.RootFile != "" {
			dep.RootFile = m.Package.RootFile
			logger.Info("Using root file from yuki.toml: %s", dep.RootFile)
		} else {
			logger.Warn("No root file specified. You may need to specify --root_file flag or set root_file in [package] section of yuki.toml")
		}
	}

	logger.Info("Checking if dependency can be resolved...")
	resolverInstance := resolver.New()

	tempManifest := *m
	if tempManifest.Dependencies == nil {
		tempManifest.Dependencies = make(map[string]manifest.Dependency)
	}
	tempManifest.Dependencies[dependencyName] = dep

	_, err = resolverInstance.Resolve(&tempManifest)
	if err != nil {
		logger.Error("Failed to resolve dependency '%s': %v", dependencyName, err)
		return fmt.Errorf("dependency resolution failed: %w", err)
	}
	logger.Success("✓ Dependency can be resolved")

	logger.Info("Validating dependency availability...")
	fetcher := fetch.NewFetcher()
	fetchResult, err := fetcher.FetchDependency(dependencyName, dep)
	if err != nil {
		logger.Error("Failed to fetch dependency '%s': %v", dependencyName, err)
		return fmt.Errorf("dependency fetch failed: %w", err)
	}
	logger.Success("✓ Dependency is accessible and valid")
	logger.Info("Found version: %s", fetchResult.Version)

	logger.Info("Adding dependency to manifest...")
	
	if m.Dependencies == nil {
		m.Dependencies = make(map[string]manifest.Dependency)
	}
	if m.DevDeps == nil {
		m.DevDeps = make(map[string]manifest.Dependency)
	}
	if m.BuildDeps == nil {
		m.BuildDeps = make(map[string]manifest.Dependency)
	}

	if existsInDependencies(m, dependencyName) {
		logger.Warn("Dependency '%s' already exists, updating...", dependencyName)
	}

	if dev {
		m.DevDeps[dependencyName] = dep
		logger.Success("Added '%s' as development dependency", dependencyName)
	} else if build {
		m.BuildDeps[dependencyName] = dep
		logger.Success("Added '%s' as build dependency", dependencyName)
	} else {
		m.Dependencies[dependencyName] = dep
		logger.Success("Added '%s' as dependency", dependencyName)
	}

	if err := m.Save(cwd); err != nil {
		return fmt.Errorf("failed to save manifest: %w", err)
	}
	logger.Info("📦 Dependency '%s@%s' has been added to your project!", dependencyName, fetchResult.Version)
	if alias != "" {
		logger.Info("   Original package: %s", packageName)
		logger.Info("   Alias: %s", alias)
	}
	if dep.RootFile != "" {
		logger.Info("   Root file: %s", dep.RootFile)
	}
	if dep.Branch != "" {
		logger.Info("   Branch: %s", dep.Branch)
	}
	logger.Info("")
	logger.Info("🚀 Next steps:")
	logger.Info("   Run 'yuki install' to install the dependency")
	logger.Info("   The dependency will be available in your build.zig through yuki.zig")

	return nil
}

func existsInDependencies(m *manifest.Manifest, name string) bool {
	if m.Dependencies != nil {
		if _, exists := m.Dependencies[name]; exists {
			return true
		}
	}
	if m.DevDeps != nil {
		if _, exists := m.DevDeps[name]; exists {
			return true
		}
	}
	if m.BuildDeps != nil {
		if _, exists := m.BuildDeps[name]; exists {
			return true
		}
	}
	return false
}

func parsePackageSpec(packageSpec string) (name, version, gitURL string, err error) {
	// Handle different formats:
	// - package@version
	// - username/repo@version (with version)
	// - username/repo (without version - will use latest)
	// - https://github.com/username/repo@version
	// - https://github.com/username/repo (without version - will use latest)

	parts := strings.Split(packageSpec, "@")
	if len(parts) > 2 {
		return "", "", "", fmt.Errorf("invalid package specification format")
	}

	packageURL := parts[0]
	hasVersion := len(parts) == 2
	if hasVersion {
		version = parts[1]
	}

	if !strings.Contains(packageURL, "/") {
		return "", "", "", fmt.Errorf("package must be in 'username/repo' format")
	}

	if !strings.HasPrefix(packageURL, "http") {
		gitURL = fmt.Sprintf("https://github.com/%s", packageURL)
	} else {
		gitURL = packageURL
	}

	if strings.HasPrefix(gitURL, "https://github.com/") {
		parts := strings.Split(strings.TrimPrefix(gitURL, "https://github.com/"), "/")
		if len(parts) >= 2 {
			name = parts[1]
			if strings.HasSuffix(name, ".git") {
				name = name[:len(name)-4]
			}
		}
	}

	if name == "" {
		return "", "", "", fmt.Errorf("could not determine package name")
	}

	if !hasVersion {
		version = ""
	}

	return name, version, gitURL, nil
}