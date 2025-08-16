package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"yuki/internal/github"
	"yuki/internal/logger"
)

func SearchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search for packages",
		Long:  "Search for Zig packages on GitHub",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSearch(args[0])
		},
	}
}

func runSearch(query string) error {
	logger.Info("Searching for packages matching '%s'...", query)

	client := github.NewClient()
	repos, err := client.SearchRepositories(query)
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}

	if len(repos) == 0 {
		logger.Info("No packages found matching '%s'", query)
		return nil
	}

	logger.Info("Found %d packages:", len(repos))
	fmt.Println()

	for i, repo := range repos {
		if i >= 10 { 
			logger.Info("... and %d more results", len(repos)-10)
			break
		}

		fmt.Printf("📦 %s\n", repo.FullName)
		if repo.Description != "" {
			fmt.Printf("   %s\n", repo.Description)
		}
		fmt.Printf("   ⭐ %d stars | Language: %s\n", repo.Stars, repo.Language)
		fmt.Printf("   🔗 %s\n", repo.HTMLURL)
		fmt.Printf("   📥 yuki add %s\n", repo.FullName)
		fmt.Println()
	}

	return nil
}
