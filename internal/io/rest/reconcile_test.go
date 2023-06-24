package rest_test

import (
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/gnames/gnfmt"
	"github.com/gnames/gnlib/ent/reconciler"
	"github.com/stretchr/testify/assert"
)

func TestReconcileManifest(t *testing.T) {
	var response reconciler.Manifest
	assert := assert.New(t)
	resp, err := http.Get(restURL +
		"reconcile")
	assert.Nil(err)
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(err)
	err = gnfmt.GNjson{}.Decode(respBytes, &response)
	assert.Nil(err)
	assert.Equal("GlobalNames", response.Name)
}

func TestReconcileExact(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		msg, name string
		len       int
		id        string
		score     float64
		match     bool
	}{
		{"1", "Not name", 0, "", 0.0, false},
		{"2", "Bubo bubo", 4, "0eeccd70-eaf2-5c51-ad8b-46cfb3db1645", 1, true},
		{"3", "Pomatomus", 3, "82110143-0b8d-50f6-b34d-e2ae118f4e2e", 1, true},
		{"4", "Pardosa moesta", 7, "e2fdf10b-6a36-5cc7-b6ca-be4d3b34b21f", 1, true},
		{"5", "Plantago major var major", 7, "bdfc5d4c-478b-5b3f-8f03-375e4daadc04", 1, true},
		{"6", "Cytospora ribis mitovirus 2", 1, "bd8cc487-9a28-5910-8d98-38d2b43d1dcb", 0.675, false},
		{"7", "A-shaped rods", 0, "", 0.0, false},
		{"8", "Alb. alba", 0, "", 0.0, false},
		{"9", "Pisonia grandis", 4, "97e46f64-2673-54aa-8687-7b7bad7c9b64", 1, true},
		{"10", "Acacia vestita may", 2, "290d25e5-ce87-5cfe-b092-1bd12cf55bc1", 0.75, false},
		{"11", "Candidatus Aenigmarchaeum subterraneum", 1, "1b406033-fc5e-5f90-b3cf-fd1e9a42e282", 0.9, false},
	}
	q := make(map[string]reconciler.Query)
	for _, v := range tests {
		q[v.msg] = reconciler.Query{
			Query: v.name,
		}
	}
	req, err := gnfmt.GNjson{}.Encode(q)
	assert.Nil(err)
	resp, err := http.PostForm(
		restURL+"reconcile",
		url.Values{"queries": {string(req)}},
	)
	assert.Nil(err)
	defer resp.Body.Close()

	var o reconciler.Output
	respBytes, err := io.ReadAll(resp.Body)
	assert.Nil(err)
	err = gnfmt.GNjson{}.Decode(respBytes, &o)
	assert.Nil(err)
	for _, v := range tests {
		res := o[v.msg]
		assert.Equal(v.len, len(res.Result), v.msg)
		if len(res.Result) > 0 {
			assert.Equal(v.match, res.Result[0].Match, v.msg)
			assert.InDelta(v.score, res.Result[0].Score, 0.01, v.msg)
			assert.Equal(v.id, res.Result[0].ID, v.msg)
		}
	}
}
