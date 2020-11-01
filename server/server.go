package server

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/panjiang/gohazel/cache"
	"github.com/panjiang/gohazel/config"
	"github.com/panjiang/gohazel/handler"
	ginpkg "github.com/panjiang/gohazel/pkg/gin"
	"github.com/rs/zerolog/log"
)

// Server is the main service.
type Server struct {
	conf     *config.Config
	cache    *cache.GithubCache
	engine   *gin.Engine
	srv      *http.Server
	shutdown bool
	mu       sync.Mutex
}

// Run up the server
func (s *Server) Run() error {
	srv := &http.Server{
		Addr:    s.conf.Addr,
		Handler: s.engine,
	}
	s.mu.Lock()
	s.srv = srv
	s.mu.Unlock()
	log.Info().Str("addr", s.conf.Addr).Msg("Server run")
	return srv.ListenAndServe()
}

// Shutdown the server gracefully
func (s *Server) Shutdown() {
	s.mu.Lock()
	if s.shutdown {
		s.mu.Unlock()
		return
	}
	s.shutdown = true

	if s.srv != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.srv.Shutdown(ctx)
		s.srv = nil
	}

	if s.cache != nil {
		s.cache.Stop()
		s.cache = nil
	}
	s.mu.Unlock()
}

// NewServer will setup a new server with specific config.
func NewServer(conf *config.Config) *Server {
	// Router
	r := ginpkg.New(conf.Debug)
	r.Use()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	logev := log.Info().Bool("open", conf.ProxyDownload)
	if conf.ProxyDownload {
		r.Static(conf.CacheURLPath(), conf.CacheDir)
		logev.Str("url", conf.CacheURLBase())
	}
	logev.Msg("Proxy download")

	// Cache
	cache := cache.NewGithubCache(&conf.Github, conf.CacheDir, conf.ProxyDownload, conf.CacheURLBase())

	// Handler
	h := handler.NewHandler(conf, cache)
	r.GET("/", h.Overview)
	r.GET("/download", h.Download)
	r.GET("/download/:platform", h.DownloadPlatform)
	r.GET("/update/:platform/:version", h.Update)
	r.GET("/update/:platform/:version/RELEASES", h.Releases) // `/update/win32/:version/RELEASES`
	r.GET("/update/:platform/:version/latest.yml", h.UpdateLatestYml)
	return &Server{
		conf:   conf,
		cache:  cache,
		engine: r,
	}
}
