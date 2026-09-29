package main 

import (
    "context"
    "encoding/json"
    "fmt"
    "log/slog"
    "net/http"
    "sync"
    "time"
)

type Repository struct {
	Name		string		`json:"name"`
	Description	string		`json:"description"`
	URL		string		`json:"html_url"`
	Language	string		`json:"language"`
	Stars		int		`json:"stargazers_count"`
	Fork		bool		`json:"fork"`
	UpdatedAt	time.Time	`json:"pushed_at"`
}
type GitHubCache struct {
	mu        sync.RWMutex
	repos     []Repository
	updatedAt time.Time
	ttl       time.Duration
	client    *http.Client
	logger    *slog.Logger
}

// NewGitHubCache initializes a thread-safe cache with a configurable TTL.
func NewGitHubCache(ttl time.Duration, logger *slog.Logger) *GitHubCache {
	return &GitHubCache{
		ttl:    ttl,
		logger: logger,
		client: &http.Client{
			Timeout: 10 * time.Second, // Resilient timeouts are non-negotiable
		},
	}
}

// Get retrieves the cached repositories or triggers an on-demand background 
// update if the cache has expired.
func (c *GitHubCache) Get(ctx context.Context, username string) ([]Repository, error) {
	c.mu.RLock()
	isFresh := time.Since(c.updatedAt) < c.ttl && len(c.repos) > 0
	if isFresh {
		repos := c.repos
		c.mu.RUnlock()
		return repos, nil
	}
	c.mu.RUnlock()

	// Cache expired or empty: acquire write lock to fetch fresh data
	c.mu.Lock()
	defer c.mu.Unlock()

	// Double-check freshness under write lock (prevents duplicate API storm)
	if time.Since(c.updatedAt) < c.ttl && len(c.repos) > 0 {
		return c.repos, nil
	}

	c.logger.Info("fetching fresh repositories from GitHub", "username", username)
	freshRepos, err := c.fetch(ctx, username)
	if err != nil {
		if len(c.repos) > 0 {
			c.logger.Error("github fetch failed; serving stale fallback cache", "error", err)
			return c.repos, nil // Graceful degradation: serve stale data rather than crashing
		}
		return nil, err
	}

	c.repos = freshRepos
	c.updatedAt = time.Now()
	return c.repos, nil
}

// fetch performs the actual outbound HTTPS call to the GitHub API.
func (c *GitHubCache) fetch(ctx context.Context, username string) ([]Repository, error) {
	url := fmt.Sprintf("https://api.github.com/users/%s/repos?sort=pushed&per_page=100", username)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Identify our user-agent cleanly as required by GitHub API guidelines
	req.Header.Set("User-Agent", "gopherit-dev-portfolio")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var rawRepos []Repository
	if err := json.NewDecoder(resp.Body).Decode(&rawRepos); err != nil {
		return nil, err
	}

	// Filter out forks so only your authentic, original repositories display
	var originalRepos []Repository
	for _, r := range rawRepos {
		if !r.Fork {
			originalRepos = append(originalRepos, r)
		}
	}

	return originalRepos, nil
}
