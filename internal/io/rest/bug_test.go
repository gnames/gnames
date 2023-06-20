package rest_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sort"
	"testing"

	"github.com/gnames/gnfmt"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/stretchr/testify/assert"
)

const urlTest = "http://0.0.0.0:8888/api/v1/"

var bugs = []struct {
	name           string
	matchType      vlib.MatchTypeValue
	matchCanonical string
	matchCurrent   string
	desc           string
}{
	{
		name:           "Tillaudsia utriculata",
		matchType:      vlib.Fuzzy,
		matchCanonical: "Tillandsia utriculata",
		matchCurrent:   "Tillandsia utriculata",
		desc:           "Misspelling of Tillandsia",
	},
	{
		name:           "Drosohila melanogaster",
		matchType:      vlib.Fuzzy,
		matchCanonical: "Drosophila melanogaster",
		matchCurrent:   "Drosophila melanogaster",
		desc:           "Misspelling of Drosophila",
	},
	{
		name:           "Acacia nur",
		matchType:      vlib.PartialExact,
		matchCanonical: "Acacia",
		matchCurrent:   "Acacia",
		desc:           "Should not match 'Acacia dura', ep. is too short",
	},
	{
		name:           "Bubo",
		matchType:      vlib.Exact,
		matchCanonical: "Bubo",
		matchCurrent:   "Bubo",
		desc:           "Uninomials should match correctly #32 gnmatcher",
	},
	{
		name:           "Trichoderma",
		matchType:      vlib.Exact,
		matchCanonical: "Trichoderma",
		matchCurrent:   "Trichoderma",
		desc:           "Synonyms should be sorted down #94 gnmatcher",
	},
}

func TestMoreBugs(t *testing.T) {
	req, err := gnfmt.GNjson{}.Encode(params())
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(urlTest+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	var verif vlib.Output
	err = gnfmt.GNjson{}.Decode(respBytes, &verif)
	assert.Nil(t, err)

	for i, v := range bugs {
		msg := fmt.Sprintf("%s -> %s", v.name, v.matchCanonical)
		assert.Equal(t, v.matchCanonical, verif.Names[i].BestResult.MatchedCanonicalSimple, msg)
		assert.Equal(t, v.matchType, verif.Names[i].MatchType, msg)
	}
}

func TestSortBugs(t *testing.T) {
	tests := []struct {
		msg, name, matchName string
	}{
		{"Willd", "Trichomanes bifidum Willd", "Trichomanes bifidum Willd."},
	}

	ns := make([]string, len(tests))
	for i := range tests {
		ns[i] = tests[i].name
	}
	inp := vlib.Input{NameStrings: ns}

	req, err := gnfmt.GNjson{}.Encode(inp)
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(urlTest+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	var verif vlib.Output
	err = gnfmt.GNjson{}.Decode(respBytes, &verif)
	assert.Nil(t, err)

	for i, v := range tests {
		assert.Equal(t, v.matchName, verif.Names[i].BestResult.MatchedName, v.msg)
	}

}

// related to #84
func TestMissedMatchType(t *testing.T) {
	inp := vlib.Input{
		NameStrings:    []string{"Jsoetes cf. longissimum Bory"},
		DataSources:    []int{196, 197, 198, 158},
		WithAllMatches: true,
	}
	req, err := gnfmt.GNjson{}.Encode(inp)
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(urlTest+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	var verif vlib.Output
	err = gnfmt.GNjson{}.Decode(respBytes, &verif)
	assert.Nil(t, err)

	assert.Equal(t, vlib.Fuzzy.String(), verif.Names[0].MatchType.String())
	assert.True(t, len(verif.Names[0].Results) > 0)
	assert.Equal(t, vlib.Fuzzy.String(), verif.Names[0].Results[0].MatchType.String())

	matchedCanonicals := make(map[string]struct{})
	for _, v := range verif.Names[0].Results {
		matchedCanonicals[v.MatchedCanonicalSimple] = struct{}{}
	}
	ary := make([]string, len(matchedCanonicals))
	var i int
	for k := range matchedCanonicals {
		ary[i] = k
		i++
	}
	sort.Strings(ary)
	assert.Equal(t, ary, []string{"Isoetes longissima", "Isoetes longissimum"})
}

// related to #87
func TestWrongMatchType(t *testing.T) {
	inp := vlib.Input{
		NameStrings:    []string{"Jsoetes longissimum"},
		DataSources:    []int{195},
		WithAllMatches: true,
	}
	req, err := gnfmt.GNjson{}.Encode(inp)
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(urlTest+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	var verif vlib.Output
	err = gnfmt.GNjson{}.Decode(respBytes, &verif)
	assert.Nil(t, err)

	assert.Nil(t, verif.Names[0].BestResult)
	assert.Equal(t, 0, len(verif.Names[0].Results))
	assert.Equal(t, vlib.NoMatch.String(), verif.Names[0].MatchType.String())
}

func params() vlib.Input {
	ns := make([]string, len(bugs))
	for i, v := range bugs {
		ns[i] = v.name
	}
	return vlib.Input{NameStrings: ns}
}
