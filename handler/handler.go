package handler

import (
	"github.com/panjiang/gohazel/cache"
	"github.com/panjiang/gohazel/config"
)

var aliases = map[string][]string{
	"darwin":   {"mac", "macos", "osx"},
	"exe":      {"win32", "windows", "win"},
	"deb":      {"debian"},
	"rpm":      {"fedora"},
	"AppImage": {"appimage"},
	"dmg":      {"dmg"},
}

func checkAlias(platform string) (string, bool) {
	_, ok := aliases[platform]
	if ok {
		return platform, true
	}

	for k, v := range aliases {
		for _, a := range v {
			if a == platform {
				return k, true
			}
		}
	}
	return "", false
}

// Handler handles requests of clients.
type Handler struct {
	cache *cache.GithubCache
	conf  *config.Config
}

// NewHandler returns a handler instance.
func NewHandler(conf *config.Config, cache *cache.GithubCache) *Handler {
	return &Handler{
		conf:  conf,
		cache: cache,
	}
}
