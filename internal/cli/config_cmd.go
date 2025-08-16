package cli

import (
        "fmt"
        "os"
        "strings"

        "github.com/spf13/cobra"
        "yuki/internal/config"
        "yuki/internal/logger"
)

func ConfigCmd() *cobra.Command {
        cmd := &cobra.Command{
                Use:   "config",
                Short: "Manage configuration",
                Long:  "Manage Yuki's global configuration",
        }

        cmd.AddCommand(configGetCmd())
        cmd.AddCommand(configSetCmd())
        cmd.AddCommand(configListCmd())
        cmd.AddCommand(configDeleteCmd())

        return cmd
}

func configGetCmd() *cobra.Command {
        return &cobra.Command{
                Use:   "get <key>",
                Short: "Get configuration value",
                Long:  "Get the value of a configuration key",
                Args:  cobra.ExactArgs(1),
                RunE: func(cmd *cobra.Command, args []string) error {
                        return runConfigGet(args[0])
                },
        }
}

func configSetCmd() *cobra.Command {
        return &cobra.Command{
                Use:   "set <key> <value>",
                Short: "Set configuration value",
                Long:  "Set the value of a configuration key",
                Args:  cobra.ExactArgs(2),
                RunE: func(cmd *cobra.Command, args []string) error {
                        return runConfigSet(args[0], args[1])
                },
        }
}

func configListCmd() *cobra.Command {
        return &cobra.Command{
                Use:   "list",
                Short: "List all configuration",
                Long:  "List all configuration keys and values",
                RunE:  runConfigList,
        }
}

func configDeleteCmd() *cobra.Command {
        return &cobra.Command{
                Use:   "delete <key>",
                Short: "Delete configuration key",
                Long:  "Delete a configuration key",
                Args:  cobra.ExactArgs(1),
                RunE: func(cmd *cobra.Command, args []string) error {
                        return runConfigDelete(args[0])
                },
        }
}

func runConfigGet(key string) error {
        config := config.GetGlobalConfig()

        value, exists := config.Get(key)
        if !exists {
                logger.Error("Configuration key '%s' not found", key)
                return fmt.Errorf("key not found")
        }

        
        if isSensitiveKey(key) {
                fmt.Printf("%s = [HIDDEN]\n", key)
        } else {
                fmt.Printf("%s = %v\n", key, value)
        }

        return nil
}

func runConfigSet(key, value string) error {
        config := config.GetGlobalConfig()

        
        if err := validateConfigKey(key, value); err != nil {
                return fmt.Errorf("invalid configuration: %w", err)
        }

        if err := config.Set(key, value); err != nil {
                return fmt.Errorf("failed to set configuration: %w", err)
        }

        if isSensitiveKey(key) {
                logger.Success("Set %s = [HIDDEN]", key)
        } else {
                logger.Success("Set %s = %s", key, value)
        }

        return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
        config := config.GetGlobalConfig()
        configMap := config.List()

        if len(configMap) == 0 {
                logger.Info("No configuration set")
                return nil
        }

        logger.Info("Configuration:")
        for key, value := range configMap {
                if isSensitiveKey(key) {
                        fmt.Printf("  %s = [HIDDEN]\n", key)
                } else {
                        fmt.Printf("  %s = %v\n", key, value)
                }
        }

        
        fmt.Println("\nEnvironment variables (override config):")
        envVars := []string{"GITHUB_TOKEN", "GH_TOKEN"}
        for _, envVar := range envVars {
                if value := os.Getenv(envVar); value != "" {
                        fmt.Printf("  %s = [SET]\n", envVar)
                } else {
                        fmt.Printf("  %s = [NOT SET]\n", envVar)
                }
        }

        return nil
}

func runConfigDelete(key string) error {
        config := config.GetGlobalConfig()

        _, exists := config.Get(key)
        if !exists {
                logger.Error("Configuration key '%s' not found", key)
                return fmt.Errorf("key not found")
        }

        if err := config.Delete(key); err != nil {
                return fmt.Errorf("failed to delete configuration: %w", err)
        }

        logger.Success("Deleted configuration key '%s'", key)
        return nil
}

func isSensitiveKey(key string) bool {
        sensitiveKeys := []string{
                "github_token",
                "token",
                "password",
                "secret",
                "key",
        }

        key = strings.ToLower(key)
        for _, sensitive := range sensitiveKeys {
                if strings.Contains(key, sensitive) {
                        return true
                }
        }
        return false
}

func validateConfigKey(key, value string) error {
        switch key {
        case "github_token":
                if len(value) < 10 {
                        return fmt.Errorf("GitHub token appears to be too short")
                }
                if !strings.HasPrefix(value, "ghp_") && !strings.HasPrefix(value, "github_pat_") {
                        logger.Warn("GitHub token format may be incorrect (should start with 'ghp_' or 'github_pat_')")
                }
        case "default_branch":
                if value == "" {
                        return fmt.Errorf("default branch cannot be empty")
                }
        case "cache_size_limit":
                
        }

        return nil
}
