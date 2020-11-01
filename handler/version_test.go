package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/mod/semver"
)

func Test(t *testing.T) {
	assert.True(t, semver.IsValid("v1.0.0"))
	assert.True(t, semver.IsValid("v1.0"))
	assert.True(t, semver.IsValid(ToSemver("1.0.0")))
	assert.True(t, semver.IsValid(ToSemver("1.0.0-test1")))
}
