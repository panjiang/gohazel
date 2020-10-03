package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/panjiang/gohazel/pkg/api"
)

// Overview responses information of the latest version.
func (h *Handler) Overview(c *gin.Context) {
	latest := h.cache.LoadCache()
	if latest == nil {
		api.Ok(c, gin.H{
			"owner": h.conf.Github.Owner,
			"repo":  h.conf.Github.Repo,
		})
		return
	}
	api.Ok(c, gin.H{
		"owner":   h.conf.Github.Owner,
		"repo":    h.conf.Github.Repo,
		"release": latest,
	})
}
