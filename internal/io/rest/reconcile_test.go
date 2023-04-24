package rest_test

import (
	"io"
	"net/http"
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
