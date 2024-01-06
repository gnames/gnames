package rest_test

import (
	"io"
	"net/http"
	"testing"

	"github.com/gnames/gnfmt"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/stretchr/testify/assert"
)

func TestNameStringsID(t *testing.T) {
	var response vlib.NameStringOutput
	id := "0eeccd70-eaf2-5c51-ad8b-46cfb3db1645"
	assert := assert.New(t)
	resp, err := http.Get(restURL +
		"name_strings/" + id)
	assert.Nil(err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(err)
	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(err)
	assert.Equal(id, response.NameStringMeta.ID)
	assert.NotNil(response.Name)
	assert.Equal(id, response.Name.ID)
	assert.NotNil(response.Name.BestResult)
	assert.Equal(1, response.Name.BestResult.DataSourceID)
}

func TestNameStrings(t *testing.T) {
	var response vlib.NameStringOutput
	id := "0eeccd70-eaf2-5c51-ad8b-46cfb3db1645"
	name := "Bubo bubo (Linnaeus, 1758)"
	assert := assert.New(t)
	resp, err := http.Get(restURL +
		"name_strings/" + name)
	assert.Nil(err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(err)
	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(err)
	assert.Equal(id, response.NameStringMeta.ID)
	assert.NotNil(response.Name)
	assert.Equal(id, response.Name.ID)
	assert.Equal(vlib.Exact, response.MatchType)
	assert.Equal(name, response.Name.Name)
	assert.NotNil(response.Name.BestResult)
	assert.Equal(1, response.Name.BestResult.DataSourceID)
}

func TestNameStringsVirusID(t *testing.T) {
	var response vlib.NameStringOutput
	id := "237d7244-d8b3-5c32-9d91-65e03a4ca78f"
	assert := assert.New(t)
	resp, err := http.Get(restURL +
		"name_strings/" + id)
	assert.Nil(err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(err)
	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(err)
	assert.Equal(id, response.NameStringMeta.ID)
	assert.NotNil(response.Name)
	assert.Equal(id, response.Name.ID)
	assert.Equal("Tobacco mosaic virus", response.Name.Name)
	assert.Equal(vlib.Virus, response.MatchType)
	assert.NotNil(response.Name.BestResult)
	assert.Equal(1, response.Name.BestResult.DataSourceID)
}
func TestNameStringsIDBad(t *testing.T) {
	var response vlib.NameStringOutput
	id := "1eeccd70-eaf2-5c51-ad8b-46cfb3db1645"
	assert := assert.New(t)
	resp, err := http.Get(restURL +
		"name_strings/" + id)
	assert.Nil(err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(err)
	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(err)
	assert.Equal(id, response.NameStringMeta.ID)
	assert.Nil(response.Name)
}

func TestNameStringsBad(t *testing.T) {
	var response vlib.NameStringOutput
	id := "7c5f6d7b-0abd-53a8-8ccd-78a7d32c5e0f"
	name := "Vubo bubo (Linnaeus, 1758"
	assert := assert.New(t)
	resp, err := http.Get(restURL +
		"name_strings/" + name)
	assert.Nil(err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(err)
	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(err)
	assert.Equal(id, response.NameStringMeta.ID)
	assert.Nil(response.Name)
}
