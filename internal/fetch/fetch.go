package fetch

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"yuki/internal/cache"
	"yuki/internal/github"
	"yuki/internal/integrity"
	"yuki/internal/logger"
	"yuki/internal/manifest"
)

type Fetcher struct {
	cache        *cache.Cache
	githubClient *github.Client
}

type FetchResult struct {
	Path     string
	Checksum string
	Version  string
}

func NewFetcher() *Fetcher {
	return &Fetcher{
		cache:        cache.New(),
		githubClient: github.NewClient(),
	}
}

// checkTagExists checks if a tag exists in the remote repository
func (f *Fetcher) checkTagExists(owner, repo, tag string) (bool, error) {
	repoURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)
	
	cmd := exec.Command("git", "ls-remote", "--tags", repoURL, tag)
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	
	return strings.Contains(string(output), tag), nil
}

func (f *Fetcher) getLatestReleaseTag(owner, repo string) (string, error) {
	release, err := f.githubClient.GetLatestRelease(owner, repo)
	if err != nil {
		return "", err
	}
	return release.TagName, nil
}

func (f *Fetcher) determineRefWithValidation(owner, repo string, dep manifest.Dependency) (string, string, error) {
	if dep.Rev != "" {
		return dep.Rev, dep.Rev, nil
	}
	if dep.Tag != "" {
		return dep.Tag, dep.Tag, nil
	}
	if dep.Branch != "" {
		return dep.Branch, dep.Branch, nil
	}
	
	if dep.Version != "" {
		version := strings.TrimPrefix(dep.Version, "^")
		version = strings.TrimPrefix(version, "~")
		version = strings.TrimPrefix(version, "=")
		
		possibleTags := []string{
			version,        // example 0.10.0
			"v" + version,  // example v0.10.0
		}
		
		for _, tag := range possibleTags {
			exists, err := f.checkTagExists(owner, repo, tag)
			if err != nil {
				logger.Debug("Failed to check tag %s: %v", tag, err)
				continue
			}
			if exists {
				return tag, dep.Version, nil
			}
		}
		
		return "", "", fmt.Errorf("no matching tag found for version %s (tried: %s)", dep.Version, strings.Join(possibleTags, ", "))
	}

	latestTag, err := f.getLatestReleaseTag(owner, repo)
	if err == nil && latestTag != "" {
		logger.Debug("Using latest release tag: %s", latestTag)
		return latestTag, latestTag, nil
	}
	
	logger.Debug("No releases found, using main branch")
	return "main", "main", nil
}

func (f *Fetcher) FetchDependency(name string, dep manifest.Dependency) (*FetchResult, error) {
	logger.Info("Fetching dependency '%s'", name)
	
	owner, repo, err := github.ParseRepoURL(dep.Git)
	if err != nil {
		return nil, fmt.Errorf("invalid git URL for '%s': %w", name, err)
	}
	
	cacheKey := generateCacheKey(owner, repo, dep)

	if cached, exists := f.cache.Get(cacheKey); exists {
		logger.Debug("Using cached version of '%s'", name)
		return &FetchResult{
			Path:     cached.Path,
			Checksum: cached.Checksum,
			Version:  cached.Version,
		}, nil
	}
	
	ref, resolvedVersion, err := f.determineRefWithValidation(owner, repo, dep)
	if err != nil {
		return nil, fmt.Errorf("failed to determine reference for '%s': %w", name, err)
	}

	repoPath, err := f.cloneRepository(owner, repo, ref)
	if err != nil {
		return nil, fmt.Errorf("failed to clone repository '%s/%s': %w", owner, repo, err)
	}

	checksum, err := integrity.CalculateDirectoryChecksum(repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate checksum for '%s': %w", name, err)
	}

	result := &FetchResult{
		Path:     repoPath,
		Checksum: checksum,
		Version:  resolvedVersion,
	}

	f.cache.Set(cacheKey, cache.Entry{
		Path:     repoPath,
		Checksum: checksum,
		Version:  resolvedVersion,
	})

	logger.Success("Successfully fetched '%s@%s'", name, resolvedVersion)
	return result, nil
}

func (f *Fetcher) cloneRepository(owner, repo, ref string) (string, error) {
	repoURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)

	targetDir := filepath.Join(f.cache.GetCacheDir(), "repos", owner, repo, ref)
	if err := os.RemoveAll(targetDir); err != nil {
		return "", fmt.Errorf("failed to clean target directory: %w", err)
	}
	
	if err := os.MkdirAll(filepath.Dir(targetDir), 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	logger.Debug("Cloning %s@%s to %s", repoURL, ref, targetDir)

	cmd := exec.Command("git", "clone", "--depth=1", "--branch", ref, repoURL, targetDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "does not exist") || strings.Contains(string(output), "not found") {
			logger.Debug("Shallow clone failed, trying full clone: %s", string(output))

			os.RemoveAll(targetDir)
			
			cmd = exec.Command("git", "clone", repoURL, targetDir)
			if err := cmd.Run(); err != nil {
				return "", fmt.Errorf("failed to clone repository: %w\nOutput: %s", err, output)
			}

			cmd = exec.Command("git", "checkout", ref)
			cmd.Dir = targetDir
			if err := cmd.Run(); err != nil {
				return "", fmt.Errorf("failed to checkout ref '%s': %w", ref, err)
			}
		} else {
			return "", fmt.Errorf("failed to clone repository: %w\nOutput: %s", err, output)
		}
	}
	
	gitDir := filepath.Join(targetDir, ".git")
	os.RemoveAll(gitDir)

	return targetDir, nil
}

func generateCacheKey(owner, repo string, dep manifest.Dependency) string {
	var parts []string
	parts = append(parts, owner, repo)
	
	if dep.Rev != "" {
		parts = append(parts, "rev", dep.Rev)
	} else if dep.Tag != "" {
		parts = append(parts, "tag", dep.Tag)
	} else if dep.Branch != "" {
		parts = append(parts, "branch", dep.Branch)
	} else if dep.Version != "" {
		parts = append(parts, "version", dep.Version)
	} else {
		parts = append(parts, "latest")
	}
	
	return strings.Join(parts, "-")
}

func (f *Fetcher) VerifyDependency(path, expectedChecksum string) error {
	actualChecksum, err := integrity.CalculateDirectoryChecksum(path)
	if err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

func CheckGitAvailable() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git is not installed or not available in PATH")
	}
	return nil
}