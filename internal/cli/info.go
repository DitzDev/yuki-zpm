package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"yuki/internal/github"
	"yuki/internal/logger"
)

func InfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <package>",
		Short: "Show package information",
		Long:  "Display detailed information about a package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInfo(args[0])
		},
	}
}

func runInfo(packageSpec string) error {
	owner, repo, err := github.ParseRepoURL(packageSpec)
	if err != nil {
		return fmt.Errorf("invalid package specification: %w", err)
	}

	logger.Info("Getting information for '%s/%s'...", owner, repo)

	client := github.NewClient()
	
	repository, err := client.GetRepository(owner, repo)
	if err != nil {
		return fmt.Errorf("failed to get repository info: %w", err)
	}

	releases, err := client.GetReleases(owner, repo)
	if err != nil {
		logger.Debug("Failed to get releases: %v", err)
		releases = []github.Release{} 
	}

	fmt.Printf("📦 %s\n", repository.FullName)
	fmt.Printf("📝 %s\n", repository.Description)
	fmt.Printf("🔗 %s\n", repository.HTMLURL)
	fmt.Printf("⭐ %d stars\n", repository.Stars)
	fmt.Printf("🔧 Language: %s\n", repository.Language)
	fmt.Printf("📅 Updated: %s\n", repository.UpdatedAt)
	fmt.Printf("📥 Clone: %s\n", repository.CloneURL)

	if len(releases) > 0 {
		fmt.Println("\n🏷️  Recent Releases:")
		count := len(releases)
		if count > 5 {
			count = 5 
		}
		
		for i := 0; i < count; i++ {
			release := releases[i]
			status := ""
			if release.Prerelease {
				status = " (prerelease)"
			}
			if release.Draft {
				status = " (draft)"
			}
			fmt.Printf("   %s%s", release.TagName, status)
			if release.Name != "" && release.Name != release.TagName {
				fmt.Printf(" - %s", release.Name)
			}
			fmt.Println()
		}
	}

	fmt.Println("\n📥 Installation:")
	fmt.Printf("   yuki add %s\n", repository.FullName)

	return nil
}
