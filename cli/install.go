package cli

import (
	"fmt"
    "os/exec"
    "os"
    
	"github.com/spf13/cobra"
	"yuki_zpm.org/fetch"
	"yuki_zpm.org/logger"
	"yuki_zpm.org/manifest"
	"yuki_zpm.org/resolver"
	"yuki_zpm.org/internal/vendor"
)

func InstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install all dependencies",
		Long:  "Install all dependencies according to the manifest and lock file",
		RunE:  runInstall,
	}
	
	cmd.Flags().Bool("skip-build-update", false, "Skip updating build.zig with dependencies")
	
	return cmd
}

func runInstall(cmd *cobra.Command, args []string) error {
	cwd := "."
	skipBuildUpdate, _ := cmd.Flags().GetBool("skip-build-update")

	m, err := manifest.Load(cwd)
	if err != nil {
		logger.Error("No yuki.toml found. Run 'yuki init' first.")
		return fmt.Errorf("manifest not found: %w", err)
	}

	logger.Info("Installing dependencies...")

	resolver := resolver.New()
	resolution, err := resolver.Resolve(m)
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	if len(resolution.Dependencies) == 0 {
		logger.Info("No dependencies to install")
		return nil
	}

	fetcher := fetch.NewFetcher()
	vendorer := vendor.New()

	lockFile := &manifest.LockFile{
		Metadata: manifest.LockMetadata{Version: "1"},
		Package:  []manifest.LockedPackage{},
	}

	for _, dep := range resolution.Dependencies {
		logger.Info("Installing '%s'", dep.Name)
		
		var manifestDep manifest.Dependency
		found := false
		
		if d, exists := m.Dependencies[dep.Name]; exists {
			manifestDep = d
			found = true
		} else if d, exists := m.DevDeps[dep.Name]; exists {
			manifestDep = d
			found = true
		} else if d, exists := m.BuildDeps[dep.Name]; exists {
			manifestDep = d
			found = true
		}
		
		if !found {
			logger.Warn("Dependency '%s' not found in manifest, skipping", dep.Name)
			continue
		}

		result, err := fetcher.FetchDependency(dep.Name, manifestDep)
		if err != nil {
			return fmt.Errorf("failed to fetch dependency '%s': %w", dep.Name, err)
		}

		if err := vendorer.VendorDependency(dep.Name, result.Path, cwd); err != nil {
			return fmt.Errorf("failed to vendor dependency '%s': %w", dep.Name, err)
		}

		lockFile.Package = append(lockFile.Package, manifest.LockedPackage{
			Name:     dep.Name,
			Version:  result.Version,
			Source:   manifestDep.Git,
			Checksum: result.Checksum,
		})

		logger.Success("Installed '%s@%s'", dep.Name, result.Version)
	}
	
	if err := lockFile.Save(cwd); err != nil {
		return fmt.Errorf("failed to save lock file: %w", err)
	}

	if err := vendorer.GenerateYukiZig(cwd, lockFile, m); err != nil {
		return fmt.Errorf("failed to generate yuki.zig: %w", err)
	}

	if !skipBuildUpdate {
		logger.Info("Updating build.zig with dependencies...")
		if err := vendorer.UpdateBuildZig(cwd, lockFile, m); err != nil {
			logger.Warn("Failed to update build.zig: %v", err)
			logger.Info("You may need to manually add dependencies to your build.zig")
		} else {
			logger.Success("Successfully updated build.zig")
		}
	}
	
	logger.Info("Formatting build.zig...")
    cmdFmt := exec.Command("zig", "fmt", "build.zig")
    cmdFmt.Stdout = os.Stdout
    cmdFmt.Stderr = os.Stderr
    if err := cmdFmt.Run(); err != nil {
       logger.Warn("Failed to format build.zig: %v", err)
    } else {
       logger.Success("Finished formatting!")
    }

	logger.Success("Successfully installed %d dependencies", len(resolution.Dependencies))
	if !skipBuildUpdate {
		logger.Info("Dependencies have been automatically added to build.zig")
	}
	
	return nil
}