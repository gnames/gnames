package matcher_test

import (
	"testing"

	"github.com/gnames/gnames/internal/io/matcher"
	"github.com/gnames/gnames/pkg/config"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/stretchr/testify/assert"
)

func getConfig() config.Config {
	cfg := config.New()
	config.LoadEnv(&cfg)
	return cfg
}

var url = getConfig().MatcherURL

func TestVer(t *testing.T) {
	m := matcher.New(url)
	ver := m.GetVersion()
	assert.Regexp(t, `^v\d+\.\d+\.\d+`, ver.Version)
}

func TestMatch(t *testing.T) {
	m := matcher.New(url)
	res := m.MatchNames([]string{"Pardosa moeste"})
	assert.Equal(t, "Pardosa moeste", res.Matches[0].Name)
	assert.Equal(t, vlib.Fuzzy, res.Matches[0].MatchType)
}
