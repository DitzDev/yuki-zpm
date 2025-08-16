package cli

import (
        "fmt"

        "github.com/spf13/cobra"
        "yuki/internal/logger"
        "yuki/internal/manifest"
)

func ListCmd() *cobra.Command {
        var tree bool

        cmd := &cobra.Command{
                Use:   "list",
                Short: "List dependencies",
                Long:  "List all dependencies in the project",
                RunE: func(cmd *cobra.Command, args []string) error {
                        return runList(tree)
                },
        }

        cmd.Flags().BoolVarP(&tree, "tree", "t", false, "Show dependencies in tree format")

        return cmd
}

func runList(tree bool) error {
        cwd := "."
        
        m, err := manifest.Load(cwd)
        if err != nil {
                logger.Error("No yuki.toml found. Run 'yuki init' first.")
                return fmt.Errorf("manifest not found: %w", err)
        }
        
        lockFile, err := manifest.LoadLockFile(cwd)
        if err != nil {
                logger.Warn("Lock file not found, showing manifest dependencies only")
        }

        installedVersions := make(map[string]string)
        if lockFile != nil {
                for _, pkg := range lockFile.Package {
                        installedVersions[pkg.Name] = pkg.Version
                }
        }

        logger.Info("Dependencies for %s:", m.Package.Name)

        if len(m.Dependencies) > 0 {
                fmt.Println("\nDependencies:")
                listDependencies(m.Dependencies, installedVersions, tree, "")
        }

        if len(m.DevDeps) > 0 {
                fmt.Println("\nDev Dependencies:")
                listDependencies(m.DevDeps, installedVersions, tree, "")
        }

        if len(m.BuildDeps) > 0 {
                fmt.Println("\nBuild Dependencies:")
                listDependencies(m.BuildDeps, installedVersions, tree, "")
        }

        if len(m.Dependencies) == 0 && len(m.DevDeps) == 0 && len(m.BuildDeps) == 0 {
                logger.Info("No dependencies found")
        }

        return nil
}

func listDependencies(deps map[string]manifest.Dependency, installedVersions map[string]string, tree bool, prefix string) {
        for name, dep := range deps {
                installedVersion := installedVersions[name]
                
                var versionInfo string
                if installedVersion != "" {
                        versionInfo = fmt.Sprintf(" (installed: %s)", installedVersion)
                } else {
                        versionInfo = " (not installed)"
                }
                
                if tree {
                        fmt.Printf("%s├── %s@%s%s\n", prefix, name, dep.Version, versionInfo)
                        fmt.Printf("%s│   └── %s\n", prefix, dep.Git)
                } else {
                        fmt.Printf("  %s@%s%s\n", name, dep.Version, versionInfo)
                        fmt.Printf("    Source: %s\n", dep.Git)
                }
        }
}
