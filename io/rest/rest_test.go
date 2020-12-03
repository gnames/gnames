package rest_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/gnames/gnlib/domain/entity/gn"
	vlib "github.com/gnames/gnlib/domain/entity/verifier"
	"github.com/gnames/gnlib/encode"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	// log "github.com/sirupsen/logrus"
)

const url = "http://:8888/api/v1/"

func TestPing(t *testing.T) {
	resp, err := http.Get(url + "ping")
	assert.Nil(t, err)

	response, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)

	assert.Equal(t, string(response), "pong")
}

func TestVer(t *testing.T) {
	resp, err := http.Get(url + "version")
	assert.Nil(t, err)
	respBytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)

	enc := encode.GNjson{}
	var response gn.Version
	err = enc.Decode(respBytes, &response)
	assert.Nil(t, err)
	assert.Regexp(t, `^v\d+\.\d+\.\d+`, response.Version)
}

func TestVerifyExact(t *testing.T) {
	var response []vlib.Verification
	names := []string{
		"Not name", "Bubo bubo", "Pomatomus",
		"Pardosa moesta", "Plantago major var major",
		"Cytospora ribis mitovirus 2",
		"A-shaped rods", "Alb. alba",
		"Pisonia grandis", "Acacia vestita may",
	}
	request := vlib.VerifyParams{NameStrings: names}
	req, err := encode.GNjson{}.Encode(request)
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(url+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	err = encode.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	assert.Equal(t, len(response), len(names))

	bad := response[0]
	assert.Equal(t, bad.InputID, "82dbfb99-fe6c-5882-99f2-17c7d3955599")
	assert.Equal(t, bad.Input, "Not name")
	assert.Equal(t, bad.MatchType, vlib.NoMatch)
	assert.Nil(t, bad.BestResult)
	assert.Equal(t, bad.DataSourcesNum, 0)
	assert.Equal(t, bad.Curation, vlib.NotCurated)
	assert.Equal(t, bad.Error, "")

	binom := response[1]
	fmt.Printf("bubo: %+v\n", binom)
	fmt.Printf("buboBest: %+v\n", binom.BestResult)
	assert.Equal(t, binom.InputID, "4431a0f3-e901-519a-886f-9b97e0c99d8e")
	assert.Equal(t, binom.Input, "Bubo bubo")
	assert.NotNil(t, binom.BestResult)
	assert.Equal(t, binom.BestResult.DataSourceID, 1)
	assert.Equal(t, binom.BestResult.MatchType, vlib.Exact)
	assert.Equal(t, binom.Curation, vlib.Curated)
	assert.Equal(t, binom.Error, "")

	acceptFilter := response[8]
	assert.Equal(t, acceptFilter.InputID, "4c8848f2-7271-588c-ba81-e4d5efcc1e92")
	assert.Equal(t, acceptFilter.Input, "Pisonia grandis")
	assert.Equal(t, acceptFilter.BestResult.DataSourceID, 1)
	assert.Equal(t, acceptFilter.BestResult.MatchType, vlib.Exact)
	assert.Equal(t, acceptFilter.BestResult.CurrentCanonicalSimple, "Pisonia grandis")

	partial := response[9]
	assert.Equal(t, partial.InputID, "0f84ed48-3a57-59ac-ac1a-2e9221439fdc")
	assert.Equal(t, partial.Input, "Acacia vestita may")
	assert.Equal(t, partial.BestResult.DataSourceID, 1)
	assert.Equal(t, partial.MatchType, vlib.PartialExact)
	assert.Equal(t, partial.BestResult.CurrentCanonicalSimple, "Acacia vestita")
}

func TestFuzzy(t *testing.T) {
	var response []vlib.Verification
	names := []string{
		"Abras precatorius",
	}
	request := vlib.VerifyParams{NameStrings: names, PreferredSources: []int{1, 12, 169, 182}}
	req, err := encode.GNjson{}.Encode(request)
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(url+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	err = encode.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	assert.Equal(t, len(response), len(names))

	fuz1 := response[0]
	assert.Equal(t, fuz1.Input, "Abras precatorius")
	assert.Equal(t, fuz1.BestResult.EditDistance, 1)
}

// TestPrefDS checks if prefferred data sources works correclty.
func TestPrefDS(t *testing.T) {
	var response []vlib.Verification
	names := []string{
		"Bubo bubo", "Pomatomus",
		"Pardosa moesta", "Plantago major var major",
		"Cytospora ribis mitovirus 2",
		"Pisonia grandis",
	}
	request := vlib.VerifyParams{NameStrings: names, PreferredSources: []int{1, 12, 169, 182}}
	req, err := encode.GNjson{}.Encode(request)
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(url+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	err = encode.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	assert.Equal(t, len(response), len(names))

	binom := response[0]
	assert.Equal(t, binom.InputID, "4431a0f3-e901-519a-886f-9b97e0c99d8e")
	assert.Equal(t, binom.Input, "Bubo bubo")
	assert.NotNil(t, binom.BestResult)
	assert.Equal(t, binom.BestResult.DataSourceID, 1)
	assert.Equal(t, binom.BestResult.MatchType, vlib.Exact)
	assert.Equal(t, binom.Curation, vlib.Curated)
	assert.Equal(t, len(binom.PreferredResults), 3)
	assert.Equal(t, binom.Error, "")

	acceptFilter := response[5]
	assert.Equal(t, acceptFilter.InputID, "4c8848f2-7271-588c-ba81-e4d5efcc1e92")
	assert.Equal(t, acceptFilter.Input, "Pisonia grandis")
	assert.Equal(t, acceptFilter.BestResult.DataSourceID, 1)
	assert.Equal(t, acceptFilter.BestResult.MatchType, vlib.Exact)
	assert.Equal(t, acceptFilter.BestResult.CurrentCanonicalSimple, "Pisonia grandis")
	assert.Equal(t, len(binom.PreferredResults), 3)
}

func TestBugs(t *testing.T) {
	var response []vlib.Verification
	names := []string{
		"Aceratagallia fuscosscripta (Oman )",
		"Ampullaria immersa",
		"Abacetine",
	}
	request := vlib.VerifyParams{NameStrings: names}
	req, err := encode.GNjson{}.Encode(request)
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(url+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)
	err = encode.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	assert.Equal(t, len(response), len(names))
}

func TestDataSources(t *testing.T) {
	var response []vlib.DataSource
	resp, err := http.Get(url + "data_sources")
	assert.Nil(t, err)
	respBytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)

	err = encode.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	assert.Greater(t, len(response), 50)
	log.Printf("%+v", response[0].ID)
	col := response[0]
	assert.Equal(t, col.Title, "Catalogue of Life")
}

func TestOneDataSource(t *testing.T) {
	var ds vlib.DataSource
	resp, err := http.Get(url + "data_sources/12")
	assert.Nil(t, err)
	respBytes, err := ioutil.ReadAll(resp.Body)
	assert.Nil(t, err)

	err = encode.GNjson{}.Decode(respBytes, &ds)
	assert.Nil(t, err)
	assert.Equal(t, ds.Title, "Encyclopedia of Life")
	assert.Equal(t, ds.WebsiteURL, "https://eol.org")
	assert.Equal(t, ds.UUID, "dba5f880-a40d-479b-a1ad-a646835edde4")
}
