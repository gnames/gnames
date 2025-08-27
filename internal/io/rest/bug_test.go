package rest_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"slices"
	"testing"

	"github.com/gnames/gnfmt"
	"github.com/gnames/gnlib/ent/verifier"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/stretchr/testify/assert"
)

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
	{
		name:           "Diptera",
		matchType:      vlib.Exact,
		matchCanonical: "Diptera",
		matchCurrent:   "Diptera",
		desc:           "Synonyms should be sorted down #94 gnmatcher",
	},
	{
		name:           "Phegopteris",
		matchType:      vlib.Exact,
		matchCanonical: "Phegopteris",
		matchCurrent:   "Phegopteris",
		desc:           "Name should not break search #123 gnverifier",
	},
	{
		name:           "Oecetis complex",
		matchType:      vlib.Exact,
		matchCanonical: "Oecetis complex",
		matchCurrent:   "Oecetis complex",
		desc:           "Ambiguous name #132 gnames",
	},
	{
		name:           "Canis lupus",
		matchType:      vlib.Exact,
		matchCanonical: "Canis lupus",
		matchCurrent:   "Canis lupus",
		desc:           "Not a synonym #122 gnverifier",
	},
}

func TestMoreBugs(t *testing.T) {
	input := params()
	req, err := gnfmt.GNjson{}.Encode(input)
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(restURL+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	var verif vlib.Output
	err = gnfmt.GNjson{}.Decode(respBytes, &verif)
	assert.Nil(t, err)

	for i, v := range bugs {
		msg := fmt.Sprintf("%s -> %s", v.name, v.matchCanonical)
		assert.Equal(t, v.matchCanonical, verif.Names[i].BestResult.MatchedCanonicalSimple, msg)
		assert.Equal(t, v.matchCurrent, verif.Names[i].BestResult.CurrentCanonicalSimple, msg)
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
	resp, err := http.Post(restURL+"verifications", "application/json", r)
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
	resp, err := http.Post(restURL+"verifications", "application/json", r)
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
	slices.Sort(ary)
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
	resp, err := http.Post(restURL+"verifications", "application/json", r)
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

// issue #130: VASCAN should have classificationIDs
func TestVascanClassificationIDs(t *testing.T) {
	inp := vlib.Input{
		NameStrings:    []string{"Acer saccharum"},
		DataSources:    []int{147},
		WithAllMatches: false,
	}
	req, err := gnfmt.GNjson{}.Encode(inp)
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(restURL+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	var verif vlib.Output
	err = gnfmt.GNjson{}.Decode(respBytes, &verif)
	assert.Nil(t, err)

	res := verif.Names[0].BestResult
	assert.Contains(t, res.ClassificationIDs, "|")
}

// issue #131: Diptera results order in CoL.
// We want the first result to be the Diptera order, and the
// second to be a synonym.
func TestDipteraCoL(t *testing.T) {
	assert := assert.New(t)
	inp := vlib.Input{
		NameStrings:    []string{"Diptera"},
		DataSources:    []int{1},
		WithAllMatches: true,
	}
	req, err := gnfmt.GNjson{}.Encode(inp)
	assert.Nil(err)
	r := bytes.NewReader(req)
	resp, err := http.Post(restURL+"verifications", "application/json", r)
	assert.Nil(err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(err)

	var verif vlib.Output
	err = gnfmt.GNjson{}.Decode(respBytes, &verif)
	assert.Nil(err)

	res := verif.Names[0].Results
	// should be at least 2 results
	assert.Greater(len(res), 1)

	res1, res2 := res[0], res[1]
	assert.Equal(verifier.AcceptedTaxStatus, res1.TaxonomicStatus, "res1")
	assert.Equal(verifier.SynonymTaxStatus, res2.TaxonomicStatus, "res2")
}

// issue gnverifier #131: genus in hierarchy bread crumbs returns
// author (Linnaeus) instead of genus (Bistorta) for Bistorta vivipara.
func TestGenusVascan(t *testing.T) {
	assert := assert.New(t)
	inp := vlib.Input{
		NameStrings: []string{"Bistorta vivipara"},
		DataSources: []int{147},
	}
	req, err := gnfmt.GNjson{}.Encode(inp)
	assert.Nil(err)
	r := bytes.NewReader(req)
	resp, err := http.Post(restURL+"verifications", "application/json", r)
	assert.Nil(err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(err)

	var verif vlib.Output
	err = gnfmt.GNjson{}.Decode(respBytes, &verif)
	assert.Nil(err)

	res := verif.Names[0].BestResult
	assert.NotContains(res.ClassificationPath, "Linnaeus")
	assert.Contains(res.ClassificationPath, "|Bistorta|")
}

// issue #143: Searching WFO for `Beta corolliflora` returns bare name
// as Best Result and marks it as Accepted.
func TestBetaCorollifloraWFO(t *testing.T) {
	assert := assert.New(t)
	inp := vlib.Input{
		NameStrings:    []string{"Beta corolliflora"},
		DataSources:    []int{196},
		WithAllMatches: true,
	}
	req, err := gnfmt.GNjson{}.Encode(inp)
	assert.Nil(err)
	r := bytes.NewReader(req)
	resp, err := http.Post(restURL+"verifications", "application/json", r)
	assert.Nil(err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(err)

	var verif vlib.Output
	err = gnfmt.GNjson{}.Decode(respBytes, &verif)
	assert.Nil(err)

	res := verif.Names[0].Results
	// should be at least 2 results
	assert.Greater(len(res), 1)

	res1, res2 := res[0], res[1]
	assert.Equal(verifier.AcceptedTaxStatus, res1.TaxonomicStatus, "res1")
	assert.Equal("Beta corolliflora Zosimovic ex Buttler", res1.MatchedName, "res1 name")
	assert.Equal(verifier.UnknownTaxStatus, res2.TaxonomicStatus, "res2")
	assert.Equal("Beta corolliflora Zosimov.", res2.MatchedName, "res2 name")
}

// issue #141: VASCAN lost all its synonyms as a result of a bug in
// `sf from dwca`.
func TestVascanSynonyms(t *testing.T) {
	assert := assert.New(t)
	inp := vlib.Input{
		NameStrings:    []string{"Actaea alba"},
		DataSources:    []int{147},
		WithAllMatches: true,
	}
	req, err := gnfmt.GNjson{}.Encode(inp)
	assert.Nil(err)
	r := bytes.NewReader(req)
	resp, err := http.Post(restURL+"verifications", "application/json", r)
	assert.Nil(err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(err)

	var verif vlib.Output
	err = gnfmt.GNjson{}.Decode(respBytes, &verif)
	assert.Nil(err)

	res := verif.Names[0].Results
	// should be at least 2 results
	assert.Greater(len(res), 0)

	res1 := res[0]
	assert.Equal(verifier.SynonymTaxStatus, res1.TaxonomicStatus, "res1")
	assert.Equal("Actaea alba (Linnaeus) Miller", res1.MatchedName, "res1 name")
	assert.Equal("Actaea pachypoda Elliott", res1.CurrentName,
		"res1 current name")
}

func params() vlib.Input {
	ns := make([]string, len(bugs))
	for i, v := range bugs {
		ns[i] = v.name
	}
	return vlib.Input{NameStrings: ns}
}
