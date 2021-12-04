package rest_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/gnames/gnfmt"
	"github.com/gnames/gnlib/ent/gnvers"
	vlib "github.com/gnames/gnlib/ent/verifier"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	// log "github.com/sirupsen/logrus"
)

const restURL = "http://:8888/api/v0/"

func TestPing(t *testing.T) {
	resp, err := http.Get(restURL + "ping")
	assert.Nil(t, err)

	response, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	assert.Equal(t, string(response), "pong")
}

func TestVer(t *testing.T) {
	resp, err := http.Get(restURL + "version")
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	enc := gnfmt.GNjson{}
	var response gnvers.Version
	err = enc.Decode(respBytes, &response)
	assert.Nil(t, err)
	assert.Regexp(t, `^v\d+\.\d+\.\d+`, response.Version)
}

func TestVerifyExact(t *testing.T) {
	var response vlib.Output
	names := []string{
		"Not name",
		"Bubo bubo",
		"Pomatomus",
		"Pardosa moesta",
		"Plantago major var major",
		"Cytospora ribis mitovirus 2",
		"A-shaped rods",
		"Alb. alba",
		"Pisonia grandis",
		"Acacia vestita may",
		"Candidatus Aenigmarchaeum subterraneum",
	}
	request := vlib.Input{NameStrings: names}
	req, err := gnfmt.GNjson{}.Encode(request)
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(restURL+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	assert.Equal(t, len(response.Names), len(names))

	bad := response.Names[0]
	assert.Equal(t, bad.ID, "82dbfb99-fe6c-5882-99f2-17c7d3955599")
	assert.Equal(t, bad.Name, "Not name")
	assert.Equal(t, bad.MatchType, vlib.NoMatch)
	assert.Nil(t, bad.BestResult)
	assert.Equal(t, bad.DataSourcesNum, 0)
	assert.Equal(t, bad.Curation, vlib.NotCurated)
	assert.Equal(t, bad.Error, "")

	binom := response.Names[1]
	assert.Equal(t, binom.ID, "4431a0f3-e901-519a-886f-9b97e0c99d8e")
	assert.Equal(t, binom.Name, "Bubo bubo")
	assert.NotNil(t, binom.BestResult)
	assert.Equal(t, binom.BestResult.DataSourceID, 1)
	assert.Equal(t, binom.BestResult.MatchType, vlib.Exact)
	assert.Equal(t, binom.Curation, vlib.Curated)
	assert.Equal(t, binom.Error, "")

	acceptFilter := response.Names[8]
	assert.Equal(t, acceptFilter.ID, "4c8848f2-7271-588c-ba81-e4d5efcc1e92")
	assert.Equal(t, acceptFilter.Name, "Pisonia grandis")
	assert.Equal(t, acceptFilter.BestResult.DataSourceID, 1)
	assert.Equal(t, acceptFilter.BestResult.MatchType, vlib.Exact)
	assert.Equal(t, acceptFilter.BestResult.CurrentCanonicalSimple, "Ceodes grandis")

	partial := response.Names[9]
	assert.Equal(t, partial.ID, "0f84ed48-3a57-59ac-ac1a-2e9221439fdc")
	assert.Equal(t, partial.Name, "Acacia vestita may")
	assert.Equal(t, partial.BestResult.DataSourceID, 1)
	assert.Equal(t, partial.MatchType, vlib.PartialExact)
	assert.Equal(t, partial.BestResult.CurrentCanonicalSimple, "Acacia vestita")

	cand := response.Names[10]
	assert.Equal(t, cand.ID, "1b406033-fc5e-5f90-b3cf-fd1e9a42e282")
	assert.Equal(t, cand.Name, "Candidatus Aenigmarchaeum subterraneum")
	assert.NotNil(t, cand.BestResult)
	assert.Equal(t, cand.BestResult.DataSourceID, 179)
	assert.Equal(t, cand.BestResult.MatchType, vlib.Exact)
	assert.Equal(t, cand.Curation, vlib.AutoCurated)
	assert.Equal(t, cand.Error, "")
}

func TestFuzzy(t *testing.T) {
	var response vlib.Output
	names := []string{
		"Abras precatorius",
	}
	request := vlib.Input{NameStrings: names, DataSources: []int{1, 12, 169, 182}}
	req, err := gnfmt.GNjson{}.Encode(request)
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(restURL+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	assert.Equal(t, len(response.Names), len(names))

	fuz1 := response.Names[0]
	assert.Equal(t, fuz1.Name, "Abras precatorius")
	assert.Equal(t, fuz1.BestResult.EditDistance, 1)
}

// TestPrefDS checks if prefferred data sources works correclty.
func TestPrefDS(t *testing.T) {
	var response vlib.Output
	names := []string{
		"Bubo bubo", "Pomatomus",
		"Pardosa moesta", "Plantago major var major",
		"Cytospora ribis mitovirus 2",
		"Pisonia grandis",
	}
	request := vlib.Input{NameStrings: names, DataSources: []int{1, 12, 169, 182}}
	req, err := gnfmt.GNjson{}.Encode(request)
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(restURL+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	assert.Equal(t, len(response.Names), len(names))

	binom := response.Names[0]
	assert.Equal(t, binom.ID, "4431a0f3-e901-519a-886f-9b97e0c99d8e")
	assert.Equal(t, binom.Name, "Bubo bubo")
	assert.NotNil(t, binom.BestResult)
	assert.Equal(t, binom.BestResult.DataSourceID, 1)
	assert.Contains(t, binom.BestResult.Outlink, "NKSD")
	assert.Equal(t, binom.BestResult.MatchType, vlib.Exact)
	assert.Equal(t, binom.Curation, vlib.Curated)
	assert.Equal(t, len(binom.Results), 3)
	assert.Equal(t, binom.Error, "")

	acceptFilter := response.Names[5]
	assert.Equal(t, acceptFilter.ID, "4c8848f2-7271-588c-ba81-e4d5efcc1e92")
	assert.Equal(t, acceptFilter.Name, "Pisonia grandis")
	assert.Equal(t, acceptFilter.BestResult.DataSourceID, 1)
	assert.Equal(t, acceptFilter.BestResult.MatchType, vlib.Exact)
	assert.Equal(t, acceptFilter.BestResult.CurrentCanonicalSimple, "Ceodes grandis")
	assert.Equal(t, len(binom.Results), 3)
}

func TestPrefCapitalize(t *testing.T) {
	var response vlib.Output
	names := []string{
		"bubo bubo", "pomatomus",
		"pardosa moesta", "plantago major var major",
		"cytospora ribis mitovirus 2",
		"pisonia grandis",
	}
	request := vlib.Input{NameStrings: names, WithCapitalization: true}
	req, err := gnfmt.GNjson{}.Encode(request)
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(restURL+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	assert.Equal(t, len(response.Names), len(names))
	assert.True(t, response.WithCapitalization)

	bubo := response.Names[0]
	assert.Equal(t, bubo.ID, "7e4c9a7c-0e90-5d1e-96be-bbea21fcfdd3")
	assert.Equal(t, bubo.Name, "bubo bubo")
	assert.NotNil(t, bubo.BestResult)
	assert.Equal(t, bubo.BestResult.DataSourceID, 1)
	assert.Contains(t, bubo.BestResult.Outlink, "NKSD")
	assert.Equal(t, bubo.BestResult.MatchType, vlib.Exact)
}

func TestAllSources(t *testing.T) {
	var response vlib.Output
	names := []string{
		"Bubo bubo",
	}
	request := vlib.Input{NameStrings: names, DataSources: []int{0}}
	req, err := gnfmt.GNjson{}.Encode(request)
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(restURL+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	assert.Equal(t, len(response.Names), len(names))
	assert.False(t, response.WithCapitalization)
	bubo := response.Names[0]
	assert.Equal(t, bubo.ID, "4431a0f3-e901-519a-886f-9b97e0c99d8e")
	assert.Equal(t, bubo.Name, "Bubo bubo")
	assert.NotNil(t, bubo.BestResult)
	assert.Equal(t, len(bubo.Results), bubo.DataSourcesNum)
	assert.True(t, bubo.DataSourcesNum > 20)
}

func TestAllMatches(t *testing.T) {
	var response vlib.Output
	names := []string{
		"Solanum tuberosum",
	}
	request := vlib.Input{
		NameStrings:    names,
		DataSources:    []int{1},
		WithAllMatches: true,
	}
	req, err := gnfmt.GNjson{}.Encode(request)
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(restURL+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	assert.Equal(t, len(response.Names), len(names))
	solanum := response.Names[0]
	assert.NotNil(t, solanum.BestResult)
	assert.Greater(t, len(solanum.Results), 1)
}

func TestAll(t *testing.T) {
	var response vlib.Output
	names := []string{
		"Solanum tuberosum",
	}
	request := vlib.Input{
		NameStrings:    names,
		DataSources:    []int{0},
		WithAllMatches: true,
	}
	req, err := gnfmt.GNjson{}.Encode(request)
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(restURL+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	assert.Equal(t, len(response.Names), len(names))
	solanum := response.Names[0]
	assert.NotNil(t, solanum.BestResult)
	assert.Greater(t, len(solanum.Results), 20)
}

func TestBugs(t *testing.T) {
	var response vlib.Output
	names := []string{
		"Aceratagallia fuscosscripta (Oman )",
		"Ampullaria immersa",
		"Abacetine",
	}
	request := vlib.Input{NameStrings: names}
	req, err := gnfmt.GNjson{}.Encode(request)
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(restURL+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	assert.Equal(t, len(response.Names), len(names))
}

// NCBI used to return "Homo sapiens subsp. Denisova" as the best result
// for "Homo sapiens" match. With #52 we introduced scoring by parsing quality
// and it should fix the match. This test is brittle, as it depends on
// NCBI keeping non-standard "Homo sapiens substp. Denisova" name-string.
func TestHomoNCBI(t *testing.T) {
	var response vlib.Output
	request := vlib.Input{
		NameStrings: []string{"Homo sapiens"},
		DataSources: []int{4},
	}
	req, err := gnfmt.GNjson{}.Encode(request)
	assert.Nil(t, err)
	r := bytes.NewReader(req)
	resp, err := http.Post(restURL+"verifications", "application/json", r)
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)
	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	homo := response.Names[0]
	assert.Equal(t, homo.BestResult.MatchedCanonicalSimple, "Homo sapiens")
	assert.NotContains(t, homo.Results[0].MatchedName, "Denisova")
}

func TestGetVerifications(t *testing.T) {
	var response vlib.Output
	resp, err := http.Get(restURL + "verifications/Homo+sapiens?data_sources=4")
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	homo := response.Names[0]
	assert.Equal(t, homo.BestResult.MatchedCanonicalSimple, "Homo sapiens")
	assert.NotContains(t, homo.Results[0].MatchedName, "Denisova")
}

func TestContext(t *testing.T) {
	var response vlib.Output
	resp, err := http.Get(restURL + "verifications/Homo+sapiens|Pan+troglodytes?context=true")
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	homo := response.Names[0]
	assert.Equal(t, homo.BestResult.MatchedCanonicalSimple, "Homo sapiens")
	assert.Equal(t, response.Context, "Hominidae")
	assert.Equal(t, response.ContextPercentage, float32(1.0))
}

func TestDataSources(t *testing.T) {
	var response []vlib.DataSource
	resp, err := http.Get(restURL + "data_sources")
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	assert.Greater(t, len(response), 50)
	log.Printf("%+v", response[0].ID)
	col := response[0]
	assert.Equal(t, col.Title, "Catalogue of Life")
}

func TestOneDataSource(t *testing.T) {
	var ds vlib.DataSource
	resp, err := http.Get(restURL + "data_sources/12")
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	err = gnfmt.GNjson{}.Decode(respBytes, &ds)
	assert.Nil(t, err)
	assert.Equal(t, ds.Title, "Encyclopedia of Life")
	assert.True(t, ds.IsOutlinkReady)
	assert.Equal(t, ds.WebsiteURL, "https://eol.org")
	assert.Equal(t, ds.UUID, "dba5f880-a40d-479b-a1ad-a646835edde4")
}
