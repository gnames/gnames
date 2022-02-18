package rest_test

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/gnames/gnfmt"
	"github.com/gnames/gnquery"
	"github.com/gnames/gnquery/ent/search"
	"github.com/stretchr/testify/assert"
)

const searchURL = "http://:8888/api/v0/search"

func TestGetSearch(t *testing.T) {
	query := url.PathEscape("n:Bubo bubo tx:Aves ds:1,2 all:t au:Linn.")
	resp, err := http.Get(searchURL + "/" + query)
	assert.Nil(t, err)
	var response search.Output
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	meta := response.Meta
	names := response.Names
	assert.True(t, meta.Input.WithAllResults)
	assert.Equal(t, []int{1, 2}, meta.Input.DataSourceIDs)
	assert.Equal(t, "Linn.", meta.Input.Author)
	assert.True(t, len(names) > 1)
	assert.True(t, len(names[0].Results) > 0)
}

func TestPostSearch(t *testing.T) {
	tests := []struct {
		msg, query string
		hasResults bool
	}{
		{
			msg:        "Pomatomus",
			query:      "g:Pomatomus sp:sal. tx:Actinopterygii au:Linn.",
			hasResults: true,
		},
		{
			msg:        "P. wilsoni Animalia",
			query:      "n:P. wilsoni tx:Animalia",
			hasResults: true,
		},
		{
			msg:        "P. wilsoni",
			query:      "n:P. wilsoni tx:Carnivora",
			hasResults: true,
		},
	}
	for _, v := range tests {
		inp := gnquery.New().Parse(v.query)
		assert.Equal(t, 0, len(inp.Warnings))
		assert.False(t, inp.WithAllResults)
		req, err := gnfmt.GNjson{}.Encode(inp)
		assert.Nil(t, err)
		r := bytes.NewReader(req)
		resp, err := http.Post(searchURL, "application/json", r)
		assert.Nil(t, err)

		var response search.Output
		respBytes, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		err = gnfmt.GNjson{}.Decode(respBytes, &response)
		assert.Nil(t, err)
		assert.Equal(t, len(response.Names) > 0, v.hasResults, v.msg)
	}
}
