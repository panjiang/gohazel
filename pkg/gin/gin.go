package gin

import (
	"os"

	"github.com/rs/zerolog/log"

	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
)

func init() {
	if os.Getenv(gin.EnvGinMode) == "" {
		gin.SetMode(gin.ReleaseMode)
	}
}

// New returns gin engine with custom settings.
func New(debug bool) *gin.Engine {
	r := gin.New()
	if debug {
		gin.SetMode(gin.DebugMode)
		r.Use(ginLogger())
	}
	r.Use(gin.Recovery())
	return r
}

func ginLogger() gin.HandlerFunc {
	subLog := log.Logger.With().Str("mod", "gin").Logger()
	return logger.SetLogger(logger.Config{
		Logger: &subLog,
	})
}
