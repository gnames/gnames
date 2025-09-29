package rest_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/gnames/gnames/pkg/config"
	"github.com/gnames/gnfmt"
	"github.com/gnames/gnlib"
	"github.com/gnames/gnlib/ent/gnvers"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var restURL = getConfig().GnamesHostURL + "/api/v1/"

func getConfig() config.Config {
	cfg := config.New()
	config.LoadEnv(&cfg)
	return cfg
}

// Test helper functions

// makePostRequest sends a POST request with JSON payload and returns the response
func makePostRequest(t *testing.T, endpoint string, payload any) *http.Response {
	// t.Helper() marks this as a helper function so test failures report the caller's
	// line number instead of this function's line number in the stack trace.
	t.Helper()

	reqBytes, err := gnfmt.GNjson{}.Encode(payload)
	// require stops test execution immediately on failure, preventing cascading errors
	require.NoError(t, err)

	resp, err := http.Post(restURL+endpoint, "application/json", bytes.NewReader(reqBytes))
	require.NoError(t, err)

	return resp
}

// makeGetRequest sends a GET request and returns the response
func makeGetRequest(t *testing.T, endpoint string) *http.Response {
	t.Helper()

	resp, err := http.Get(restURL + endpoint)
	require.NoError(t, err)

	return resp
}

// readResponseBody reads and returns the response body
func readResponseBody(t *testing.T, resp *http.Response) []byte {
	t.Helper()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	defer resp.Body.Close()

	return body
}

// decodeJSONResponse decodes JSON response into the provided target
func decodeJSONResponse(t *testing.T, body []byte, target any) {
	t.Helper()

	err := gnfmt.GNjson{}.Decode(body, target)
	require.NoError(t, err)
}

// postVerificationRequest is a helper for verification POST requests
func postVerificationRequest(t *testing.T, input vlib.Input) vlib.Output {
	t.Helper()

	resp := makePostRequest(t, "verifications", input)
	body := readResponseBody(t, resp)

	var output vlib.Output
	decodeJSONResponse(t, body, &output)

	return output
}

// getVerificationRequest is a helper for verification GET requests
func getVerificationRequest(t *testing.T, query string) vlib.Output {
	t.Helper()

	resp := makeGetRequest(t, "verifications/"+query)
	body := readResponseBody(t, resp)

	var output vlib.Output
	decodeJSONResponse(t, body, &output)

	return output
}

// TestPing tests the ping endpoint
func TestPing(t *testing.T) {
	resp := makeGetRequest(t, "ping")
	body := readResponseBody(t, resp)

	assert.Equal(t, "pong", string(body))
}

// TestVersion tests the version endpoint
func TestVersion(t *testing.T) {
	resp := makeGetRequest(t, "version")
	body := readResponseBody(t, resp)

	var response gnvers.Version
	decodeJSONResponse(t, body, &response)

	assert.Regexp(t, `^v\d+\.\d+\.\d+`, response.Version)
}

