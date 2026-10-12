package matcher_test

import (
	"os"
	"testing"

	"github.com/gnames/gnames/internal/io/matcher"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/stretchr/testify/assert"
)

// url is a standalone gnmatcher service for REST client tests. It is
// separate from GN_MATCHER_URL, so tests never reach a service by
// accident. The service should use the gntest database.
var url = os.Getenv("GN_TEST_MATCHER_URL")

func TestVer(t *testing.T) {
	if url == "" {
		t.Skip("GN_TEST_MATCHER_URL not set, skipping REST test")
	}
	m := matcher.NewREST(url)
	ver := m.GetVersion()
	assert.Regexp(t, `^v\d+\.\d+\.\d+`, ver.Version)
}

func TestMatch(t *testing.T) {
	if url == "" {
		t.Skip("GN_TEST_MATCHER_URL not set, skipping REST test")
	}
	m := matcher.NewREST(url)
	res := m.MatchNames([]string{"Pardosa moeste"})
	assert.Equal(t, "Pardosa moeste", res.Matches[0].Name)
	assert.Equal(t, vlib.Fuzzy, res.Matches[0].MatchType)
}
