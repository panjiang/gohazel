package test

import (
	"encoding/json"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestOverview(t *testing.T) {
	conf := DefaultConfig()
	s := RunServer(conf)
	defer s.Shutdown()

	code, data := Request(conf.BaseURL, "/")
	if code != 200 {
		t.Errorf("Expected code is 200, got %v", code)
	}
	var h gin.H
	if err := json.Unmarshal(data, &h); err != nil {
		panic(err)
	}
	if h["owner"] != "atom" {
		t.Errorf("Expected owner is atom, got %v", h["owner"])
	}
}
