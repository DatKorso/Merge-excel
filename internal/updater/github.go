package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	githubAPIURL = "https://api.github.com/repos/%s/%s/releases/latest"
	requestTimeout = 10 * time.Second
)

// GitHubRelease представляет информацию о релизе из GitHub API
type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	Body        string    `json:"body"`
	HTMLURL     string    `json:"html_url"`
	PublishedAt time.Time `json:"published_at"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
}

// ReleaseInfo информация об обновлении для отображения
type ReleaseInfo struct {
	Version     string
	ReleaseDate time.Time
	DownloadURL string
	Changelog   string
	IsNewer     bool
}

// GitHubClient клиент для работы с GitHub API
type GitHubClient struct {
	owner      string
	repo       string
	httpClient *http.Client
}

// NewGitHubClient создает новый клиент для GitHub API
func NewGitHubClient(owner, repo string) *GitHubClient {
	return &GitHubClient{
		owner: owner,
		repo:  repo,
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
	}
}

// GetLatestRelease получает информацию о последнем релизе из GitHub
func (gc *GitHubClient) GetLatestRelease(ctx context.Context) (*GitHubRelease, error) {
	url := fmt.Sprintf(githubAPIURL, gc.owner, gc.repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	// Устанавливаем заголовки для GitHub API
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "Excel-Merger-Updater")

	resp, err := gc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API вернул статус %d: %s", resp.StatusCode, string(body))
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	return &release, nil
}

// ToReleaseInfo преобразует GitHubRelease в ReleaseInfo
func (r *GitHubRelease) ToReleaseInfo() *ReleaseInfo {
	return &ReleaseInfo{
		Version:     r.TagName,
		ReleaseDate: r.PublishedAt,
		DownloadURL: r.HTMLURL,
		Changelog:   r.Body,
		IsNewer:     false, // Будет установлено при сравнении версий
	}
}
