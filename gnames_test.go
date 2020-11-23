package gnames_test

import (
	"testing"

	"github.com/gnames/gnames"
	"github.com/gnames/gnames/config"
	"github.com/gnames/gnames/entity/verifier"
	mlib "github.com/gnames/gnlib/domain/entity/matcher"
	vlib "github.com/gnames/gnlib/domain/entity/verifier"
	"github.com/stretchr/testify/assert"
)

func TestVer(t *testing.T) {
	var g gnames.GNames
	cfg := config.NewConfig()
	vf := mockVerifier{}
	g = gnames.NewGNames(cfg, vf)
	testData := []struct {
		name string
	}{
		{"Bubo bubo"},
	}
	for _, v := range testData {
		_, err := g.Verify(vlib.VerifyParams{NameStrings: []string{v.name}})
		assert.Nil(t, err)
	}
}

type mockVerifier struct{}

func (m mockVerifier) DataSources(ids ...int) ([]*vlib.DataSource, error) {
	var res []*vlib.DataSource
	return res, nil
}

func (m mockVerifier) MatchRecords(matches []*mlib.Match) (map[string]*verifier.MatchRecord, error) {
	var res map[string]*verifier.MatchRecord
	return res, nil
}
