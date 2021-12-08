package score_test

import (
	"testing"

	"github.com/gnames/gnames/ent/score"
	"github.com/gnames/gnames/ent/verifier"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/stretchr/testify/assert"
)

func TestScoreDetails(t *testing.T) {
	tests := []struct {
		msg                                    string
		score                                  uint32
		rank, fuzzy, curat, auth, accept, pars float32
	}{
		{
			"empty",
			uint32(0b00000000_00000000_00000000_00000000),
			0, 0, 0, 0, 0, 0,
		},
		{
			"full",
			uint32(0b10111111_11111111_11111111_11111111),
			1, 1, 1, 1, 1, 1,
		},
		{
			"rank",
			uint32(0b01000000_00000000_00000000_00000000),
			0.5, 0, 0, 0, 0, 0,
		},
		{
			"fuzzy",
			uint32(0b00010000_00000000_00000000_00000000),
			0, 0.33, 0, 0, 0, 0,
		},
		{
			"curated",
			uint32(0b00000100_00000000_00000000_00000000),
			0, 0, 0.33, 0, 0, 0,
		},
		{
			"auth",
			uint32(0b00000000_10000000_00000000_00000000),
			0, 0, 0, 0.1428, 0, 0,
		},
		{
			"accept",
			uint32(0b00000000_01000000_00000000_00000000),
			0, 0, 0, 0, 1, 0,
		},
		{
			"parsed",
			uint32(0b00000000_00010000_00000000_00000000),
			0, 0, 0, 0, 0, 0.33,
		},
	}

	for _, v := range tests {
		s := score.New(v.score)
		res := s.Details()
		assert.Equal(t, res.InfraSpecificRankScore, v.rank, v.msg)
		assert.InDelta(t, res.FuzzynessScore, v.fuzzy, 0.01, v.msg)
		assert.InDelta(t, res.CuratedDataScore, v.curat, 0.01, v.msg)
		assert.InDelta(t, res.AuthorMatchScore, v.auth, 0.01, v.msg)
		assert.Equal(t, res.AcceptedNameScore, v.accept, v.msg)
		assert.InDelta(t, res.ParsingQualityScore, v.pars, 0.01, v.msg)
	}
}

func TestSortRecords(t *testing.T) {
	mr := &matchRec
	s := score.New()
	s.SortResults(mr)
	assert.Equal(t, mr.MatchResults[0].DataSourceID, 1)
	assert.Equal(t, mr.MatchResults[0].RecordID, "3529384")
	assert.Equal(t, int(mr.MatchResults[0].Score), 1022361600)

	assert.Equal(t, mr.MatchResults[1].DataSourceID, 1)
	assert.Equal(t, mr.MatchResults[1].RecordID, "3562751")
	assert.Equal(t, int(mr.MatchResults[1].Score), 1018167296)

	assert.Equal(t, mr.MatchResults[2].DataSourceID, 11)
	assert.Equal(t, mr.MatchResults[2].RecordID, "8638411")
	assert.Equal(t, int(mr.MatchResults[2].Score), 888143872)

	assert.Equal(t, mr.MatchResults[3].DataSourceID, 169)
	assert.Equal(t, mr.MatchResults[3].RecordID, "95877520")
	assert.Equal(t, int(mr.MatchResults[3].Score), 821035008)
}

var matchRec = verifier.MatchRecord{
	InputID:         "4c8848f2-7271-588c-ba81-e4d5efcc1e92",
	Input:           "Pisonia grandis",
	Cardinality:     2,
	CanonicalSimple: "Pisonia grandis",
	CanonicalFull:   "Pisonia grandis",
	Authors:         nil,
	Year:            0,
	MatchType:       vlib.Exact,
	Curation:        vlib.Curated,
	DataSourcesNum:  18,
	MatchResults: []*vlib.ResultData{
		{
			DataSourceID:           169,
			DataSourceTitleShort:   "uBio NameBank",
			Curation:               vlib.NotCurated,
			RecordID:               "95877520",
			MatchedName:            "Pisonia grandis",
			MatchedCardinality:     2,
			ParsingQuality:         1,
			MatchedCanonicalSimple: "Pisonia grandis",
			MatchedCanonicalFull:   "Pisonia grandis",
			MatchedAuthors:         nil,
			MatchedYear:            0,
			CurrentRecordID:        "95877520",
			CurrentName:            "Pisonia grandis",
			CurrentCardinality:     2,
			CurrentCanonicalSimple: "Pisonia grandis",
			CurrentCanonicalFull:   "Pisonia grandis",
			IsSynonym:              false,
		},
		{
			DataSourceID:           1,
			DataSourceTitleShort:   "Catalogue of Life",
			Curation:               vlib.Curated,
			RecordID:               "3529384",
			MatchedName:            "Pisonia grandis R. Br.",
			MatchedCardinality:     2,
			ParsingQuality:         1,
			MatchedCanonicalSimple: "Pisonia grandis",
			MatchedCanonicalFull:   "Pisonia grandis",
			MatchedAuthors:         []string{"R.", "Br."},
			MatchedYear:            0,
			CurrentRecordID:        "3529384",
			CurrentName:            "Pisonia grandis R. Br.",
			CurrentCardinality:     2,
			CurrentCanonicalSimple: "Pisonia grandis",
			CurrentCanonicalFull:   "Pisonia grandis",
			IsSynonym:              false,
		},
		{
			DataSourceID:           11,
			DataSourceTitleShort:   "GBIF Backbone Taxonomy",
			Curation:               vlib.AutoCurated,
			RecordID:               "8638411",
			MatchedName:            "Pisonia grandis A. Cunn.",
			MatchedCardinality:     2,
			ParsingQuality:         1,
			MatchedCanonicalSimple: "Pisonia grandis",
			MatchedCanonicalFull:   "Pisonia grandis",
			MatchedAuthors:         []string{"A.", "Cunn."},
			MatchedYear:            0,
			CurrentRecordID:        "8638411",
			CurrentName:            "Pisonia umbellifera Seem.",
			CurrentCardinality:     2,
			CurrentCanonicalSimple: "Pisonia umbellifera",
			CurrentCanonicalFull:   "Pisonia umbellifera",
			IsSynonym:              true,
		},
		{
			DataSourceID:           1,
			DataSourceTitleShort:   "Catalogue of Life",
			Curation:               vlib.Curated,
			RecordID:               "3562751",
			MatchedName:            "Pisonia grandis A. Cunn. ex Hook. fil.",
			MatchedCardinality:     2,
			ParsingQuality:         1,
			MatchedCanonicalSimple: "Pisonia grandis",
			MatchedCanonicalFull:   "Pisonia grandis",
			MatchedAuthors:         []string{"A.", "Cunn.", "Hook.", "fil."},
			MatchedYear:            0,
			CurrentRecordID:        "3529412",
			CurrentName:            "Pisonia umbellifera (J. & G. Forst.) Seem.",
			CurrentCardinality:     2,
			CurrentCanonicalSimple: "Pisonia umbellifera",
			CurrentCanonicalFull:   "Pisonia umbellifera",
			IsSynonym:              true,
		},
	},
	Sorted: false,
}
