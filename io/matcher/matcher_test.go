package matcher_test

import (
	"testing"

	"github.com/gnames/gnames/io/matcher"
	vlib "github.com/gnames/gnlib/domain/entity/verifier"
	"github.com/gnames/gnmatcher"
	"github.com/stretchr/testify/assert"
)

const url = "https://matcher.globalnames.org/api/v1/"

func TestVer(t *testing.T) {
	var m gnmatcher.GNMatcher
	m = matcher.NewGNMatcher(url)
	ver := m.GetVersion()
	assert.Regexp(t, `^v\d+\.\d+\.\d+`, ver.Version)
}

func TestMatch(t *testing.T) {
	var m gnmatcher.GNMatcher
	m = matcher.NewGNMatcher(url)
	res := m.MatchNames([]string{"Pardosa moeste"})
	assert.Equal(t, res[0].Name, "Pardosa moeste")
	assert.Equal(t, res[0].MatchType, vlib.Fuzzy)
}
