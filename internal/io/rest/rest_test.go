package rest_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/gnames/gnames/pkg/config"
	"github.com/gnames/gnfmt"
	"github.com/gnames/gnlib/ent/gnvers"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/stretchr/testify/assert"
)

var restURL = getConfig().GnamesHostURL + "/api/v1/"

func getConfig() config.Config {
	cfg := config.New()
	config.LoadEnv(&cfg)
	return cfg
}

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
	assert := assert.New(t)
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
		"Phegopteris",
	}
	request := vlib.Input{NameStrings: names}
	req, err := gnfmt.GNjson{}.Encode(request)
	assert.Nil(err)
	r := bytes.NewReader(req)
	resp, err := http.Post(restURL+"verifications", "application/json", r)
	assert.Nil(err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(err)
	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(err)
	assert.Equal(len(names), len(response.Names))

	bad := response.Names[0]
	assert.Equal("82dbfb99-fe6c-5882-99f2-17c7d3955599", bad.ID)
	assert.Equal("Not name", bad.Name)
	assert.Equal(vlib.NoMatch, bad.MatchType)
	assert.Nil(bad.BestResult)
	assert.Equal(0, bad.DataSourcesNum)
	assert.Equal(0, len(bad.DataSourcesIDs))
	assert.Equal(vlib.NotCurated, bad.Curation)
	assert.Equal("", bad.Error)

	binom := response.Names[1]
	assert.Equal("4431a0f3-e901-519a-886f-9b97e0c99d8e", binom.ID)
	assert.Equal("Bubo bubo", binom.Name)
	assert.NotNil(binom.BestResult)
	assert.Equal(1, binom.BestResult.DataSourceID)
	assert.Equal(34, len(binom.DataSourcesIDs))
	assert.Equal(34, binom.DataSourcesNum)
	assert.Equal(vlib.Exact, binom.BestResult.MatchType)
	assert.Equal(vlib.Curated, binom.Curation)
	assert.Equal("", binom.Error)

	acceptFilter := response.Names[8]
	assert.Equal("4c8848f2-7271-588c-ba81-e4d5efcc1e92", acceptFilter.ID)
	assert.Equal("Pisonia grandis", acceptFilter.Name)
	assert.Equal(1, acceptFilter.BestResult.DataSourceID)
	assert.Equal(vlib.Exact, acceptFilter.BestResult.MatchType)
	assert.Equal("Ceodes artensis", acceptFilter.BestResult.CurrentCanonicalSimple)

	partial := response.Names[9]
	assert.Equal("0f84ed48-3a57-59ac-ac1a-2e9221439fdc", partial.ID)
	assert.Equal("Acacia vestita may", partial.Name)
	assert.Equal(1, partial.BestResult.DataSourceID)
	assert.Equal(vlib.PartialExact.String(), partial.MatchType.String())
	assert.Equal("Acacia vestita", partial.BestResult.CurrentCanonicalSimple)

	cand := response.Names[10]
	assert.Equal("1b406033-fc5e-5f90-b3cf-fd1e9a42e282", cand.ID)
	assert.Equal("Candidatus Aenigmarchaeum subterraneum", cand.Name)
	assert.NotNil(cand.BestResult)
	assert.Equal(179, cand.BestResult.DataSourceID)
	assert.Equal(vlib.Exact, cand.BestResult.MatchType)
	assert.Equal(vlib.AutoCurated, cand.Curation)
	assert.Equal("", cand.Error)
}

func TestAuthors(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		msg, name, match string
		src              int
		matchType        vlib.MatchTypeValue
	}{
		{
			msg:       "Bandon",
			name:      "Rissoa abbreviata Baudon, 1853",
			match:     "Rissoa abbreviata Baudon, 1853",
			src:       9,
			matchType: vlib.Exact,
		},
		{
			msg:       "I",
			name:      "Helix acuminata Sowerby, 1841",
			match:     "Helix acuminata G. B. Sowerby I, 1841",
			src:       1,
			matchType: vlib.Exact,
		},
		{
			msg:       "filius",
			name:      "Bubo bubo Linn. f.",
			match:     "Bubo bubo (Linnaeus, 1758)",
			src:       1,
			matchType: vlib.Exact,
		},
	}
	for _, v := range tests {
		var response vlib.Output
		request := vlib.Input{NameStrings: []string{v.name}}
		req, err := gnfmt.GNjson{}.Encode(request)
		assert.Nil(err)
		r := bytes.NewReader(req)
		resp, err := http.Post(restURL+"verifications", "application/json", r)
		assert.Nil(err)
		respBytes, err := io.ReadAll(resp.Body)
		assert.Nil(err)
		err = gnfmt.GNjson{}.Decode(respBytes, &response)
		assert.Nil(err)
		name := response.Names[0]
		assert.Equal(v.match, name.BestResult.MatchedName)
		assert.Equal(v.src, name.BestResult.DataSourceID)
		assert.Equal(v.matchType, name.BestResult.MatchType)
	}
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

