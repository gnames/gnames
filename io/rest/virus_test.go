package rest_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/gnames/gnfmt"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/stretchr/testify/assert"
)

func TestVirus(t *testing.T) {
	tests := []struct {
		msg, name, matchStr string
		matchType           vlib.MatchTypeValue
		matchlen            int
	}{
		{
			msg:       "not virus",
			name:      "Something not a virus",
			matchType: vlib.NoMatch,
			matchlen:  0,
		},
		{
			msg:       "arct vir",
			name:      "Antarctic virus",
			matchStr:  "Antarctic virus 1_I_CPGEORsw001Ad",
			matchType: vlib.Virus,
			matchlen:  21,
		},
		{
			msg:       "bird",
			name:      "Bubo bubo",
			matchStr:  "Bubo bubo",
			matchType: vlib.Exact,
			matchlen:  1,
		},
		{
			msg:       "vector",
			name:      "Cloning vector pAJM.011",
			matchStr:  "Cloning vector pAJM.011",
			matchType: vlib.Virus,
			matchlen:  1,
		},
		{
			msg:       "tobacco mosaic",
			name:      "Tobacco mosaic virus",
			matchStr:  "Tobacco mosaic virus",
			matchType: vlib.Virus,
			matchlen:  14,
		},
		{
			msg:       "influenza overload",
			name:      "Influenza B virus",
			matchStr:  "Influenza B virus",
			matchType: vlib.Virus,
			matchlen:  21,
		},
	}

	names := make([]string, len(tests))
	for i := range tests {
		names[i] = tests[i].name
	}
	request := vlib.Input{NameStrings: names}
	var response vlib.Output
	req, err := gnfmt.GNjson{}.Encode(request)
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(restURL+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	assert.Equal(t, len(names), len(response.Names))

	for i, v := range tests {
		ns := response.Names
		assert.Equal(t, v.name, ns[i].Name)
    assert.Equal(t, v.matchType, ns[i].MatchType)
	}
	assert.True(t, true)
}
