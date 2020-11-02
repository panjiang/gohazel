package cache

import (
	"testing"
)

var content = `version: 1.0.0
files:
  - url: Crownote-Setup-1.0.0.exe
    sha512: iA8kYPnpimEOBx2aM43TMoQIYn6BgktyMmV1Eejz5LbC7qd/booOIcWRwPvYrU/1y475IP5wMjQhbhdH2K4CkA==
    size: 101906819
path: Crownote-Setup-1.0.0.exe
sha512: iA8kYPnpimEOBx2aM43TMoQIYn6BgktyMmV1Eejz5LbC7qd/booOIcWRwPvYrU/1y475IP5wMjQhbhdH2K4CkA==
releaseDate: '2020-11-02T14:14:25.510Z'
`

func TestLatestYml_ReplaceURL(t *testing.T) {
	yml := &LatestYml{
		Content: content,
	}
	yml.ReplaceURL("http://crownote.com:8400/assets/panjiang/crownote/v1.0.1/Crownote-Setup-1.0.1.exe")
	t.Log(yml.Content)
}
