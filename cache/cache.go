package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

var directCache = map[string]struct{}{
	"darwin":   {},
	"exe":      {},
	"dmg":      {},
	"rpm":      {},
	"deb":      {},
	"AppImage": {},
}

func checkPlatform(filename string) string {
	extension := filepath.Ext(filename)
	extension = strings.TrimLeft(extension, ".")
	if (strings.Contains(filename, "mac") || strings.Contains(filename, "darwin")) && extension == "zip" {
		return "darwin"
	}

	_, ok := directCache[extension]
	if ok {
		return extension
	}

	return ""
}

// Asset is the released package.
type Asset struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	URL                string `json:"url"`
	BrowserDownloadURL string `json:"browserDownloadURL"`
	ContentType        string `json:"contentType"`
	Size               int    `json:"size"`
}

// Release contains major info of every release record.
type Release struct {
	Version   string            `json:"version"`
	Notes     string            `json:"notes"`
	PubDate   github.Timestamp  `json:"pubDate"`
	Platforms map[string]*Asset `json:"platforms"`
	RELEASES  string            `json:"-"`
}

// ReleaseData release info data for caching into file.
type ReleaseData struct {
	Release           *Release `json:"release"`
	RepoURL           string   `json:"repoUrl"`
	OpenProxyDownload bool     `json:"openProxyDownload"`
}

// ProxyDownloadConfig of proxy download files with current server.
type ProxyDownloadConfig struct {
	SaveDir string `yaml:"saveDir"`
	BaseURL string `yaml:"baseURL"`
}

// GithubConfig of the cache
type GithubConfig struct {
	Owner string `yaml:"owner"`
	Repo  string `yaml:"repo"`
	Token string `yaml:"token"`
	Pre   bool   `yaml:"pre"`
}

// RepoURL returns repo URL on github.
func (c *GithubConfig) RepoURL() string {
	return fmt.Sprintf("github.com/%s/%s", c.Owner, c.Repo)
}

// IsPrivateRepo if is private repo should proxy assets download.
func (c *GithubConfig) IsPrivateRepo() bool {
	return c.Token != ""
}

// GithubCache caches release information fetching from github.
type GithubCache struct {
	conf              *GithubConfig
	cacheURLBase      string
	openProxyDownload bool
	cacheDir          string
	latest            *Release
	latestMu          sync.RWMutex
	latestUpdate      time.Time
}

// NewGithubCache .
func NewGithubCache(conf *GithubConfig, cacheDir string, openProxyDownload bool, cacheURLBase string) *GithubCache {
	g := &GithubCache{
		conf:              conf,
		openProxyDownload: openProxyDownload,
		cacheURLBase:      cacheURLBase,
		cacheDir:          cacheDir,
	}
	log.Info().Str("url", conf.RepoURL()).Bool("private", conf.IsPrivateRepo()).Msg("Github repo")

	g.loadReleaseCache()
	go g.runRefreshLoop()
	return g
}

// AssetFilePath generates file path for caching asset.
func (g *GithubCache) AssetFilePath(release *Release, assetName string) string {
	return filepath.Join(g.cacheDir, g.conf.Owner, g.conf.Repo, release.Version, assetName)
}

// AssetFileURL generates file download url of cached asset.
func (g *GithubCache) AssetFileURL(release *Release, assetName string) string {
	u, _ := url.Parse(g.cacheURLBase)
	u.Path = filepath.Join(u.Path, g.conf.Owner, g.conf.Repo, release.Version, assetName)
	return u.String()
}

func (g *GithubCache) newClient(ctx context.Context) *github.Client {
	if g.conf.Token == "" {
		return github.NewClient(nil)

	}
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.conf.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func (g *GithubCache) refreshCache() error {
	ctx := context.Background()
	client := g.newClient(ctx)
	releases, _, err := client.Repositories.ListReleases(ctx, g.conf.Owner, g.conf.Repo, &github.ListOptions{
		PerPage: 10,
	})

	if err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			log.Error().Msg("hit rate limit")
		}
		return err
	}

	var release *github.RepositoryRelease
	for _, item := range releases {
		if *item.Draft {
			continue
		}
		if *item.Prerelease && !g.conf.Pre {
			continue
		}
		if len(item.Assets) == 0 {
			continue
		}
		release = item
		break
	}

	if release == nil {
		return nil
	}

	g.latestMu.RLock()
	latestPrev := g.latest
	g.latestMu.RUnlock()

	if latestPrev != nil && latestPrev.Version == *release.TagName && latestPrev.PubDate.Equal(*release.PublishedAt) {
		g.latestUpdate = time.Now()
		return nil
	}

	latest := &Release{
		Version:   *release.TagName,
		Notes:     *release.Body,
		PubDate:   *release.PublishedAt,
		Platforms: make(map[string]*Asset),
	}
	log.Info().Str("version", latest.Version).Msg("Caching...")
	for _, asset := range release.Assets {
		if *asset.Name == "RELEASES" {
			log.Debug().Interface("asset", asset).Msg("RELEASES")
			content, err := g.cacheReleaseList(ctx, *asset.ID, *asset.BrowserDownloadURL)
			if err != nil {
				return err
			}

			latest.RELEASES = content
			continue
		}

		platform := checkPlatform(*asset.Name)
		if platform == "" {
			continue
		}

		a := &Asset{
			ID:                 *asset.ID,
			Name:               *asset.Name,
			URL:                *asset.URL,
			BrowserDownloadURL: *asset.BrowserDownloadURL,
			ContentType:        *asset.ContentType,
			Size:               (*asset.Size) / 1000000 * 10 / 10,
		}

		log.Info().Str("asset", *asset.Name).Str("platform", platform).Msg("Cache asset")
		// Download asset into cache dir.
		if g.openProxyDownload {
			if err := g.cacheAssetFile(latest, a); err != nil {
				return err
			}
		}

		latest.Platforms[platform] = a
	}

	g.latestMu.Lock()
	g.latest = latest
	g.latestMu.Unlock()

	// Clean old cached assets
	if latestPrev != nil {
		for _, a := range latestPrev.Platforms {
			fn := g.AssetFilePath(latestPrev, a.Name)
			if err := os.Remove(fn); err != nil {
				log.Error().Err(err).Str("file", fn).Msg("Remove old asset")
			}
		}
	}

	g.latestUpdate = time.Now()

	// Cache release data for loading as basic data at next startup.
	// In case there is no any data while network error occurred at startup.
	g.cacheReleaseLastest(latest)

	log.Info().Str("version", latest.Version).Msg("Finished caching")
	return nil
}

