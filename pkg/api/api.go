package api

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// BadRequest 400
func BadRequest(c *gin.Context, field string, msg string) {
	if msg == "" {
		msg = fmt.Sprintf("invalid %s", field)
	}
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": msg, "field": field})
}

// NotFound 404
func NotFound(c *gin.Context) {
	c.AbortWithStatus(http.StatusNotFound)
}

// Found 302
func Found(c *gin.Context, data gin.H) {
	c.JSON(http.StatusFound, data)
}

// NoContent 204
func NoContent(c *gin.Context) {
	c.AbortWithStatus(http.StatusNoContent)
}

// Ok 200
func Ok(c *gin.Context, data gin.H) {
	c.JSON(http.StatusOK, data)
}
