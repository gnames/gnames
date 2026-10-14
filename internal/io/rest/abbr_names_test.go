package rest_test

import (
	"net/http"
	"net/url"
	"testing"

	alib "github.com/gnames/gnlib/ent/abbr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func postAbbrNames(t *testing.T, input alib.Input) alib.Output {
	t.Helper()

	resp := makePostRequest(t, "abbr_names", input)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var res alib.Output
	decodeJSONResponse(t, readResponseBody(t, resp), &res)
	return res
}

func TestAbbrNamesPOST(t *testing.T) {
	assert := assert.New(t)
	res := postAbbrNames(t, alib.Input{
		Names: []alib.Name{
			{ID: "1", Name: "C. elegans"},
			{ID: "2", Name: "C. briggsae Dougherty & Nigon, 1949"},
			{ID: "3", Name: "C. remanei"},
			{ID: "4", Name: "C. elegans"},
		},
		GeneraLimit: 3,
	})

	assert.Empty(res.Error)
	assert.Equal("C.", res.Abbreviation)
	assert.Equal(4, res.NamesNumber)
	assert.Equal(4, res.MatchedNamesNumber)
	require.NotEmpty(t, res.Genera)

	g := res.Genera[0]
	assert.Equal("Caenorhabditis", g.Name)
	assert.Contains(g.ClassificationPath, "Rhabditidae")
	assert.Greater(g.EpithetsNumber, 10)
	assert.Equal(4, len(g.MatchedNames))
	assert.Empty(g.UnmatchedNames)
	for _, m := range g.MatchedNames {
		assert.Equal(alib.Full, m.MatchType)
		assert.False(m.StemMatch)
	}
	assert.Equal("Caenorhabditis briggsae", g.MatchedNames[2].MatchedCanonical)
	assert.Greater(g.Score.Total, 0.0)
}

func TestAbbrNamesHomonym(t *testing.T) {
	assert := assert.New(t)
	names := []alib.Name{
		{Name: "M. alba"}, {Name: "M. nigra"}, {Name: "M. bassana"},
	}

	tests := []struct {
		class, path string
	}{
		{"Aves", "Sulidae"},
		{"Magnoliopsida", "Moraceae"},
	}
	for _, v := range tests {
		res := postAbbrNames(t, alib.Input{Names: names, DominantClass: v.class})
		require.NotEmpty(t, res.Genera, v.class)
		g := res.Genera[0]
		assert.Equal("Morus", g.Name, v.class)
		assert.Contains(g.ClassificationPath, v.path, v.class)
		assert.Equal(0.5, g.Score.DominantClass, v.class)
	}
}

func TestAbbrNamesError(t *testing.T) {
	assert := assert.New(t)
	res := postAbbrNames(t, alib.Input{
		Names: []alib.Name{{Name: "C. elegans"}, {Name: "Mus musculus"}},
	})
	assert.Contains(res.Error, "Mus musculus")
	assert.Empty(res.Genera)
	assert.Equal(2, res.NamesNumber)
}

func TestAbbrNamesGET(t *testing.T) {
	assert := assert.New(t)
	names := url.PathEscape("P. tigris altaica|P. leo|P. pardus")
	resp := makeGetRequest(t, "abbr_names/"+names+"?genera_limit=2&tolerance=0")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var res alib.Output
	decodeJSONResponse(t, readResponseBody(t, resp), &res)
	assert.Equal(2, res.GeneraLimit)
	assert.Equal(0, res.Tolerance)
	require.NotEmpty(t, res.Genera)
	assert.Equal("Panthera", res.Genera[0].Name)
	assert.Equal(
		"Panthera tigris altaica",
		res.Genera[0].MatchedNames[0].MatchedCanonical,
	)
}
