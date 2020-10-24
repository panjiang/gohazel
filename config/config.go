package config

import (
	"errors"
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
	Bind              string              `yaml:"bind"`
	Debug             bool                `yaml:"debug"`
	DebugGin          bool                `yaml:"debugGin"`
	BaseURL           string              `yaml:"baseURL"`
	CacheDir          string              `yaml:"cacheDir"`
	OpenProxyDownload bool                `yaml:"proxyDownload"`
	Github            *cache.GithubConfig `yaml:"github"`
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

func (c *Config) validate() error {
	if c.Github == nil {
		return errors.New("no github config")
	}

	if _, err := os.Stat(c.CacheDir); err != nil {
		return err
	}

	if _, err := url.Parse(c.BaseURL); err != nil {
		return err
	}

	if c.Github.IsPrivateRepo() && !c.OpenProxyDownload {
		return errors.New("Private repo should open proxyDownload")
	}

	return nil
}

// Parse config instance from yaml file
func Parse(filename string) (*Config, error) {
	log.Info().Str("filename", filename).Msg("Read config")
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var conf Config
	if err := yaml.Unmarshal(b, &conf); err != nil {
		return nil, err
	}

	if err := conf.validate(); err != nil {
		return nil, err
	}
	return &conf, nil
}
