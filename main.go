package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/panjiang/gohazel/config"
	"github.com/panjiang/gohazel/pkg/logger"
	"github.com/panjiang/gohazel/server"
	"github.com/rs/zerolog/log"
)

var usageStr = `Usage: gohazel [options]
Server Options:
    -addr             Server listen address.
    -base_url         The server base URL.
    -debug            Open log debug level.
    -cache_dir        Cache files store in.
    -proxy_download   Proxy assets download with the server.
    -github_owner     Gihtub owner name.
    -github_repo      Github repository name.
    -github_token     Github api token for private repo.
    -config           Or specify a YAML configuration file.
`

func usage() {
	fmt.Println(usageStr)
	os.Exit(0)
}

func main() {
	fs := flag.NewFlagSet("gohazel", flag.ExitOnError)
	fs.Usage = usage

	conf, err := config.Parse(fs, os.Args[1:])
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