func (g *GithubCache) loadReleaseCache() {
	filename := filepath.Join(g.cacheDir, "release.json")
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		log.Error().Err(err).Msg("Read release")
		return
	}

	var data ReleaseData
	if err := json.Unmarshal(b, &data); err != nil {
		log.Error().Err(err).Msg("Marshal release")
		return
	}

	if data.RepoURL != g.conf.RepoURL() {
		return
	}

	if data.OpenProxyDownload != g.openProxyDownload {
		return
	}

	g.latest = data.Release
	log.Info().Str("version", g.latest.Version).Str("file", filename).Msg("Loaded release data from cache")
}

func (g *GithubCache) cacheReleaseLastest(release *Release) {
	data := &ReleaseData{
		Release:           release,
		RepoURL:           g.conf.RepoURL(),
		OpenProxyDownload: g.openProxyDownload,
	}
	b, err := json.Marshal(data)
	if err != nil {
		log.Error().Err(err).Msg("Marshal release data")
		return
	}

	filename := filepath.Join(g.cacheDir, "release.json")
	if err := ioutil.WriteFile(filename, b, 0644); err != nil {
		log.Error().Err(err).Msg("Write release data")
		return
	}
}

func (g *GithubCache) cacheAssetFile(release *Release, asset *Asset) error {
	assetPath := g.AssetFilePath(release, asset.Name)
	if _, err := os.Stat(assetPath); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		log.Info().Str("path", assetPath).Msg("Cache asset exist")
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(assetPath), os.ModePerm); err != nil {
		return err
	}

	tempPath := g.AssetFilePath(release, fmt.Sprintf("%s.tmp", asset.Name))
	finalURL := strings.Replace(asset.URL, "https://api.github.com/", fmt.Sprintf("https://%s@api.github.com/", g.conf.Token), 1)
	log.Info().Str("url", finalURL).Str("to", assetPath).Str("name", asset.Name).Str("size", fmt.Sprintf("%dM", asset.Size)).Msg("Downloading...")

	var b io.Reader
	if os.Getenv("MODE") == "TESTING" {
		b = bytes.NewBuffer([]byte(""))
	} else {
		client := http.Client{}
		req, err := http.NewRequest("GET", finalURL, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Accept", "application/octet-stream")

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		b = resp.Body
	}

	out, err := os.Create(tempPath)
	if err != nil {
		return err
	}
	defer out.Close()

	startAt := time.Now()
	if _, err := io.Copy(out, b); err != nil {
		return err
	}

	if err := os.Rename(tempPath, assetPath); err != nil {
		return err
	}

	log.Info().Str("path", assetPath).Dur("duration", time.Since(startAt)).Msg("Download completed")
	return nil
}

func (g *GithubCache) cacheReleaseList(ctx context.Context, id int64, url string) (string, error) {
	client := g.newClient(ctx)
	rc, redirectURL, err := client.Repositories.DownloadReleaseAsset(ctx, g.conf.Owner, g.conf.Repo, id, nil)
	if err != nil {
		return "", err
	}
	if redirectURL != "" {
		resp, err := http.Get(redirectURL)
		if err != nil {
			return "", err
		}
		rc = resp.Body
	}
	log.Debug().Interface("rc", rc).Str("redirectURL", redirectURL).Msg("cacheReleaseList")
	defer rc.Close()

	bs, err := ioutil.ReadAll(rc)
	if err != nil {
		return "", err
	}

	content := string(bs)
	re := regexp.MustCompile(`[^ ]*\.nupkg`)
	matches := re.FindAllString(content, -1)
	if len(matches) == 0 {
		return "", errors.New("RELEASES content doesn't contain nupkg")
	}

	for _, match := range matches {
		nuPKG := strings.ReplaceAll(url, "RELEASES", match)
		content = strings.ReplaceAll(content, match, nuPKG)
	}

	return content, nil
}

func (g *GithubCache) isOutdated() bool {
	return time.Now().After(g.latestUpdate.Add(time.Minute * 3))
}

func (g *GithubCache) runRefreshLoop() {
	for {
		if g.isOutdated() {
			if err := g.refreshCache(); err != nil {
				log.Error().Err(err).Msg("Refresh cache")
			}
		}

		<-time.After(time.Minute * 1)
	}
}

// LoadCache gets latest asset info.
func (g *GithubCache) LoadCache() *Release {
	g.latestMu.RLock()
	latest := g.latest
	g.latestMu.RUnlock()
	return latest
}
