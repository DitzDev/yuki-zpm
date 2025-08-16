package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"yuki/internal/cache"
	"yuki/internal/logger"
)

func CacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage cache",
		Long:  "Manage Yuki's local cache",
	}

	cmd.AddCommand(cacheCleanCmd())
	cmd.AddCommand(cacheInfoCmd())

	return cmd
}

func cacheCleanCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Clean cache",
		Long:  "Remove cached packages and data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCacheClean(all)
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Remove all cache including internal cache")

	return cmd
}

func cacheInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show cache information",
		Long:  "Display information about the local cache",
		RunE:  runCacheInfo,
	}
}

func runCacheClean(all bool) error {
	cache := cache.New()

	logger.Info("Cleaning cache...")

	if all {
		if err := cache.Clear(); err != nil {
			return fmt.Errorf("failed to clear cache: %w", err)
		}
		logger.Success("Cleared all cache data")
	} else {
		entries := cache.ListEntries()
		if len(entries) == 0 {
			logger.Info("Cache is already empty")
			return nil
		}
		
		for key := range entries {
			cache.Delete(key)
		}

		logger.Success("Cleaned %d cached packages", len(entries))
	}

	return nil
}

func runCacheInfo(cmd *cobra.Command, args []string) error {
	cache := cache.New()

	logger.Info("Cache information:")

	size, err := cache.GetSize()
	if err != nil {
		return fmt.Errorf("failed to get cache size: %w", err)
	}

	entries := cache.ListEntries()

	fmt.Printf("📍 Cache directory: %s\n", cache.GetCacheDir())
	fmt.Printf("📦 Cached packages: %d\n", len(entries))
	fmt.Printf("💾 Cache size: %s\n", formatBytes(size))

	if len(entries) > 0 {
		fmt.Println("\n📋 Cached packages:")
		for key, entry := range entries {
			fmt.Printf("  ├── %s\n", key)
			fmt.Printf("  │   ├── Version: %s\n", entry.Version)
			fmt.Printf("  │   ├── Path: %s\n", entry.Path)
			fmt.Printf("  │   └── Checksum: %s\n", entry.Checksum[:16]+"...")
		}
	}

	return nil
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
