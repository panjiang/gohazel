package github

import (
	"context"
	"errors"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v32/github"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

type Asset struct {
	ID          int64
	Name        string
	URL         string
	DownloadURL string
	ContentType string
	Size        int
}

var directCache = map[string]struct{}{
	"exe":      {},
	"dmg":      {},
	"rpm":      {},
	"deb":      {},
	"AppImage": {},
}

func checkPlatform(filename string) string {
	extension := filepath.Ext(filename)
	if (strings.Contains(filename, "mac") || strings.Contains(filename, "darwin")) && extension == "zip" {
		return "darwin"
	}

	_, ok := directCache[extension]
	if ok {
		return extension
	}

	return ""
}

type Release struct {
	Version   string
	Notes     string
	PubDate   github.Timestamp
	Platforms map[string]*Asset
	RELEASES  string
}

type Config struct {
	Owner string
	Repo  string
	Token string
	Pre   bool
}

type Github struct {
	conf         *Config
	latest       *Release
	latestUpdate time.Time
}

func (g *Github) newClient(ctx context.Context) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.conf.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	return github.NewClient(tc)
}

func (g *Github) refreshCache() error {
	ctx := context.Background()
	client := g.newClient(ctx)
	releases, _, err := client.Repositories.ListReleases(ctx, g.conf.Owner, g.conf.Repo, &github.ListOptions{
		PerPage: 100,
	})
	if err != nil {
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

	if g.latest != nil && g.latest.Version == *release.TagName {
		log.Info().Msg("Cached version is the same as latest")
		g.latestUpdate = time.Now()
		return nil
	}

	latest := &Release{
		Version: *release.TagName,
		Notes:   *release.Body,
		PubDate: *release.PublishedAt,
	}
	log.Debug().Str("version", latest.Version).Msg("Caching...")
	for _, asset := range release.Assets {
		if *asset.Name == "RELEASES" {
			content, err := g.cacheReleaseList(ctx, *asset.ID, *asset.BrowserDownloadURL)
			if err != nil {
				log.Error().Err(err).Msg("cacheReleaseList")
			}

			g.latest.RELEASES = content
			continue
		}

		platform := checkPlatform(*asset.Name)
		if platform == "" {
			continue
		}

		latest.Platforms[platform] = &Asset{
			ID:          *asset.ID,
			Name:        *asset.Name,
			URL:         *asset.URL,
			DownloadURL: *asset.BrowserDownloadURL,
			ContentType: *asset.ContentType,
			Size:        (*asset.Size) / 1000000 * 10 / 10,
		}
	}

	g.latest = latest
	g.latestUpdate = time.Now()
	log.Debug().Str("version", latest.Version).Msg("Finished caching")
	return nil
}

func (g *Github) cacheReleaseList(ctx context.Context, id int64, url string) (string, error) {
	client := g.newClient(ctx)
	rc, _, err := client.Repositories.DownloadReleaseAsset(ctx, g.conf.Owner, g.conf.Repo, id, nil)
	if err != nil {
		return "", err
	}
	defer rc.Close()

	bs, err := ioutil.ReadAll(rc)
	if err != nil {
		return "", err
	}

	content := string(bs)
	re := regexp.MustCompile(`[^ ]*\.nupkg`)
	matches := re.FindAllString(content, -1)
	if len(matches) == 0 {
		return "", errors.New("Tried to cache RELEASES, but failed. RELEASES content doesn't contain nupkg")
	}

	for _, match := range matches {
		nuPKG := strings.ReplaceAll(url, "RELEASES", match)
		content = strings.ReplaceAll(content, match, nuPKG)
	}

	return content, nil
}

func (g *Github) isOutdated() bool {
	return time.Now().After(g.latestUpdate.Add(time.Minute * 3))
}

func (g *Github) loadCache() *Release {
	if g.isOutdated() {
		g.refreshCache()
	}
	return g.latest
}
