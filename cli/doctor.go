package cli

import (
        "fmt"
        "net/http"
        "os/exec"
        "runtime"

        "github.com/spf13/cobra"
        "yuki_zpm.org/build"
        "yuki_zpm.org/cache"
        "yuki_zpm.org/config"
        "yuki_zpm.org/logger"
        "yuki_zpm.org/manifest"
        "yuki_zpm.org/utils"
)

func DoctorCmd() *cobra.Command {
        return &cobra.Command{
                Use:   "doctor",
                Short: "Diagnose environment issues",
                Long:  "Check the environment and diagnose potential issues",
                RunE:  runDoctor,
        }
}

func runDoctor(cmd *cobra.Command, args []string) error {
        logger.Info("Running Yuki doctor...")
        fmt.Println()

        issues := 0
        
        fmt.Printf("🖥️  Operating System: %s/%s\n", runtime.GOOS, runtime.GOARCH)
        
        fmt.Print("⚡ Zig: ")
        builder := build.New()
        if err := builder.CheckZigInstalled(); err != nil {
                fmt.Printf("❌ Not found or not working\n")
                fmt.Printf("   Error: %v\n", err)
                fmt.Printf("   Please install Zig from https://ziglang.org/download/\n")
                issues++
        } else {
                fmt.Printf("✅ Found and working\n")
                
                if cmd := exec.Command("zig", "version"); cmd != nil {
                        if output, err := cmd.Output(); err == nil {
                                fmt.Printf("   Version: %s", string(output))
                        }
                }
        }

        fmt.Print("🔧 Git: ")
        if err := utils.CheckGitAvailable(); err != nil {
                fmt.Printf("❌ Not found\n")
                fmt.Printf("   Error: %v\n", err)
                fmt.Printf("   Git is required to fetch dependencies\n")
                issues++
        } else {
                fmt.Printf("✅ Found\n")
                
                if cmd := exec.Command("git", "--version"); cmd != nil {
                        if output, err := cmd.Output(); err == nil {
                                fmt.Printf("   %s", string(output))
                        }
                }
        }

        fmt.Print("📁 Current Project: ")
        cwd := "."
        if manifest.Exists(cwd) {
                fmt.Printf("✅ Valid Yuki project\n")
                
                if m, err := manifest.Load(cwd); err != nil {
                        fmt.Printf("   ⚠️  Warning: Failed to load manifest: %v\n", err)
                        issues++
                } else {
                        fmt.Printf("   Project: %s v%s\n", m.Package.Name, m.Package.Version)
                        
                        if err := m.Validate(); err != nil {
                                fmt.Printf("   ❌ Manifest validation failed: %v\n", err)
                                issues++
                        }
                }
        } else {
                fmt.Printf("⚠️  Not a Yuki project (no yuki.toml)\n")
        }

        fmt.Print("🌐 GitHub Access: ")
        config := config.GetGlobalConfig()
        token := config.GetGitHubToken()
        if token != "" {
                fmt.Printf("✅ GitHub token configured\n")
        } else {
                fmt.Printf("⚠️  No GitHub token (may hit rate limits)\n")
                fmt.Printf("   Set GITHUB_TOKEN environment variable or run:\n")
                fmt.Printf("   yuki config set github_token <your-token>\n")
        }

        fmt.Print("💾 Cache: ")
        cache := cache.New()
        if size, err := cache.GetSize(); err != nil {
                fmt.Printf("❌ Error accessing cache: %v\n", err)
                issues++
        } else {
                fmt.Printf("✅ Working (size: %d bytes)\n", size)
                
                entries := cache.ListEntries()
                fmt.Printf("   Cached packages: %d\n", len(entries))
        }

        fmt.Print("🌐 Internet Connectivity: ")
        if err := checkInternetConnectivity(); err != nil {
                fmt.Printf("❌ Unable to reach GitHub\n")
                fmt.Printf("   Error: %v\n", err)
                issues++
        } else {
                fmt.Printf("✅ Can reach GitHub\n")
        }

        fmt.Println()
        
        if issues == 0 {
                logger.Success("All checks passed! Your environment is ready to use Yuki.")
        } else {
                logger.Warn("Found %d issue(s). Please address the problems above.", issues)
                return fmt.Errorf("environment issues detected")
        }

        return nil
}

func checkInternetConnectivity() error {
        resp, err := http.Get("https://api.github.com")
        if err != nil {
                return err
        }
        defer resp.Body.Close()
        
        if resp.StatusCode >= 200 && resp.StatusCode < 300 {
                return nil
        }
        
        return fmt.Errorf("HTTP %d", resp.StatusCode)
}
