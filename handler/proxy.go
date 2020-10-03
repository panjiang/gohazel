package handler

import (
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/panjiang/gohazel/cache"
	"github.com/panjiang/gohazel/pkg/api"
	"github.com/rs/zerolog/log"
)

func (h *Handler) shouldProxyPrivateDownload() bool {
	return h.conf.Github.Token != ""
}

func (h *Handler) proxyPrivateDownload(c *gin.Context, asset *cache.Asset) {
	assetPath := filepath.Join(h.conf.AssetsDir, asset.Name)
	_, err := os.Stat(assetPath)
	if err != nil {
		if os.IsNotExist(err) {
			api.NoContent(c)
			return
		}
		log.Error().Err(err).Msg("Stat asset path")
		return
	}

	u, _ := url.Parse(h.conf.BaseURL)
	u.Path = path.Join(u.Path, h.conf.AssetsDir, asset.Name)
	api.Found(c, gin.H{
		"Location": u.String(),
	})
	return
}
