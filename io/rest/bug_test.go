package rest

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnfmt"
	"github.com/stretchr/testify/assert"
)

const urlTest = "http://:8888/api/v1/"

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

func TestBugs(t *testing.T) {
	req, err := gnfmt.GNjson{}.Encode(params())
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(urlTest+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)

	var verif []vlib.Verification
	err = gnfmt.GNjson{}.Decode(respBytes, &verif)
	assert.Nil(t, err)

	for i, v := range bugs {
		msg := fmt.Sprintf("%s -> %s", v.name, v.matchCanonical)
		assert.Equal(t, verif[i].BestResult.MatchedCanonicalSimple, v.matchCanonical, msg)
		assert.Equal(t, verif[i].MatchType, v.matchType, msg)
	}
}

func params() vlib.VerifyParams {
	ns := make([]string, len(bugs))
	for i, v := range bugs {
		ns[i] = v.name
	}
	return vlib.VerifyParams{NameStrings: ns}
}
