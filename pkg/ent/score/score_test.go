package score_test

import (
	"testing"

	"github.com/gnames/gnames/pkg/ent/score"
	"github.com/gnames/gnames/pkg/ent/verif"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/stretchr/testify/assert"
)

func TestSortRecords(t *testing.T) {
	mr := &matchRec
	s := score.New()
	s.SortResults(mr)
	assert.Equal(t, 1, mr.MatchResults[0].DataSourceID)
	assert.Equal(t, "3529384", mr.MatchResults[0].RecordID)
	assert.InDelta(t, 9.414964025906754, mr.MatchResults[0].SortScore, 0.00001)

	assert.Equal(t, 1, mr.MatchResults[1].DataSourceID)
	assert.Equal(t, "3562751", mr.MatchResults[1].RecordID)
	assert.InDelta(t, 9.414613576436937, mr.MatchResults[1].SortScore, 0.00001)

	assert.Equal(t, 11, mr.MatchResults[2].DataSourceID)
	assert.Equal(t, "8638411", mr.MatchResults[2].RecordID)
	assert.InDelta(t, 9.41356052807642, mr.MatchResults[2].SortScore, 0.00001)

	assert.Equal(t, 169, mr.MatchResults[3].DataSourceID)
	assert.Equal(t, "95877520", mr.MatchResults[3].RecordID)
	assert.InDelta(t, 9.41003181086182, mr.MatchResults[3].SortScore, 0.00001)
}

var matchRec = verif.MatchRecord{
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
