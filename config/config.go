package config

import (
	"io/ioutil"
	"net/url"
	"os"

	"github.com/panjiang/gohazel/cache"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// Config of the server
type Config struct {
	Bind      string              `yaml:"bind"`
	Debug     bool                `yaml:"debug"`
	DebugGin  bool                `yaml:"debugGin"`
	BaseURL   string              `yaml:"baseURL"`
	AssetsDir string              `yaml:"assetsDir"`
	Github    *cache.GithubConfig `yaml:"github"`
}

func (c *Config) validate() error {
	if _, err := url.Parse(c.BaseURL); err != nil {
		return err
	}
	if _, err := os.Stat(c.AssetsDir); err != nil {
		return err
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
