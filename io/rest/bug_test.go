package rest

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	vlib "github.com/gnames/gnlib/domain/entity/verifier"
	"github.com/gnames/gnlib/encode"
	"github.com/stretchr/testify/assert"
)

const url = "http://:8888/api/v1/"

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
}

func TestBugs(t *testing.T) {
	req, err := encode.GNjson{}.Encode(params())
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(url+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)

	var verif []vlib.Verification
	err = encode.GNjson{}.Decode(respBytes, &verif)
	assert.Nil(t, err)

	for i, v := range bugs {
		assert.Equal(t, verif[i].BestResult.MatchedCanonicalSimple, v.matchCanonical)
		assert.Equal(t, verif[i].MatchType, v.matchType)
	}
}

func params() vlib.VerifyParams {
	ns := make([]string, len(bugs))
	for i, v := range bugs {
		ns[i] = v.name
	}
	return vlib.VerifyParams{NameStrings: ns}
}
