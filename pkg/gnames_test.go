package gnames_test

import (
	"context"
	"testing"

	gnames "github.com/gnames/gnames/pkg"
	"github.com/gnames/gnames/pkg/config"
	"github.com/gnames/gnames/pkg/ent/verif"
	mlib "github.com/gnames/gnlib/ent/matcher"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnquery/ent/search"
	"github.com/stretchr/testify/assert"
)

func TestVerifier(t *testing.T) {
	var g gnames.GNames
	cfg := config.New()
	config.LoadEnv(&cfg)
	ctx := context.Background()
	vf := mockVerifier{}
	fct := mockFacet{}
	g = gnames.New(cfg, vf, fct)
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

func (m mockVerifier) DataSources(ids ...int) []*vlib.DataSource {
	var res []*vlib.DataSource
	return res
}

func (m mockVerifier) MatchRecords(
	ctx context.Context,
	fmatches []mlib.Match,
	input vlib.Input,
) (map[string]*verif.MatchRecord, error) {
	var res map[string]*verif.MatchRecord
	return res, nil
}

func (m mockVerifier) NameByID(
	nsi vlib.NameStringInput,
) (*verif.MatchRecord, error) {
	var res *verif.MatchRecord
	return res, nil
}

func (m mockVerifier) NameStringByID(s string) (string, error) {
	return "", nil
}

type mockFacet struct{}

func (mf mockFacet) AdvancedSearch(
	ctx context.Context,
	inp search.Input,
) (map[string]*verif.MatchRecord, error) {
	var res map[string]*verif.MatchRecord
	return res, nil
}
