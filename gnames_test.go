package gnames_test

import (
	"context"
	"testing"

	"github.com/gnames/gnames"
	"github.com/gnames/gnames/config"
	"github.com/gnames/gnames/ent/verifier"
	mlib "github.com/gnames/gnlib/ent/matcher"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnquery/ent/search"
	"github.com/stretchr/testify/assert"
)

func TestVerifier(t *testing.T) {
	var g gnames.GNames
	cfg := config.New()
	ctx := context.Background()
	vf := mockVerifier{}
	fct := mockFacet{}
	g = gnames.NewGNames(cfg, vf, fct)
	testData := []struct {
		name string
	}{
		{"Bubo bubo"},
	}
	for _, v := range testData {
		_, err := g.Verify(ctx, vlib.Input{NameStrings: []string{v.name}})
		assert.Nil(t, err)
	}
}

type mockVerifier struct{}

func (m mockVerifier) DataSources(ids ...int) ([]*vlib.DataSource, error) {
	var res []*vlib.DataSource
	return res, nil
}

func (m mockVerifier) MatchRecords(
	ctx context.Context,
	fmatches []mlib.Output,
	input vlib.Input,
) (map[string]*verifier.MatchRecord, error) {
	var res map[string]*verifier.MatchRecord
	return res, nil
}

type mockFacet struct{}

func (mf mockFacet) Search(
	ctx context.Context,
	inp search.Input,
) (map[string]*verifier.MatchRecord, error) {
	var res map[string]*verifier.MatchRecord
	return res, nil
}
