package pgio

import (
	"database/sql"
	"testing"

	"github.com/gnames/gnames/pkg/ent/verif"
	mlib "github.com/gnames/gnlib/ent/matcher"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnparser"
	"github.com/stretchr/testify/assert"
)

func nullStr(s string) sql.NullString {
	return sql.NullString{String: s, Valid: true}
}

// Names imported with an older gnparser can keep a canonical_id even though
// the current parser rejects them. One such record must not break the batch.
func TestProduceResultDataUnparsableRecord(t *testing.T) {
	p := &pgio{dsm: map[int]*vlib.DataSource{1: {ID: 1, Title: "Test"}}}
	parser := gnparser.New(gnparser.NewConfig(gnparser.OptWithDetails(true)))

	const canID = "canonical-id"
	bad := "Erythrocebus (patas)Schreber,1774 Schreber, 1774"
	good := "Erythrocebus Trouessart, 1897"
	assert.False(t, parser.ParseName(bad).Parsed, "name must stay unparsable")

	vers := []*verifSQL{
		{CanonicalID: nullStr(canID), Name: nullStr(bad), DataSourceID: 1},
		{CanonicalID: nullStr(canID), Name: nullStr(good), DataSourceID: 1},
	}
	ms := verif.MatchSplit{Canonical: []*mlib.Match{{
		ID:         "input-id",
		Name:       "Erythrocebus",
		MatchType:  vlib.Exact,
		MatchItems: []mlib.MatchItem{{ID: canID, MatchType: vlib.Exact}},
	}}}

	res := p.produceResultData(ms, parser, vers, nil)

	assert.Len(t, res["input-id"].MatchResults, 1)
	assert.Equal(t, good, res["input-id"].MatchResults[0].MatchedName)
}

// An input name that cannot be parsed comes back without results.
func TestProduceResultDataUnparsableInput(t *testing.T) {
	p := &pgio{dsm: map[int]*vlib.DataSource{}}
	parser := gnparser.New(gnparser.NewConfig(gnparser.OptWithDetails(true)))

	ms := verif.MatchSplit{Canonical: []*mlib.Match{{
		ID:   "input-id",
		Name: "1234",
	}}}

	res := p.produceResultData(ms, parser, nil, nil)

	assert.Equal(t, "1234", res["input-id"].Name)
	assert.Empty(t, res["input-id"].MatchResults)
}
