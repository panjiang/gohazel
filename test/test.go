package test

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/panjiang/gohazel/cache"
	"github.com/panjiang/gohazel/config"
	"github.com/panjiang/gohazel/pkg/logger"
	"github.com/panjiang/gohazel/server"
)

// DefaultConfig .
func DefaultConfig() *config.Config {
	return &config.Config{
		Addr:          ":18080",
		BaseURL:       "http://localhost:18080",
		CacheDir:      "/tmp/assets",
		Debug:         false,
		ProxyDownload: false,
		Github: cache.GithubConfig{
			Owner: "atom",
			Repo:  "atom",
			Token: "",
		},
	}
}

// RunServer startups a test server.
func RunServer(conf *config.Config) *server.Server {
	os.Setenv("MODE", "TESTING")
	logger.Setup(conf.Debug)
	s := server.NewServer(conf)
	go s.Run()
	return s
}

// Request send HTTP request to server.
func Request(baseURL string, uri string) (int, []byte) {
	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}
	u.Path = path.Join(u.Path, uri)
	resp, err := http.Get(u.String())
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			<-time.After(time.Second)
			resp, err = http.Get(u.String())
			if err != nil {
				panic(err)
			}
		}
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return resp.StatusCode, b
}
