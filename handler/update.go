package handler

import (
	"net/url"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/panjiang/gohazel/pkg/api"
	"golang.org/x/mod/semver"
)

// Update handles checking update request.
func (h *Handler) Update(c *gin.Context) {
	platform := c.Param("platform")
	version := c.Param("version")
	if !semver.IsValid(version) {
		api.BadRequest(c, "version", "is not SemVer-compatible")
		return
	}

	platform, ok := checkAlias(platform)
	if !ok {
		api.BadRequest(c, "platform", "")
		return
	}

	release := h.cache.LoadCache()
	if release == nil {
		api.NoContent(c)
		return
	}

	asset, ok := release.Platforms[platform]
	if !ok {
		api.NoContent(c)
		return
	}

	if semver.Compare(release.Version, version) == 0 {
		api.NoContent(c)
		return
	}

	var downloadURL string
	if h.conf.OpenProxyDownload {
		u, _ := url.Parse(h.conf.BaseURL)
		u.Path = path.Join(u.Path, "download", platform)
		q := u.Query()
		q.Add("update", "true")
		u.RawQuery = q.Encode()
		downloadURL = u.String()
	} else {
		downloadURL = asset.BrowserDownloadURL
	}

	api.Ok(c, gin.H{
		"name":     release.Version,
		"notes":    release.Notes,
		"pub_data": release.PubDate,
		"url":      downloadURL,
	})
	return
}
