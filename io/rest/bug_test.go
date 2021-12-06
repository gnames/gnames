package rest_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/gnames/gnfmt"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/stretchr/testify/assert"
)

const urlTest = "http://:8888/api/v0/"

var bugs = []struct {
	name           string
	matchType      vlib.MatchTypeValue
	matchCanonical string
	desc           string
}{
	{
		name:           "Tillaudsia utriculata",
		matchType:      vlib.Fuzzy,
		matchCanonical: "Tillandsia utriculata",
		desc:           "Misspelling of Tillandsia",
	},
	{
		name:           "Drosohila melanogaster",
		matchType:      vlib.Fuzzy,
		matchCanonical: "Drosophila melanogaster",
		desc:           "Misspelling of Drosophila",
	},
	{
		name:           "Acacia nur",
		matchType:      vlib.PartialExact,
		matchCanonical: "Acacia",
		desc:           "Should not match 'Acacia dura', ep. is too short",
	},
	{
		name:           "Bubo",
		matchType:      vlib.Exact,
		matchCanonical: "Bubo",
		desc:           "Uninomials should match correctly #32 gnmatcher",
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
		assert.Equal(t, verif.Names[i].BestResult.MatchedCanonicalSimple, v.matchCanonical, msg)
		assert.Equal(t, verif.Names[i].MatchType, v.matchType, msg)
	}
}

func TestSortBugs(t *testing.T) {
	tests := []struct {
		msg, name, matchName string
	}{
		{"Willd", "Trichomanes bifidum Willd", "Trichomanes bifidum Vent. ex Willd."},
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
		assert.Equal(t, verif.Names[i].BestResult.MatchedName, v.matchName, v.msg)
	}

}

func params() vlib.Input {
	ns := make([]string, len(bugs))
	for i, v := range bugs {
		ns[i] = v.name
	}
	return vlib.Input{NameStrings: ns}
}
