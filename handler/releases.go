package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/panjiang/gohazel/pkg/api"
)

// Releases responses Releases text.
func (h *Handler) Releases(c *gin.Context) {
	if c.Param("platform") != "win32" {
		api.NotFound(c)
		return
	}

	release := h.cache.LoadCache()
	if release == nil {
		api.NoContent(c)
		return
	}

	if release.RELEASES == "" {
		api.NoContent(c)
		return
	}

	b := []byte(release.RELEASES)
	c.Data(http.StatusOK, "application/octet-stream", b)
}
