package handler

import (
	"strconv"

	"github.com/avct/uasurfer"
	"github.com/gin-gonic/gin"
	"github.com/panjiang/gohazel/pkg/api"
)

func (h *Handler) download(c *gin.Context, platform string) {
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

	if h.conf.ProxyDownload {
		h.proxyDownload(c, release, asset)
		return
	}

	api.Found(c, gin.H{
		"Location": asset.BrowserDownloadURL,
	})
}

// Download reponses the download location url for
// the platform parsed from user agent.
func (h *Handler) Download(c *gin.Context) {
	userAgent := uasurfer.Parse(c.Request.UserAgent())
	isUpdate, _ := strconv.ParseBool(c.Query("update"))
	osPlatform := userAgent.OS.Platform

	var platform string
	switch osPlatform {
	case uasurfer.PlatformMac:
		if isUpdate {
			platform = "darwin"
		} else {
			platform = "dmg"
		}

	case uasurfer.PlatformWindows:
		platform = "exe"
	default:
		api.NoContent(c)
		return
	}

	h.download(c, platform)
}

// DownloadPlatform get download with specific platform.
func (h *Handler) DownloadPlatform(c *gin.Context) {
	isUpdate, _ := strconv.ParseBool(c.Query("update"))
	platform := c.Param("platform")

	if platform == "mac" && !isUpdate {
		platform = "dmg"
	}

	platform, ok := checkAlias(platform)
	if !ok {
		api.BadRequest(c, "platform", "")
		return
	}

	h.download(c, platform)
}