// TestVerifyExact tests the verification endpoint with exact matches
func TestVerifyExact(t *testing.T) {
	names := []string{
		"Not name",
		"Bubo bubo",
		"Pomatomus",
		"Pardosa moesta",
		"Plantago major var major",
		"Cytospora ribis mitovirus 2",
		"A-shaped rods",
		"Alb. alba",
		"Acacia vestita may",
		"Candidatus Aenigmarchaeum subterraneum",
		"Phegopteris",
	}

	request := vlib.Input{NameStrings: names}
	response := postVerificationRequest(t, request)

	require.Equal(t, len(names), len(response.Names))

	// Test cases with expected results
	testCases := []struct {
		index              int
		expectedID         string
		expectedName       string
		expectedMatchType  vlib.MatchTypeValue
		expectedCuration   vlib.CurationLevel
		hasBestResult      bool
		expectedDataSrcID  int
		expectedDataSrcNum int
		additionalChecks   func(t *testing.T, name *vlib.Name)
	}{
		{
			index:              0,
			expectedID:         "82dbfb99-fe6c-5882-99f2-17c7d3955599",
			expectedName:       "Not name",
			expectedMatchType:  vlib.NoMatch,
			expectedCuration:   vlib.NotCurated,
			hasBestResult:      false,
			expectedDataSrcNum: 0,
		},
		{
			index:              1,
			expectedID:         "4431a0f3-e901-519a-886f-9b97e0c99d8e",
			expectedName:       "Bubo bubo",
			expectedMatchType:  vlib.Exact,
			expectedCuration:   vlib.Curated,
			hasBestResult:      true,
			expectedDataSrcID:  1,
			expectedDataSrcNum: 34,
		},
		{
			index:             8,
			expectedID:        "0f84ed48-3a57-59ac-ac1a-2e9221439fdc",
			expectedName:      "Acacia vestita may",
			expectedMatchType: vlib.PartialExact,
			expectedCuration:  vlib.Curated,
			hasBestResult:     true,
			expectedDataSrcID: 1,
			additionalChecks: func(t *testing.T, name *vlib.Name) {
				assert.Equal(t, "Acacia vestita", name.BestResult.CurrentCanonicalSimple)
			},
		},
		{
			index:             9,
			expectedID:        "1b406033-fc5e-5f90-b3cf-fd1e9a42e282",
			expectedName:      "Candidatus Aenigmarchaeum subterraneum",
			expectedMatchType: vlib.Exact,
			expectedCuration:  vlib.AutoCurated,
			hasBestResult:     true,
			expectedDataSrcID: 179,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.expectedName, func(t *testing.T) {
			name := response.Names[tc.index]

			assert.Equal(t, tc.expectedID, name.ID)
			assert.Equal(t, tc.expectedName, name.Name)
			assert.Equal(t, tc.expectedCuration, name.Curation)
			assert.Empty(t, name.Error)

			if tc.hasBestResult {
				require.NotNil(t, name.BestResult)
				assert.Equal(t, tc.expectedDataSrcID, name.BestResult.DataSourceID)
				assert.Equal(t, tc.expectedMatchType, name.BestResult.MatchType)
			} else {
				assert.Nil(t, name.BestResult)
				assert.Equal(t, tc.expectedMatchType, name.MatchType)
			}

			if tc.expectedDataSrcNum > 0 {
				assert.Equal(t, tc.expectedDataSrcNum, name.DataSourcesNum)
				assert.Equal(t, tc.expectedDataSrcNum, len(name.DataSourcesIDs))
			}

			if tc.additionalChecks != nil {
				tc.additionalChecks(t, &name)
			}
		})
	}
}

func TestAuthors(t *testing.T) {
	tests := []struct {
		name          string
		inputName     string
		expectedMatch string
		expectedSrc   int
		expectedType  vlib.MatchTypeValue
	}{
		{
			name:          "Baudon abbreviation",
			inputName:     "Rissoa abbreviata Baudon, 1853",
			expectedMatch: "Rissoa abbreviata Baudon, 1853",
			expectedSrc:   9,
			expectedType:  vlib.Exact,
		},
		{
			name:          "Sowerby I expansion",
			inputName:     "Helix acuminata Sowerby, 1841",
			expectedMatch: "Helix acuminata G. B. Sowerby I, 1841",
			expectedSrc:   1,
			expectedType:  vlib.Exact,
		},
		{
			name:          "Linnaeus filius",
			inputName:     "Bubo bubo Linn. f.",
			expectedMatch: "Bubo bubo (Linnaeus, 1758)",
			expectedSrc:   1,
			expectedType:  vlib.Exact,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := vlib.Input{NameStrings: []string{tc.inputName}}
			response := postVerificationRequest(t, request)

			require.Len(t, response.Names, 1)
			name := response.Names[0]
			require.NotNil(t, name.BestResult)

			assert.Equal(t, tc.expectedMatch, name.BestResult.MatchedName)
			assert.Equal(t, tc.expectedSrc, name.BestResult.DataSourceID)
			assert.Equal(t, tc.expectedType, name.BestResult.MatchType)
		})
	}
}

