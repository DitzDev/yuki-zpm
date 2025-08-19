package github

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"yuki_zpm.org/semver"
)

type Client struct {
	token      string
	httpClient *http.Client
}

type Repository struct {
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	HTMLURL     string `json:"html_url"`
	CloneURL    string `json:"clone_url"`
	Stars       int    `json:"stargazers_count"`
	Language    string `json:"language"`
	UpdatedAt   string `json:"updated_at"`
}

type Release struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Draft   bool   `json:"draft"`
	Prerelease bool `json:"prerelease"`
}

type SearchResult struct {
	Items []Repository `json:"items"`
}

func NewClient() *Client {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = os.Getenv("GH_TOKEN")
	}

	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) makeRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "yuki-package-manager")
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	if c.token != "" {
		req.Header.Set("Authorization", "token "+c.token)
	}

	return c.httpClient.Do(req)
}

func ParseRepoURL(repoURL string) (owner, repo string, err error) {
	// Handle different formats:
	// - username/repo
	// - https://github.com/username/repo
	// - https://github.com/username/repo.git
	// - git@github.com:username/repo.git

	if !strings.Contains(repoURL, "/") {
		return "", "", fmt.Errorf("invalid repository format: %s", repoURL)
	}

	if !strings.Contains(repoURL, "://") && !strings.Contains(repoURL, "@") {
		parts := strings.Split(repoURL, "/")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid repository format: %s", repoURL)
		}
		return parts[0], parts[1], nil
	}
	
	if strings.HasPrefix(repoURL, "http") {
		u, err := url.Parse(repoURL)
		if err != nil {
			return "", "", fmt.Errorf("invalid URL: %s", repoURL)
		}

		if u.Host != "github.com" {
			return "", "", fmt.Errorf("only GitHub repositories are supported")
		}

		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		if len(parts) < 2 {
			return "", "", fmt.Errorf("invalid repository URL: %s", repoURL)
		}

		repo := parts[1]
		if strings.HasSuffix(repo, ".git") {
			repo = repo[:len(repo)-4]
		}

		return parts[0], repo, nil
	}

	if strings.Contains(repoURL, "@") {
		re := regexp.MustCompile(`git@github\.com:([^/]+)/(.+?)(?:\.git)?$`)
		matches := re.FindStringSubmatch(repoURL)
		if len(matches) != 3 {
			return "", "", fmt.Errorf("invalid SSH repository URL: %s", repoURL)
		}
		return matches[1], matches[2], nil
	}

	return "", "", fmt.Errorf("unsupported repository URL format: %s", repoURL)
}

func (c *Client) GetRepository(owner, repo string) (*Repository, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	
	resp, err := c.makeRequest(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("repository not found: %s/%s", owner, repo)
	}
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var repository Repository
	if err := json.Unmarshal(body, &repository); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &repository, nil
}

func (c *Client) GetReleases(owner, repo string) ([]Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", owner, repo)
	
	resp, err := c.makeRequest(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var releases []Release
	if err := json.Unmarshal(body, &releases); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return releases, nil
}

func (c *Client) GetTags(owner, repo string) ([]string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/tags", owner, repo)
	
	resp, err := c.makeRequest(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tags: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var tags []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(body, &tags); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var tagNames []string
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}

	return tagNames, nil
}

func (c *Client) SearchRepositories(query string) ([]Repository, error) {
	searchQuery := fmt.Sprintf("%s+topic:zig-package", url.QueryEscape(query))
	url := fmt.Sprintf("https://api.github.com/search/repositories?q=%s", searchQuery)
	
	resp, err := c.makeRequest(url)
	if err != nil {
		return nil, fmt.Errorf("failed to search repositories: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var searchResult SearchResult
	if err := json.Unmarshal(body, &searchResult); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return searchResult.Items, nil
}

func (c *Client) GetAvailableVersions(owner, repo string) ([]semver.Version, error) {
	tags, err := c.GetTags(owner, repo)
	if err != nil {
		return nil, err
	}

	var versions []semver.Version
	for _, tag := range tags {
		if version, err := semver.ParseVersion(tag); err == nil {
			versions = append(versions, version)
		}
	}

	return versions, nil
}

func (c *Client) GetLatestRelease(owner, repo string) (*Release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	
	resp, err := c.makeRequest(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("no releases found")
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var release Release
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &release, nil
}
