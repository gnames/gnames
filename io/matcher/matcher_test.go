package matcher_test

import (
	"testing"

	"github.com/gnames/gnames/io/matcher"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/stretchr/testify/assert"
)

const url = "https://matcher.globalnames.org/api/v1/"

func TestVer(t *testing.T) {
	m := matcher.New(url)
	ver := m.GetVersion()
	assert.Regexp(t, `^v\d+\.\d+\.\d+`, ver.Version)
}

func TestMatch(t *testing.T) {
	m := matcher.New(url)
	res := m.MatchNames([]string{"Pardosa moeste"})
	assert.Equal(t, "Pardosa moeste", res[0].Name)
	assert.Equal(t, vlib.Fuzzy, res[0].MatchType)
}