func TestFuzzy(t *testing.T) {
	request := vlib.Input{
		NameStrings: []string{"Abras precatorius"},
		DataSources: []int{1, 12, 169, 182},
	}
	response := postVerificationRequest(t, request)

	require.Len(t, response.Names, 1)
	name := response.Names[0]

	assert.Equal(t, "Abras precatorius", name.Name)
	require.NotNil(t, name.BestResult)
	assert.Equal(t, 1, name.BestResult.EditDistance)
}

func TestRelaxedFuzzy(t *testing.T) {
	tests := []struct {
		name              string
		inputName         string
		expectedMatch     string
		useUniFuzzy       bool
		expectedEditDist  int
		expectedMatchType vlib.MatchTypeValue
	}{
		{
			name:              "partial fuzzy relaxed - missing letter",
			inputName:         "Bbo bubo onetwo",
			expectedMatch:     "Bubo bubo",
			useUniFuzzy:       false,
			expectedEditDist:  1,
			expectedMatchType: vlib.PartialFuzzyRelaxed,
		},
		{
			name:              "partial exact match",
			inputName:         "Bubo bubo onetwo",
			expectedMatch:     "Bubo bubo",
			useUniFuzzy:       false,
			expectedEditDist:  0,
			expectedMatchType: vlib.PartialExact,
		},
		{
			name:              "fuzzy relaxed - missing letter",
			inputName:         "Pomatomu saltator",
			expectedMatch:     "Pomatomus saltator",
			useUniFuzzy:       false,
			expectedEditDist:  1,
			expectedMatchType: vlib.FuzzyRelaxed,
		},
		{
			name:              "uninomial fuzzy with relaxed",
			inputName:         "Pomatomu L.",
			expectedMatch:     "Pomatomus",
			useUniFuzzy:       true,
			expectedEditDist:  1,
			expectedMatchType: vlib.FuzzyRelaxed,
		},
		{
			name:              "partial fuzzy relaxed with extra text",
			inputName:         "Pomatomu saltator aadsdss",
			expectedMatch:     "Pomatomus saltator",
			useUniFuzzy:       false,
			expectedEditDist:  1,
			expectedMatchType: vlib.PartialFuzzyRelaxed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := vlib.Input{
				NameStrings:           []string{tc.inputName},
				WithRelaxedFuzzyMatch: true,
			}
			if tc.useUniFuzzy {
				request.WithUninomialFuzzyMatch = true
			}

			response := postVerificationRequest(t, request)

			require.Len(t, response.Names, 1)
			name := response.Names[0]
			require.NotNil(t, name.BestResult)

			assert.Equal(t, tc.expectedMatch, name.BestResult.MatchedCanonicalSimple)
			assert.Equal(t, tc.expectedEditDist, name.BestResult.EditDistance)
			assert.Equal(t, tc.expectedMatchType, name.BestResult.MatchType)
		})
	}
}

// Issue  https://github.com/gnames/gnames/issues/108
// Checks if uninomials go through fuzzy matching
func TestUniFuzzy(t *testing.T) {
	tests := []struct {
		name              string
		inputName         string
		expectedResult    string
		dataSources       []int
		expectedMatchType vlib.MatchTypeValue
	}{
		{
			name:              "fuzzy uninomial match",
			inputName:         "Simulidae",
			expectedResult:    "Simuliidae",
			dataSources:       []int{3},
			expectedMatchType: vlib.Fuzzy,
		},
		{
			name:              "partial fuzzy uninomial match",
			inputName:         "Pomatmus abcdefg",
			expectedResult:    "Pomatomus",
			dataSources:       []int{1},
			expectedMatchType: vlib.PartialFuzzy,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := vlib.Input{
				NameStrings:             []string{tc.inputName},
				DataSources:             tc.dataSources,
				WithUninomialFuzzyMatch: true,
				WithAllMatches:          true,
			}
			response := postVerificationRequest(t, request)

			require.Len(t, response.Names, 1)
			name := response.Names[0]

			var foundMatch bool
			for _, result := range name.Results {
				if tc.expectedResult == result.CurrentCanonicalSimple {
					assert.Equal(t, tc.expectedMatchType, result.MatchType)
					foundMatch = true
					break
				}
			}
			assert.True(t, foundMatch, "Expected result %s not found in results", tc.expectedResult)
		})
	}
}

