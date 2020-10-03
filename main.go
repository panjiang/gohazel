package main

import (
	"flag"

	"github.com/gin-gonic/gin"
	"github.com/panjiang/gohazel/cache"
	"github.com/panjiang/gohazel/config"
	"github.com/panjiang/gohazel/handler"
	ginpkg "github.com/panjiang/gohazel/pkg/gin"
	"github.com/panjiang/gohazel/pkg/logger"
	"github.com/rs/zerolog/log"
)

var (
	flagConf string
)

func init() {
	flag.StringVar(&flagConf, "conf", "config.yml", "config file in yaml formating")

	gin.SetMode(gin.ReleaseMode)
}

func main() {
	flag.Parse()

	// Config
	conf, err := config.Parse(flagConf)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	// Logger
	logger.Setup(conf.Debug)

	// Router
	r := ginpkg.New(conf.DebugGin)
	r.Use()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Static("/assets", conf.AssetsDir)

	// Cache
	cache := cache.NewGithubCache(conf.Github, conf.AssetsDir)

	// Handler
	h := handler.NewHandler(conf, cache)
	r.GET("/", h.Overview)
	r.GET("/download", h.Download)
	r.GET("/download/:platform", h.DownloadPlatform)
	r.GET("/update/:platform/:version", h.Update)
	r.GET("/update/:platform/:version/RELEASES", h.Releases) // `/update/win32/:version/RELEASES`

	r.Run(conf.Bind)
}
