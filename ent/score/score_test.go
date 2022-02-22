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
		assert.Equal(t, v.rank, res.InfraSpecificRankScore, v.msg)
		assert.InDelta(t, res.FuzzyLessScore, v.fuzzy, 0.01, v.msg)
		assert.InDelta(t, res.CuratedDataScore, v.curat, 0.01, v.msg)
		assert.InDelta(t, res.AuthorMatchScore, v.auth, 0.01, v.msg)
		assert.Equal(t, v.accept, res.AcceptedNameScore, v.msg)
		assert.InDelta(t, res.ParsingQualityScore, v.pars, 0.01, v.msg)
	}
}

func TestSortRecords(t *testing.T) {
	mr := &matchRec
	s := score.New()
	s.SortResults(mr)
	assert.Equal(t, 1, mr.MatchResults[0].DataSourceID)
	assert.Equal(t, "3529384", mr.MatchResults[0].RecordID)
	assert.Equal(t, 1030750208, int(mr.MatchResults[0].Score))

	assert.Equal(t, 1, mr.MatchResults[1].DataSourceID)
	assert.Equal(t, "3562751", mr.MatchResults[1].RecordID)
	assert.Equal(t, 1026555904, int(mr.MatchResults[1].Score))

	assert.Equal(t, 11, mr.MatchResults[2].DataSourceID)
	assert.Equal(t, "8638411", mr.MatchResults[2].RecordID)
	assert.Equal(t, 896532480, int(mr.MatchResults[2].Score))

	assert.Equal(t, 169, mr.MatchResults[3].DataSourceID)
	assert.Equal(t, "95877520", mr.MatchResults[3].RecordID)
	assert.Equal(t, 821035008, int(mr.MatchResults[3].Score))
}

var matchRec = verifier.MatchRecord{
	ID:              "4c8848f2-7271-588c-ba81-e4d5efcc1e92",
	Name:            "Pisonia grandis",
	Cardinality:     2,
	CanonicalSimple: "Pisonia grandis",
	CanonicalFull:   "Pisonia grandis",
	Authors:         nil,
	Year:            0,
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