// TestPrefDS checks if preferred data sources works correctly.
func TestPrefDS(t *testing.T) {
	names := []string{
		"Bubo bubo", "Pomatomus",
		"Pardosa moesta", "Plantago major var major",
		"Cytospora ribis mitovirus 2",
	}
	request := vlib.Input{
		NameStrings:    names,
		DataSources:    []int{1, 12, 169, 182},
		WithAllMatches: true,
	}
	response := postVerificationRequest(t, request)

	require.Equal(t, len(names), len(response.Names))

	bubo := response.Names[0]
	assert.Equal(t, "4431a0f3-e901-519a-886f-9b97e0c99d8e", bubo.ID)
	assert.Equal(t, "Bubo bubo", bubo.Name)
	assert.Nil(t, bubo.BestResult)
	assert.Len(t, bubo.Results, 7)
	assert.Equal(t, 1, bubo.Results[0].DataSourceID)
	assert.Contains(t, bubo.Results[0].Outlink, "NKSD")
	assert.Equal(t, vlib.Exact, bubo.Results[0].MatchType)
	assert.Equal(t, vlib.Curated, bubo.Curation)
	assert.Empty(t, bubo.Error)
}

func TestPrefCapitalize(t *testing.T) {
	names := []string{
		"bubo bubo", "pomatomus",
		"pardosa moesta", "plantago major var major",
		"cytospora ribis mitovirus 2",
	}
	request := vlib.Input{NameStrings: names, WithCapitalization: true}
	response := postVerificationRequest(t, request)

	require.Equal(t, len(names), len(response.Names))
	assert.True(t, response.WithCapitalization)

	bubo := response.Names[0]
	assert.Equal(t, "7e4c9a7c-0e90-5d1e-96be-bbea21fcfdd3", bubo.ID)
	assert.Equal(t, "bubo bubo", bubo.Name)
	require.NotNil(t, bubo.BestResult)
	assert.Equal(t, 1, bubo.BestResult.DataSourceID)
	assert.Contains(t, bubo.BestResult.Outlink, "NKSD")
	assert.Equal(t, vlib.Exact, bubo.BestResult.MatchType)
}

func TestAllSources(t *testing.T) {
	request := vlib.Input{NameStrings: []string{"Bubo bubo"}}
	response := postVerificationRequest(t, request)

	require.Len(t, response.Names, 1)
	assert.False(t, response.WithCapitalization)

	bubo := response.Names[0]
	assert.Equal(t, "4431a0f3-e901-519a-886f-9b97e0c99d8e", bubo.ID)
	assert.Equal(t, "Bubo bubo", bubo.Name)
	require.NotNil(t, bubo.BestResult)
	assert.Greater(t, bubo.DataSourcesNum, 20)
	assert.Empty(t, bubo.Results)
}

func TestAllMatches(t *testing.T) {
	request := vlib.Input{
		NameStrings:    []string{"Solanum tuberosum"},
		DataSources:    []int{1},
		WithAllMatches: true,
	}
	response := postVerificationRequest(t, request)

	require.Len(t, response.Names, 1)
	solanum := response.Names[0]
	assert.Nil(t, solanum.BestResult)
	assert.Greater(t, len(solanum.Results), 1)
}

func TestAll(t *testing.T) {
	request := vlib.Input{
		NameStrings:    []string{"Solanum tuberosum"},
		WithAllMatches: true,
	}
	response := postVerificationRequest(t, request)

	require.Len(t, response.Names, 1)
	solanum := response.Names[0]
	assert.Nil(t, solanum.BestResult)
	assert.Greater(t, len(solanum.Results), 20)
}

func TestBugs(t *testing.T) {
	names := []string{
		"Aceratagallia fuscosscripta (Oman )",
		"Ampullaria immersa",
		"Abacetine",
	}
	request := vlib.Input{NameStrings: names}
	response := postVerificationRequest(t, request)

	assert.Equal(t, len(names), len(response.Names))
}

