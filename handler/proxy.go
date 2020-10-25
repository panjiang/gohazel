package handler

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/panjiang/gohazel/cache"
	"github.com/panjiang/gohazel/pkg/api"
	"github.com/rs/zerolog/log"
)

func (h *Handler) proxyDownload(c *gin.Context, release *cache.Release, asset *cache.Asset) {
	assetPath := h.cache.AssetFilePath(release, asset.Name)
	_, err := os.Stat(assetPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warn().Str("path", assetPath).Msg("Proxy download file lost")
			api.NoContent(c)
			return
		}
		log.Error().Err(err).Msg("Stat asset path")
		return
	}

	location := h.cache.AssetFileURL(release, asset.Name)
	api.Found(c, gin.H{
		"Location": location,
	})
}
