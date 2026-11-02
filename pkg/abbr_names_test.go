package gnames_test

import (
	"context"
	"strings"
	"testing"

	gnames "github.com/gnames/gnames/pkg"
	"github.com/gnames/gnames/pkg/config"
	"github.com/gnames/gnames/pkg/ent/abbr"
	alib "github.com/gnames/gnlib/ent/abbr"
	"github.com/stretchr/testify/assert"
)

// mockAbbrSizes has every epithet in Monochamus and in Myrmica, and the
// genera with `Mon` have far fewer epithets than genera with `M`.
type mockAbbrSizes struct {
	mockAbbr
	prefixes []string
}

func (ma *mockAbbrSizes) AbbrLookups(
	ctx context.Context,
	keys []string,
) ([]abbr.Lookup, error) {
	var res []abbr.Lookup
	for _, k := range keys {
		stem := k[strings.Index(k, "-")+1:]
		for _, g := range []string{"Monochamus", "Myrmica"} {
			res = append(res, abbr.Lookup{
				Key:             k,
				Genus:           g,
				CanonicalStemID: g + "-" + stem,
				CanonicalStem:   g + " " + stem,
				DsNum:           3,
			})
		}
	}
	return res, nil
}

func (ma *mockAbbrSizes) AbbrGenera(
	ctx context.Context,
	genera []string,
) ([]abbr.GenusRecord, error) {
	var res []abbr.GenusRecord
	for _, g := range genera {
		res = append(res, abbr.GenusRecord{
			Genus:       g,
			Kingdom:     "Animalia",
			EpithetsNum: 455,
			Known:       true,
		})
	}
	return res, nil
}

func (ma *mockAbbrSizes) AbbrPrefixSize(
	ctx context.Context,
	prefix string,
) (int, error) {
	ma.prefixes = append(ma.prefixes, prefix)
	sizes := map[string]int{"M": 2_000_000, "Mon": 20_000}
	return sizes[prefix], nil
}

// TestAbbrNamesLongerPrefix checks that a longer abbreviation never
// scores lower than a shorter one with the same matches.
func TestAbbrNamesLongerPrefix(t *testing.T) {
	assert := assert.New(t)
	ctx := context.Background()
	score := func(a string) (float64, []string) {
		ab := &mockAbbrSizes{}
		g, err := gnames.New(config.New(), mockVerifier{}, mockVernacular{},
			mockFacet{}, ab, gnames.WithMatcher(mockMatcher{}))
		assert.Nil(err)
		var names []alib.Name
		for _, e := range []string{"galloprovincialis", "adamitus", "aparus"} {
			names = append(names, alib.Name{Name: a + " " + e})
		}
		res, err := g.AbbrNames(ctx, alib.Input{Names: names})
		assert.Nil(err)
		assert.Empty(res.Error)
		assert.NotEmpty(res.Genera)
		assert.Equal("Monochamus", res.Genera[0].Name)
		return res.Genera[0].Score.Total, ab.prefixes
	}

	short, prefixes := score("M.")
	assert.Equal([]string{"M"}, prefixes)
	long, prefixes := score("Mon.")
	assert.Equal([]string{"M"}, prefixes)
	assert.Greater(short, 0.0)
	assert.GreaterOrEqual(long, short)
}
