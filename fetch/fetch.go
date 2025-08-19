package fetch

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"yuki_zpm.org/cache"
	"yuki_zpm.org/github"
	"yuki_zpm.org/integrity"
	"yuki_zpm.org/logger"
	"yuki_zpm.org/manifest"
	"yuki_zpm.org/utils"
)

type Fetcher struct {
	cache        *cache.Cache
	githubClient *github.Client
}

type FetchResult struct {
	Path      string
	Checksum  string
	Version   string
	CommitSHA string
}

func NewFetcher() *Fetcher {
	return &Fetcher{
		cache        : cache.New(),
		githubClient : github.NewClient(),
	}
}

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

func (f *Fetcher) getLatestCommitSHA(owner, repo string) (string, error) {
	repoURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)

	cmd := exec.Command("git", "ls-remote", repoURL, "refs/heads/main")
	output, err := cmd.Output()
	if err == nil && len(strings.TrimSpace(string(output))) > 0 {
		parts := strings.Fields(string(output))
		if len(parts) > 0 {
			return parts[0], nil
		}
	}

	cmd = exec.Command("git", "ls-remote", repoURL, "refs/heads/master")
	output, err = cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get latest commit: %w", err)
	}
	
	parts := strings.Fields(string(output))
	if len(parts) == 0 {
		return "", fmt.Errorf("no commit found in master branch")
	}
	
	return parts[0], nil
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

	if len(ref) == 40 && utils.IsHexString(ref) {
		cmd := exec.Command("git", "clone", repoURL, targetDir)
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to clone repository: %w", err)
		}

		cmd = exec.Command("git", "checkout", ref)
		cmd.Dir = targetDir
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to checkout commit '%s': %w", ref, err)
		}
	} else {
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
	}
	
	gitDir := filepath.Join(targetDir, ".git")
	os.RemoveAll(gitDir)

	return targetDir, nil
}

func (f *Fetcher) determineRefWithValidation(owner, repo string, dep manifest.Dependency) (string, string, string, error) {
	if dep.Rev != "" {
		return dep.Rev, dep.Rev, dep.Rev, nil
	}
	if dep.Tag != "" {
		return dep.Tag, dep.Tag, "", nil
	}
	if dep.Branch != "" {
		return dep.Branch, dep.Branch, "", nil
	}
	
	if dep.UseLatestCommit {
		logger.Info("Fetching latest commit...")
		commitSHA, err := f.getLatestCommitSHA(owner, repo)
		if err != nil {
			return "", "", "", fmt.Errorf("failed to get latest commit: %w", err)
		}
		logger.Debug("Latest commit SHA: %s", commitSHA)
		return commitSHA, commitSHA, commitSHA, nil
	}
	
	if dep.Version != "" {
		if dep.Version == "latest" {
			latestTag, err := f.getLatestReleaseTag(owner, repo)
			if err != nil {
				return "", "", "", fmt.Errorf("failed to get latest release: %w", err)
			}
			logger.Debug("Using latest release tag: %s", latestTag)
			return latestTag, latestTag, "", nil
		}
		
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
				return tag, dep.Version, "", nil
			}
		}
		
		return "", "", "", fmt.Errorf("no matching tag found for version %s (tried: %s)", dep.Version, strings.Join(possibleTags, ", "))
	}

	latestTag, err := f.getLatestReleaseTag(owner, repo)
	if err == nil && latestTag != "" {
		logger.Debug("Using latest release tag: %s", latestTag)
		return latestTag, latestTag, "", nil
	}

	logger.Debug("No releases found, using latest commit")
	commitSHA, err := f.getLatestCommitSHA(owner, repo)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get latest commit: %w", err)
	}
	return commitSHA, commitSHA, commitSHA, nil
}

func (f *Fetcher) FetchDependency(name string, dep manifest.Dependency) (*FetchResult, error) {
	logger.Info("Fetching dependency '%s'", name)
	
	owner, repo, err := github.ParseRepoURL(dep.Git)
	if err != nil {
		return nil, fmt.Errorf("invalid git URL for '%s': %w", name, err)
	}
	
	cacheKey := utils.GenerateCacheKey(owner, repo, dep)

	if cached, exists := f.cache.Get(cacheKey); exists {
		logger.Debug("Using cached version of '%s'", name)
		return &FetchResult{
			Path:      cached.Path,
			Checksum:  cached.Checksum,
			Version:   cached.Version,
			CommitSHA: cached.CommitSHA,
		}, nil
	}
	
	ref, resolvedVersion, commitSHA, err := f.determineRefWithValidation(owner, repo, dep)
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
		Path:      repoPath,
		Checksum:  checksum,
		Version:   resolvedVersion,
		CommitSHA: commitSHA,
	}

	f.cache.Set(cacheKey, cache.Entry{
		Path:      repoPath,
		Checksum:  checksum,
		Version:   resolvedVersion,
		CommitSHA: commitSHA,
	})

	if commitSHA != "" {
		logger.Success("Successfully fetched '%s@%s' (commit: %s)", name, resolvedVersion, commitSHA[:8])
	} else {
		logger.Success("Successfully fetched '%s@%s'", name, resolvedVersion)
	}
	return result, nil
}