// NCBI used to return "Homo sapiens subsp. Denisova" as the best result
// for "Homo sapiens" match. With #52 we introduced scoring by parsing quality
// and it should fix the match. This test is brittle, as it depends on
// NCBI keeping non-standard "Homo sapiens substp. Denisova" name-string.
func TestHomoNCBI(t *testing.T) {
	request := vlib.Input{
		NameStrings:    []string{"Homo sapiens"},
		DataSources:    []int{4},
		WithAllMatches: true,
	}
	response := postVerificationRequest(t, request)

	require.Len(t, response.Names, 1)
	homo := response.Names[0]
	assert.Nil(t, homo.BestResult)
	assert.Greater(t, len(homo.Results), 0)
	assert.Equal(t, "Homo sapiens", homo.Results[0].MatchedCanonicalSimple)
	assert.NotContains(t, homo.Results[0].MatchedName, "Denisova")
}

func TestGetVerifications(t *testing.T) {
	response := getVerificationRequest(t, "Homo+sapiens?data_sources=4&all_matches=true")

	require.Len(t, response.Names, 1)
	homo := response.Names[0]
	assert.Nil(t, homo.BestResult)
	assert.Greater(t, len(homo.Results), 0)
	assert.Equal(t, "Homo sapiens", homo.Results[0].MatchedCanonicalSimple)
	assert.NotContains(t, homo.Results[0].MatchedName, "Denisova")
}

func TestRootClassification(t *testing.T) {
	response := getVerificationRequest(t, "Animalia?data_sources=3")

	require.Len(t, response.Names, 1)
	result := response.Names[0].BestResult
	require.NotNil(t, result)
	assert.Equal(t, "Animalia", result.ClassificationPath)
}

// TestMainTaxon adds stats attribute	that should generate data for what is the
// MainTaxon that encompasses both Homo and Pan.
func TestMainTaxon(t *testing.T) {
	response := getVerificationRequest(t, "Homo+sapiens|Pan+troglodytes?stats=true")

	require.Len(t, response.Names, 2)
	homo := response.Names[0]
	require.NotNil(t, homo.BestResult)
	assert.Equal(t, "Homo sapiens", homo.BestResult.MatchedCanonicalSimple)
	assert.Equal(t, "Homininae", response.MainTaxon)
	assert.Equal(t, float32(1.0), response.MainTaxonPercentage)
}

// TestSpeciesGroup checks species_group attribute. When this attribute is
// given, the results should include autonyms/species groups for botanical/zoological
// names. For example `Narcissus minor` would also return matches of
// `Narcissus minor minor`.
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

// TestDataSource checks data_sources endpoint.
func TestDataSources(t *testing.T) {
	resp := makeGetRequest(t, "data_sources")
	body := readResponseBody(t, resp)

	var response []vlib.DataSource
	decodeJSONResponse(t, body, &response)

	assert.Greater(t, len(response), 50)
	col := response[0]
	assert.Equal(t, "Catalogue of Life", col.Title)
}

// TestOneDataSource checks data_sources/{id} endpoint.
func TestOneDataSource(t *testing.T) {
	resp := makeGetRequest(t, "data_sources/12")
	body := readResponseBody(t, resp)

	var ds vlib.DataSource
	decodeJSONResponse(t, body, &ds)

	assert.Equal(t, "Encyclopedia of Life", ds.Title)
	assert.True(t, ds.IsOutlinkReady)
	assert.Equal(t, "https://eol.org", ds.WebsiteURL)
}

// VernacularGET checks `vernaculars` attribute with GET.
func TestVernacularGET(t *testing.T) {
	tests := []struct {
		name         string
		species      string
		sources      string
		vernLangs    string
		expectedVern []string
	}{
		{
			name:         "snowy egret eng+rus",
			species:      "Egretta thula",
			sources:      "180",
			vernLangs:    "eng|rus",
			expectedVern: []string{"Snowy Egret", "Белая американская цапля"},
		},
		{
			name:         "snowy egret all languages",
			species:      "Egretta thula",
			sources:      "1",
			vernLangs:    "all",
			expectedVern: []string{"Snowy Egret", "Aigrette neigeuse"},
		},
		{
			name:         "puma synonym",
			species:      "Felis concolor",
			sources:      "1",
			vernLangs:    "eng",
			expectedVern: []string{"Puma", "Mountain Lion"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			n := url.QueryEscape(tc.species)
			ds := url.QueryEscape(tc.sources)
			langs := url.QueryEscape(tc.vernLangs)
			query := fmt.Sprintf("%s?vernaculars=%s&data_sources=%s", n, langs, ds)

			response := getVerificationRequest(t, query)

			require.Len(t, response.Names, 1)
			require.NotNil(t, response.Names[0].BestResult)

			vernaculars := response.Names[0].BestResult.Vernaculars
			vernNames := gnlib.Map(vernaculars, func(v vlib.Vernacular) string {
				return v.Name
			})
			allNames := strings.Join(vernNames, "|")

			for _, expectedName := range tc.expectedVern {
				assert.Contains(t, allNames, expectedName)
			}
		})
	}
}

