package main

import (
	"flag"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/panjiang/gohazel/config"
	"github.com/panjiang/gohazel/pkg/logger"
	"github.com/panjiang/gohazel/server"
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

	conf, err := config.Parse(flagConf)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	logger.Setup(conf.Debug)

	s := server.NewServer(conf)
	if err := s.Run(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Run server")
	}
	log.Info().Msg("Server closed")
}
