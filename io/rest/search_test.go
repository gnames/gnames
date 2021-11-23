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
	query := url.PathEscape("n:Bubo bubo tx:Aves all:t au:Linn.")
	resp, err := http.Get(searchURL + "/" + query)
	assert.Nil(t, err)
	var response search.Output
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	meta := response.Meta
	names := response.Names
	assert.True(t, meta.WithAllResults)
	assert.Equal(t, meta.Author, "Linn.")
	assert.True(t, len(names) > 1)
	assert.True(t, len(names[0].Results) > 5)
}

func TestPostSearch(t *testing.T) {
	inp := gnquery.New().Parse("g:Pomatomus sp:sal. tx:Actinopterygii au:Linn.")
	assert.Equal(t, len(inp.Warnings), 0)
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
	assert.True(t, len(response.Names) > 0)
}