// TestVernacularPOST checks if Input.Vernaculars work correctly with POST.
func TestVernacularPOST(t *testing.T) {
	tests := []struct {
		name         string
		species      string
		dataSource   int
		languages    []string
		expectedVern []string
	}{
		{
			name:         "snowy egret eng+rus",
			species:      "Egretta thula",
			dataSource:   180,
			languages:    []string{"eng", "rus"},
			expectedVern: []string{"Snowy Egret", "Белая американская цапля"},
		},
		{
			name:         "puma synonym",
			species:      "Felis concolor",
			dataSource:   1,
			languages:    []string{"eng"},
			expectedVern: []string{"Puma", "Mountain Lion"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := vlib.Input{
				NameStrings: []string{tc.species},
				DataSources: []int{tc.dataSource},
				Vernaculars: tc.languages,
			}
			response := postVerificationRequest(t, request)

			require.Len(t, response.Names, 1)
			name := response.Names[0]
			require.NotNil(t, name.BestResult)
			assert.Equal(t, tc.dataSource, name.BestResult.DataSourceID)

			vernaculars := name.BestResult.Vernaculars
			vernNames := gnlib.Map(vernaculars, func(v vlib.Vernacular) string {
				return v.Name
			})
			allNames := strings.Join(vernNames, "|")

			for _, expectedName := range tc.expectedVern {
				assert.Contains(t, allNames, expectedName)
			}
		})
	}
}

// TestBestResults checks if BestResults field is populated correctly.
// BestResults should be empty when there's only one best match,
// and contain multiple entries when there are ties in the best score.
// Fix issue #135.
func TestBestResults(t *testing.T) {
	tests := []struct {
		name               string
		species            string
		dataSource         int
		expectEmpty        bool
		expectMultiple     bool
		minExpectedResults int
	}{
		{
			name:           "Bubo bubo - single best result",
			species:        "Bubo bubo",
			dataSource:     1,
			expectEmpty:    true,
			expectMultiple: false,
		},
		{
			name:               "Ficus variegata - multiple best results",
			species:            "Ficus variegata",
			dataSource:         1,
			expectEmpty:        false,
			expectMultiple:     true,
			minExpectedResults: 2,
		},
		{
			name:               "Pisonia grandis - multiple best results",
			species:            "Pisonia grandis",
			dataSource:         1,
			expectEmpty:        false,
			expectMultiple:     true,
			minExpectedResults: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			request := vlib.Input{
				NameStrings: []string{tc.species},
				DataSources: []int{tc.dataSource},
			}
			response := postVerificationRequest(t, request)

			require.Len(t, response.Names, 1)
			name := response.Names[0]
			require.NotNil(t, name.BestResult)

			if tc.expectEmpty {
				assert.Nil(t, name.BestResults, "BestResults should be nil for %s", tc.species)
			}

			if tc.expectMultiple {
				require.NotNil(t, name.BestResults, "BestResults should not be nil for %s", tc.species)
				assert.GreaterOrEqual(t, len(name.BestResults), tc.minExpectedResults,
					"BestResults should have at least %d entries for %s", tc.minExpectedResults, tc.species)

				// Verify all BestResults have the same score as BestResult
				for i, result := range name.BestResults {
					assert.Equal(t, name.BestResult.SortScore, result.SortScore,
						"BestResults[%d] should have the same SortScore as BestResult for %s", i, tc.species)
				}
			}
		})
	}
}