func TestRelaxedFuzzy(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		msg, name, match string
		uniFuzzy         bool
		ed               int
		typ              vlib.MatchTypeValue
	}{
		{"bubo", "Bbo bubo onetwo", "Bubo bubo", false, 1, vlib.PartialFuzzyRelaxed},
		{"bubo", "Bubo bubo onetwo", "Bubo bubo", false, 0, vlib.PartialExact},
		{"pom saltator", "Pomatomu saltator", "Pomatomus saltator", false, 1, vlib.FuzzyRelaxed},
		{"pomatomus", "Pomatomu L.", "Pomatomus", true, 1, vlib.FuzzyRelaxed},
		{
			"pomatomus part",
			"Pomatomu saltator aadsdss",
			"Pomatomus saltator",
			false,
			1,
			vlib.PartialFuzzyRelaxed,
		},
	}

	for _, v := range tests {
		var response vlib.Output
		request := vlib.Input{
			NameStrings:           []string{v.name},
			WithRelaxedFuzzyMatch: true,
		}
		if v.uniFuzzy {
			request.WithUninomialFuzzyMatch = true
		}
		req, err := gnfmt.GNjson{}.Encode(request)
		assert.Nil(err)
		r := bytes.NewReader(req)
		resp, err := http.Post(restURL+"verifications", "application/json", r)
		assert.Nil(err)
		respBytes, err := io.ReadAll(resp.Body)
		assert.Nil(err)
		err = gnfmt.GNjson{}.Decode(respBytes, &response)
		assert.Nil(err)
		name := response.Names[0]
		assert.Equal(v.match, name.BestResult.MatchedCanonicalSimple, v.msg)
		assert.Equal(v.ed, name.BestResult.EditDistance, v.msg)
		assert.Equal(v.typ, name.BestResult.MatchType, v.msg)
	}
}

// Issue  https://github.com/gnames/gnames/issues/108
// Checks if uninomials go through fuzzy matching
func TestUniFuzzy(t *testing.T) {
	tests := []struct {
		msg, name, res string
		ds             []int
		matchType      vlib.MatchTypeValue
	}{
		{
			msg:       "fuzzy",
			name:      "Simulidae",
			res:       "Simuliidae",
			ds:        []int{3},
			matchType: vlib.Fuzzy,
		},
		{
			msg:       "partialFuzzy",
			name:      "Pomatmus abcdefg",
			res:       "Pomatomus",
			ds:        []int{1},
			matchType: vlib.PartialFuzzy,
		},
	}

	for _, v := range tests {
		var response vlib.Output
		request := vlib.Input{
			NameStrings:             []string{v.name},
			DataSources:             v.ds,
			WithUninomialFuzzyMatch: true,
			WithAllMatches:          true,
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
		name := response.Names[0]
		var isFuzzy bool
		for _, vv := range name.Results {
			if v.res == vv.CurrentCanonicalSimple {
				assert.Equal(t, v.matchType, vv.MatchType, v.msg)
				isFuzzy = true
			}
		}
		assert.True(t, isFuzzy)
	}

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
	assert.Equal(t, "Ceodes artensis", acceptFilter.Results[0].CurrentCanonicalSimple)
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

func TestRootClassification(t *testing.T) {
	var response vlib.Output
	resp, err := http.Get(restURL + "verifications/Animalia?data_sources=3")
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	res := response.Names[0].BestResult
	assert.Equal(t, "Animalia", res.ClassificationPath)
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
	resp, err := http.Get(
		restURL + "verifications/Narcissus+minor?all_matches=true&species_group=true",
	)
	assert.Nil(t, err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(t, err)

	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(t, err)
	spGroup := response.Names[0]

	resp, err = http.Get(
		restURL + "verifications/Narcissus+minor?all_matches=true&species_group=false",
	)
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
