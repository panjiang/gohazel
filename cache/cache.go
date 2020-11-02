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

func checkLatestYmlPlatform(filename string) string {
	if strings.Contains(filename, "win") || filename == "latest.yml" {
		return "exe"
	}

	if strings.Contains(filename, "mac") {
		return "darwin"
	}
	if strings.Contains(filename, "linux") {
		return "AppImage"
	}

	return ""
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

var latestYmlURLReg = regexp.MustCompile(`url:\s+([\w-.]+)`)

// LatestYml stores asset update info of specific platform.
type LatestYml struct {
	Content            string `json:"content"`
	BrowserDownloadURL string `json:"URL"`
}

// ReplaceURL replace the url in content with proxy url.
func (yml *LatestYml) ReplaceURL(u string) {
	yml.Content = latestYmlURLReg.ReplaceAllString(yml.Content, "url: "+u)
}

// Asset is the released package.
type Asset struct {
	ID                 int64      `json:"id"`
	Name               string     `json:"name"`
	URL                string     `json:"url"`
	BrowserDownloadURL string     `json:"browserDownloadURL"`
	ContentType        string     `json:"contentType"`
	Size               int        `json:"size"`
	Yml                *LatestYml `json:"latestYml"`
}

// Release contains major info of every release record.
type Release struct {
	Version   string            `json:"version"`
	Notes     string            `json:"notes"`
	PubDate   github.Timestamp  `json:"pubDate"`
	Platforms map[string]*Asset `json:"platforms"`
	RELEASES  string            `json:"RELEASES"`
}

// ReleaseData release info data for caching into file.
type ReleaseData struct {
	Release       *Release `json:"release"`
	RepoURL       string   `json:"repoUrl"`
	ProxyDownload bool     `json:"proxyDownload"`
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
	quitCh        chan struct{}
	wg            sync.WaitGroup
	mu            sync.Mutex
	closed        bool
	conf          *GithubConfig
	cacheURLBase  string
	proxyDownload bool
	cacheDir      string
	latest        *Release
	latestMu      sync.RWMutex
	latestUpdate  time.Time
}

// NewGithubCache .
func NewGithubCache(conf *GithubConfig, cacheDir string, proxyDownload bool, cacheURLBase string) *GithubCache {
	g := &GithubCache{
		quitCh:        make(chan struct{}),
		conf:          conf,
		proxyDownload: proxyDownload,
		cacheURLBase:  cacheURLBase,
		cacheDir:      cacheDir,
	}
	log.Info().Str("url", conf.RepoURL()).Bool("private", conf.IsPrivateRepo()).Msg("Github repo")

	g.loadReleaseCache()
	g.wg.Add(1)
	go g.runRefreshLoop()
	return g
}

// Stop the cache services.
func (g *GithubCache) Stop() {
	g.mu.Lock()
	if g.closed {
		g.mu.Unlock()
		return
	}
	g.closed = true
	g.mu.Unlock()
	close(g.quitCh)
	g.wg.Wait()
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
		if *item.Prerelease {
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

	platformYmls := map[string]*LatestYml{}
	for _, asset := range release.Assets {
		if *asset.Name == "RELEASES" {
			log.Debug().Interface("asset", asset).Msg("RELEASES")
			content, err := g.fetchFileRELEASES(ctx, *asset.ID, *asset.BrowserDownloadURL)
			if err != nil {
				return err
			}

			latest.RELEASES = content
			continue
		}

		// latest-[win/mac/linux].yml
		if filepath.Ext(*asset.Name) == ".yml" {
			platform := checkLatestYmlPlatform(*asset.Name)
			if platform == "" {
				continue
			}
			content, err := g.fetchFileLatestYml(ctx, *asset.ID, *asset.BrowserDownloadURL)
			if err != nil {
				return err
			}
			platformYmls[platform] = &LatestYml{
				Content:            content,
				BrowserDownloadURL: *asset.BrowserDownloadURL,
			}
			log.Info().Str("asset", *asset.Name).Str("platform", platform).Msg("Cache latest yml")
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
		if g.proxyDownload {
			if err := g.cacheAssetFile(latest, a); err != nil {
				return err
			}
		}

		latest.Platforms[platform] = a
	}

	// Bind latest yml to asset.
	for platform, asset := range latest.Platforms {
		yml, ok := platformYmls[platform]
		if ok {
			asset.Yml = yml
			// Replace download url in yaml file.
			if g.proxyDownload {
				yml.ReplaceURL(g.AssetFileURL(latest, asset.Name))
			}
		} else {
			log.Error().Str("platform", platform).Msg("No latest yml")
		}
	}

	g.latestMu.Lock()
	g.latest = latest
	g.latestMu.Unlock()

	// Clean old cached assets.
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

	if data.ProxyDownload != g.proxyDownload {
		return
	}

	g.latest = data.Release
	log.Info().Str("version", g.latest.Version).Str("file", filename).Msg("Loaded release data from cache")
}

func (g *GithubCache) cacheReleaseLastest(release *Release) {
	data := &ReleaseData{
		Release:       release,
		RepoURL:       g.conf.RepoURL(),
		ProxyDownload: g.proxyDownload,
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

func (g *GithubCache) fetchAssetContent(ctx context.Context, id int64, url string) (string, error) {
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
	log.Debug().Str("redirectURL", redirectURL).Msg("Fetch asset content")
	defer rc.Close()

	bs, err := ioutil.ReadAll(rc)
	if err != nil {
		return "", err
	}

	return string(bs), nil
}

func (g *GithubCache) fetchFileRELEASES(ctx context.Context, id int64, url string) (string, error) {
	content, err := g.fetchAssetContent(ctx, id, url)
	if err != nil {
		return "", err
	}
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

func (g *GithubCache) fetchFileLatestYml(ctx context.Context, id int64, url string) (string, error) {
	content, err := g.fetchAssetContent(ctx, id, url)
	if err != nil {
		return "", err
	}
	return content, nil
}

func (g *GithubCache) isOutdated() bool {
	return time.Now().After(g.latestUpdate.Add(time.Minute * 3))
}

func (g *GithubCache) runRefreshLoop() {
	defer g.wg.Done()
	for {
		if g.isOutdated() {
			if err := g.refreshCache(); err != nil {
				log.Error().Err(err).Msg("Refresh cache")
			}
		}

		select {
		case <-g.quitCh:
			return
		case <-time.After(time.Minute * 1):
		}
	}
}

// LoadCache gets latest asset info.
func (g *GithubCache) LoadCache() *Release {
	g.latestMu.RLock()
	latest := g.latest
	g.latestMu.RUnlock()
	return latest
}
