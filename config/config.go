package config

import (
	"errors"
	"flag"
	"io/ioutil"
	"net/url"
	"os"
	"path"

	"github.com/panjiang/gohazel/cache"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// Config of the server
type Config struct {
	Addr          string             `yaml:"addr"`
	Debug         bool               `yaml:"debug"`
	BaseURL       string             `yaml:"baseURL"`
	CacheDir      string             `yaml:"cacheDir"`
	ProxyDownload bool               `yaml:"proxyDownload"`
	Github        cache.GithubConfig `yaml:"github"`
}

// CacheURLPath the url path of handling cache files.
func (c *Config) CacheURLPath() string {
	return "assets"
}

// CacheURLBase the public base url for cache dir.
func (c *Config) CacheURLBase() string {
	u, _ := url.Parse(c.BaseURL)
	u.Path = path.Join(u.Path, c.CacheURLPath())
	return u.String()
}

// Validate some config items.
func (c *Config) Validate() error {
	if c.Github.Owner == "" || c.Github.Repo == "" {
		return errors.New("no github config")
	}

	if _, err := os.Stat(c.CacheDir); err != nil {
		return err
	}

	if _, err := url.Parse(c.BaseURL); err != nil {
		return err
	}

	if c.Github.IsPrivateRepo() && !c.ProxyDownload {
		return errors.New("private repo should open proxyDownload")
	}

	return nil
}

// ParseFile parses config instance from yaml file
func (c *Config) ParseFile(filename string) error {
	log.Info().Str("filename", filename).Msg("Read config")
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, c)
}

// Parse parses config from flags and file
func Parse(fs *flag.FlagSet, args []string) (*Config, error) {
	conf := &Config{}
	var configFile string
	fs.StringVar(&conf.Addr, "addr", ":8400", "Server listen address.")
	fs.StringVar(&conf.BaseURL, "base_url", "http://localhost:8400", "The server base URL.")
	fs.BoolVar(&conf.Debug, "debug", false, "Open log debug level.")
	fs.StringVar(&conf.CacheDir, "cache_dir", "/assets", "Cache files store in.")
	fs.BoolVar(&conf.ProxyDownload, "proxy_download", false, "Proxy assets download with the server.")
	fs.StringVar(&conf.Github.Owner, "github_owner", "atom", "Gihtub owner name.")
	fs.StringVar(&conf.Github.Repo, "github_repo", "atom", "Github repository name.")
	fs.StringVar(&conf.Github.Token, "github_token", "", "Github api token for private repo.")
	fs.StringVar(&configFile, "config", "", "Configuration file.")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if configFile != "" {
		if err := conf.ParseFile(configFile); err != nil {
			return nil, err
		}
	}

	if err := conf.Validate(); err != nil {
		return nil, err
	}
	if conf.Debug {
		log.Info().Msg("Open debug log")
	}

	return conf, nil
}
