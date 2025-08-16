package resolver

import (
	"fmt"
	"sort"

	"yuki/internal/github"
	"yuki/internal/logger"
	"yuki/internal/manifest"
	"yuki/internal/semver"
)

type Resolver struct {
	githubClient *github.Client
}

type ResolvedDependency struct {
	Name     string
	Version  string
	Source   string
	Checksum string
	Deps     []string
}

type Resolution struct {
	Dependencies []ResolvedDependency
}

func New() *Resolver {
	return &Resolver{
		githubClient: github.NewClient(),
	}
}

func (r *Resolver) Resolve(m *manifest.Manifest) (*Resolution, error) {
	logger.Info("Resolving dependencies...")
	
	resolved := make(map[string]ResolvedDependency)

	allDeps := m.GetAllDependencies()
	
	for name, dep := range allDeps {
		if err := r.resolveDependency(name, dep, resolved); err != nil {
			return nil, fmt.Errorf("failed to resolve dependency '%s': %w", name, err)
		}
	}
	
	var deps []ResolvedDependency
	for _, dep := range resolved {
		deps = append(deps, dep)
	}
	
	sort.Slice(deps, func(i, j int) bool {
		return deps[i].Name < deps[j].Name
	})
	
	logger.Success("Successfully resolved %d dependencies", len(deps))
	
	return &Resolution{Dependencies: deps}, nil
}

func (r *Resolver) resolveDependency(name string, dep manifest.Dependency, resolved map[string]ResolvedDependency) error {
	if _, exists := resolved[name]; exists {
		return nil
	}
	
	logger.Debug("Resolving dependency: %s", name)
	
	owner, repo, err := github.ParseRepoURL(dep.Git)
	if err != nil {
		return fmt.Errorf("invalid git URL: %w", err)
	}
	
	resolvedVersion, err := r.resolveVersion(owner, repo, dep)
	if err != nil {
		return fmt.Errorf("failed to resolve version: %w", err)
	}

	resolvedDep := ResolvedDependency{
		Name:     name,
		Version:  resolvedVersion,
		Source:   dep.Git,
		Checksum: "", 
		Deps:     []string{}, 
	}
	
	resolved[name] = resolvedDep
	
	
	
	
	return nil
}

func (r *Resolver) resolveVersion(owner, repo string, dep manifest.Dependency) (string, error) {
	if dep.Rev != "" {
		return dep.Rev, nil
	}
	if dep.Tag != "" {
		return dep.Tag, nil
	}
	if dep.Branch != "" {
		return dep.Branch, nil
	}

	if dep.Version != "" {
		constraint, err := semver.ParseConstraint(dep.Version)
		if err != nil {
			return "", fmt.Errorf("invalid version constraint '%s': %w", dep.Version, err)
		}
	
		versions, err := r.githubClient.GetAvailableVersions(owner, repo)
		if err != nil {
			return "", fmt.Errorf("failed to get available versions: %w", err)
		}
		
		if len(versions) == 0 {
			return "", fmt.Errorf("no semantic versions found for %s/%s", owner, repo)
		}

		bestMatch, err := semver.FindBestMatch(constraint, versions)
		if err != nil {
			return "", fmt.Errorf("no version satisfies constraint '%s': %w", dep.Version, err)
		}
		
		return bestMatch.String(), nil
	}

	if release, err := r.githubClient.GetLatestRelease(owner, repo); err == nil {
		logger.Debug("Using latest release: %s", release.TagName)
		return release.TagName, nil
	}

	logger.Debug("No releases found for %s/%s, using main branch", owner, repo)
	return "main", nil
}

func (r *Resolver) CheckForUpdates(lockFile *manifest.LockFile) ([]UpdateInfo, error) {
	var updates []UpdateInfo
	
	for _, pkg := range lockFile.Package {
		owner, repo, err := github.ParseRepoURL(pkg.Source)
		if err != nil {
			continue
		}
	
		if release, err := r.githubClient.GetLatestRelease(owner, repo); err == nil {
			currentVersion, err := semver.ParseVersion(pkg.Version)
			if err != nil {
				continue
			}
			
			latestVersion, err := semver.ParseVersion(release.TagName)
			if err != nil {
				continue
			}
			
			if latestVersion.Compare(currentVersion) > 0 {
				updates = append(updates, UpdateInfo{
					Name:           pkg.Name,
					CurrentVersion: pkg.Version,
					LatestVersion:  release.TagName,
					Source:         pkg.Source,
				})
			}
		}
	}
	
	return updates, nil
}

type UpdateInfo struct {
	Name           string
	CurrentVersion string
	LatestVersion  string
	Source         string
}

func (r *Resolver) ValidateDependencies(m *manifest.Manifest) error {
	allDeps := m.GetAllDependencies()
	
	for name, dep := range allDeps {
		owner, repo, err := github.ParseRepoURL(dep.Git)
		if err != nil {
			return fmt.Errorf("dependency '%s' has invalid git URL: %w", name, err)
		}
	
		if _, err := r.githubClient.GetRepository(owner, repo); err != nil {
			return fmt.Errorf("dependency '%s' repository not accessible: %w", name, err)
		}

		if dep.Version != "" {
			if _, err := semver.ParseConstraint(dep.Version); err != nil {
				return fmt.Errorf("dependency '%s' has invalid version constraint '%s': %w", name, dep.Version, err)
			}
		}
	}
	
	return nil
}