package rest_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/gnames/gnfmt"
	"github.com/gnames/gnlib/ent/gnvers"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/stretchr/testify/assert"
)

const restURL = "http://:8888/api/v1/"

func TestPing(t *testing.T) {
	resp, err := http.Get(restURL + "ping")
	assert.Nil(t, err)

	response, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	assert.Equal(t, "pong", string(response))
}

func TestVersion(t *testing.T) {
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
	assert.Equal(t, len(names), len(response.Names))

	bad := response.Names[0]
	assert.Equal(t, "82dbfb99-fe6c-5882-99f2-17c7d3955599", bad.ID)
	assert.Equal(t, "Not name", bad.Name)
	assert.Equal(t, vlib.NoMatch, bad.MatchType)
	assert.Nil(t, bad.BestResult)
	assert.Equal(t, 0, bad.DataSourcesNum)
	assert.Equal(t, 0, len(bad.DataSourcesIDs))
	assert.Equal(t, vlib.NotCurated, bad.Curation)
	assert.Equal(t, "", bad.Error)

	binom := response.Names[1]
	assert.Equal(t, "4431a0f3-e901-519a-886f-9b97e0c99d8e", binom.ID)
	assert.Equal(t, "Bubo bubo", binom.Name)
	assert.NotNil(t, binom.BestResult)
	assert.Equal(t, 1, binom.BestResult.DataSourceID)
	assert.Equal(t, 31, len(binom.DataSourcesIDs))
	assert.Equal(t, 31, binom.DataSourcesNum)
	assert.Equal(t, vlib.Exact, binom.BestResult.MatchType)
	assert.Equal(t, vlib.Curated, binom.Curation)
	assert.Equal(t, "", binom.Error)

	acceptFilter := response.Names[8]
	assert.Equal(t, "4c8848f2-7271-588c-ba81-e4d5efcc1e92", acceptFilter.ID)
	assert.Equal(t, "Pisonia grandis", acceptFilter.Name)
	assert.Equal(t, 1, acceptFilter.BestResult.DataSourceID)
	assert.Equal(t, vlib.Exact, acceptFilter.BestResult.MatchType)
	assert.Equal(t, "Ceodes grandis", acceptFilter.BestResult.CurrentCanonicalSimple)

	partial := response.Names[9]
	assert.Equal(t, "0f84ed48-3a57-59ac-ac1a-2e9221439fdc", partial.ID)
	assert.Equal(t, "Acacia vestita may", partial.Name)
	assert.Equal(t, 1, partial.BestResult.DataSourceID)
	assert.Equal(t, vlib.PartialExact.String(), partial.MatchType.String())
	assert.Equal(t, "Acacia vestita", partial.BestResult.CurrentCanonicalSimple)

	cand := response.Names[10]
	assert.Equal(t, "1b406033-fc5e-5f90-b3cf-fd1e9a42e282", cand.ID)
	assert.Equal(t, "Candidatus Aenigmarchaeum subterraneum", cand.Name)
	assert.NotNil(t, cand.BestResult)
	assert.Equal(t, 179, cand.BestResult.DataSourceID)
	assert.Equal(t, vlib.Exact, cand.BestResult.MatchType)
	assert.Equal(t, vlib.AutoCurated, cand.Curation)
	assert.Equal(t, "", cand.Error)
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
	assert.Equal(t, len(names), len(response.Names))

	fuz1 := response.Names[0]
	assert.Equal(t, "Abras precatorius", fuz1.Name)
	assert.Equal(t, 1, fuz1.BestResult.EditDistance)
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
	request := vlib.Input{
		NameStrings:    names,
		DataSources:    []int{1, 12, 169, 182},
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
	assert.Equal(t, len(names), len(response.Names))

	binom := response.Names[0]
	assert.Equal(t, "4431a0f3-e901-519a-886f-9b97e0c99d8e", binom.ID)
	assert.Equal(t, "Bubo bubo", binom.Name)
	assert.Nil(t, binom.BestResult)
	assert.Equal(t, 7, len(binom.Results))
	assert.Equal(t, 1, binom.Results[0].DataSourceID)
	assert.Contains(t, binom.Results[0].Outlink, "NKSD")
	assert.Equal(t, vlib.Exact, binom.Results[0].MatchType)
	assert.Equal(t, vlib.Curated, binom.Curation)
	assert.Equal(t, "", binom.Error)

	acceptFilter := response.Names[5]
	assert.Equal(t, "4c8848f2-7271-588c-ba81-e4d5efcc1e92", acceptFilter.ID)
	assert.Equal(t, "Pisonia grandis", acceptFilter.Name)
	assert.Nil(t, binom.BestResult)
	assert.True(t, len(binom.Results) > 0)
	assert.Equal(t, 1, acceptFilter.Results[0].DataSourceID)
	assert.Equal(t, vlib.Exact, acceptFilter.Results[0].MatchType)
	assert.Equal(t, "Ceodes grandis", acceptFilter.Results[0].CurrentCanonicalSimple)
	assert.Equal(t, 7, len(binom.Results))
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
	assert.Equal(t, len(names), len(response.Names))
	assert.True(t, response.WithCapitalization)

	bubo := response.Names[0]
	assert.Equal(t, "7e4c9a7c-0e90-5d1e-96be-bbea21fcfdd3", bubo.ID)
	assert.Equal(t, "bubo bubo", bubo.Name)
	assert.NotNil(t, bubo.BestResult)
	assert.Equal(t, 1, bubo.BestResult.DataSourceID)
	assert.Contains(t, bubo.BestResult.Outlink, "NKSD")
	assert.Equal(t, vlib.Exact, bubo.BestResult.MatchType)
}

func TestAllSources(t *testing.T) {
	var response vlib.Output
	names := []string{
		"Bubo bubo",
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
	assert.Equal(t, len(names), len(response.Names))
	assert.False(t, response.WithCapitalization)
	bubo := response.Names[0]
	assert.Equal(t, "4431a0f3-e901-519a-886f-9b97e0c99d8e", bubo.ID)
	assert.Equal(t, "Bubo bubo", bubo.Name)
	assert.NotNil(t, bubo.BestResult)
	assert.True(t, bubo.DataSourcesNum > 20)
	assert.Equal(t, len(bubo.Results), 0)
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
	assert.Equal(t, len(names), len(response.Names))
	solanum := response.Names[0]
	assert.Nil(t, solanum.BestResult)
	assert.Greater(t, len(solanum.Results), 1)
}

func TestAll(t *testing.T) {
	var response vlib.Output
	names := []string{
		"Solanum tuberosum",
	}
	request := vlib.Input{
		NameStrings:    names,
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
	assert.Equal(t, len(names), len(response.Names))
	solanum := response.Names[0]
	assert.Nil(t, solanum.BestResult)
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
	assert.Equal(t, len(names), len(response.Names))
}

// NCBI used to return "Homo sapiens subsp. Denisova" as the best result
// for "Homo sapiens" match. With #52 we introduced scoring by parsing quality
// and it should fix the match. This test is brittle, as it depends on
// NCBI keeping non-standard "Homo sapiens substp. Denisova" name-string.
func TestHomoNCBI(t *testing.T) {
	var response vlib.Output
	request := vlib.Input{
		NameStrings:    []string{"Homo sapiens"},
		DataSources:    []int{4},
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
	homo := response.Names[0]
	assert.Nil(t, homo.BestResult)
	assert.True(t, len(homo.Results) > 0)
	assert.Equal(t, "Homo sapiens", homo.Results[0].MatchedCanonicalSimple)
	assert.NotContains(t, homo.Results[0].MatchedName, "Denisova")
}

func TestGetVerifications(t *testing.T) {
	var response vlib.Output
	resp, err := http.Get(restURL + "verifications/Homo+sapiens?data_sources=4&all_matches=true")
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	homo := response.Names[0]
	assert.Nil(t, homo.BestResult)
	assert.True(t, len(homo.Results) > 0)
	assert.Equal(t, "Homo sapiens", homo.Results[0].MatchedCanonicalSimple)
	assert.NotContains(t, homo.Results[0].MatchedName, "Denisova")
}

func TestMainTaxon(t *testing.T) {
	var response vlib.Output
	resp, err := http.Get(restURL + "verifications/Homo+sapiens|Pan+troglodytes?stats=true")
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	homo := response.Names[0]
	assert.Equal(t, "Homo sapiens", homo.BestResult.MatchedCanonicalSimple)
	assert.Equal(t, "Homininae", response.MainTaxon)
	assert.Equal(t, float32(1.0), response.MainTaxonPercentage)
}

func TestSpeciesGroup(t *testing.T) {
	var response vlib.Output
	resp, err := http.Get(restURL + "verifications/Narcissus+minor?all_matches=true&species_group=true")
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	spGroup := response.Names[0]

	resp, err = http.Get(restURL + "verifications/Narcissus+minor?all_matches=true&species_group=false")
	assert.Nil(t, err)
	respBytes, err = io.ReadAll(resp.Body)
	assert.Nil(t, err)

	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	noSpGroup := response.Names[0]
	assert.Greater(t, len(spGroup.Results), len(noSpGroup.Results))
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
	col := response[0]
	assert.Equal(t, "Catalogue of Life", col.Title)
}

func TestOneDataSource(t *testing.T) {
	var ds vlib.DataSource
	resp, err := http.Get(restURL + "data_sources/12")
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	err = gnfmt.GNjson{}.Decode(respBytes, &ds)
	assert.Nil(t, err)
	assert.Equal(t, "Encyclopedia of Life", ds.Title)
	assert.True(t, ds.IsOutlinkReady)
	assert.Equal(t, "https://eol.org", ds.WebsiteURL)
}
