package score

import (
	"fmt"
	"testing"

	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	s := score{}
	assert.Equal(t, "00000000_00000000_00000000_00000000", s.String())
}

func TestChain(t *testing.T) {
	s := score{}.
		rank("Aus bus var. cus", "Aus bus var. cus", 3, 3).
		fuzzy(2).
		curation(2, vlib.Curated).
		auth(
			[]string{"Hopkins", "L.", "Thomson"},
			[]string{"Thomson", "Linn."}, 1758, 1757,
		).
		accepted("12", "12").
		parsingQuality(3)
	assert.Equal(t, "10011010_11010000_00000000_00000000", s.String())
}

func TestRank(t *testing.T) {
	testData := []struct {
		desc, can1, can2 string
		card1, card2     int
		score            string
	}{
		{"partial", "Aus bus var. cus", "Aus bus", 3, 2, "00000000_00000000_00000000_00000000"},
		{"binomial", "Aus bus", "Aus bus", 2, 2, "00000000_00000000_00000000_00000000"},
		{"exact", "Aus bus var. cus", "Aus bus var. cus", 3, 3, "10000000_00000000_00000000_00000000"},
		{"no match", "Aus bus var. cus", "Aus bus f. cus", 3, 3, "00000000_00000000_00000000_00000000"},
		{"unknown", "Aus bus cus", "Aus bus f. cus", 3, 3, "01000000_00000000_00000000_00000000"},
		{"n/a", "Aus bus cus", "Aus bus f. cus", 3, 3, "01000000_00000000_00000000_00000000"},
		{"n/a", "Aus bus f. cus", "Aus bus cus", 3, 3, "01000000_00000000_00000000_00000000"},
	}
	for _, v := range testData {
		s := score{}
		assert.Equal(t, v.score, s.rank(v.can1, v.can2, v.card1, v.card2).String(), v.desc)
	}
}

func TestFuzzy(t *testing.T) {
	testData := []struct {
		desc     string
		editDist int
		score    string
	}{
		{"fuzzy1", 1, "00100000_00000000_00000000_00000000"},
		{"fuzzy2", 2, "00010000_00000000_00000000_00000000"},
		{"fuzzy3", 3, "00000000_00000000_00000000_00000000"},
		{"fuzzy4", 13, "00000000_00000000_00000000_00000000"},
		{"exact", 0, "00110000_00000000_00000000_00000000"},
	}
	for _, v := range testData {
		s := score{}
		assert.Equal(t, v.score, s.fuzzy(v.editDist).String(), v.desc)
	}
}

func TestFuzzyVal(t *testing.T) {
	testData := []struct {
		desc     string
		editDist int
		scoreVal float32
	}{
		{"fuzzy1", 1, 0.666},
		{"fuzzy2", 2, 0.333},
		{"fuzzy3", 3, 0},
		{"fuzzy4", 13, 0},
		{"fuzzy5", -1, 0},
		{"exact", 0, 1},
	}
	for _, v := range testData {
		s := score{}
		s = s.fuzzy(v.editDist)
		assert.InDelta(t, v.scoreVal, s.fuzzyVal(), 0.001)
	}
}

func TestCuration(t *testing.T) {
	testData := []struct {
		desc   string
		dsID   int
		curLev vlib.CurationLevel
		score  string
	}{
		{"no cur", 67, vlib.NotCurated, "00000000_00000000_00000000_00000000"},
		{"auto cur", 67, vlib.AutoCurated, "00000100_00000000_00000000_00000000"},
		{"cur", 67, vlib.Curated, "00001000_00000000_00000000_00000000"},
		{"CoL", 1, vlib.Curated, "00001100_00000000_00000000_00000000"},
	}
	for _, v := range testData {
		s := score{}
		assert.Equal(t, v.score, s.curation(v.dsID, v.curLev).String(), v.desc)
	}
}

func TestAuth(t *testing.T) {
	testData := []struct {
		desc         string
		auth1, auth2 []string
		year1, year2 int
		score        string
	}{
		{"empty1", []string{}, []string{}, 0, 0, "00000000_10000000_00000000_00000000"},
		{"empty2", []string{"L."}, []string{}, 1758, 0, "00000000_10000000_00000000_00000000"},
		{"empty3", []string{}, []string{"L."}, 0, 1758, "00000000_10000000_00000000_00000000"},
		{"no match1", []string{"Banks"}, []string{"L."}, 0, 0, "00000000_00000000_00000000_00000000"},
		{"no match2", []string{"L."}, []string{"Banks"}, 1758, 1758, "00000000_00000000_00000000_00000000"},
		{"overlap", []string{"Tomm.", "L.", "Banks", "Muetze"}, []string{"Kuntze", "Linn", "Hopkins"}, 1758, 1758, "00000010_10000000_00000000_00000000"},
		{"full subset, yes yr", []string{"Hopkins", "L.", "Thomson"}, []string{"Thomson", "Linn."}, 1758, 1758, "00000011_00000000_00000000_00000000"},
		{"full subset, aprx yr1", []string{"Hopkins", "L.", "Thomson"}, []string{"Thomson", "Linn."}, 1757, 1758, "00000010_10000000_00000000_00000000"},
		{"full subset, aprx yr2", []string{"L.", "Thomson"}, []string{"Thomson", "Linn.", "Hopkins"}, 1757, 1756, "00000010_10000000_00000000_00000000"},
		{"full subset, n/a yr1", []string{"L.", "Thomson"}, []string{"Thomson", "Linn.", "Hopkins"}, 0, 1756, "00000010_00000000_00000000_00000000"},
		{"full subset, n/a yr2", []string{"L.", "Thomson"}, []string{"Thomson", "Linn.", "Hopkins"}, 1756, 0, "00000010_00000000_00000000_00000000"},
		{"full subset, no yr", []string{"L.", "Thomson"}, []string{"Thomson", "Linn.", "Hopkins"}, 1756, 1800, "00000001_10000000_00000000_00000000"},
		{"match, yes yr", []string{"L.", "Thomson"}, []string{"Linn", "Thomson"}, 1800, 1800, "00000011_10000000_00000000_00000000"},
		{"match, aprx yr", []string{"Herenson", "Thomson"}, []string{"Thomson", "H."}, 1799, 1800, "00000011_00000000_00000000_00000000"},
		{"match, n/a yr", []string{"Herenson", "Thomson"}, []string{"Thomson", "H."}, 0, 0, "00000010_10000000_00000000_00000000"},
		{"match, bad yr", []string{"Herenson", "Thomson"}, []string{"Thomson", "H."}, 1750, 1755, "00000010_00000000_00000000_00000000"},
	}
	for _, v := range testData {
		s := score{}
		assert.Equal(t, v.score, s.auth(v.auth1, v.auth2, v.year1, v.year2).String(), v.desc)
	}
}

func TestAccepted(t *testing.T) {
	testData := []struct{ desc, recordID, acceptedID, score string }{
		{"synonym", "123", "234", "00000000_00000000_00000000_00000000"},
		{"accepted1", "123", "123", "00000000_01000000_00000000_00000000"},
		{"accepted2", "123", "", "00000000_01000000_00000000_00000000"},
	}
	for _, v := range testData {
		s := score{}
		assert.Equal(t,
			v.score, s.accepted(v.recordID, v.acceptedID).String(), v.desc)
	}
}

func TestParserQuality(t *testing.T) {
	testData := []struct {
		desc    string
		quality int
		score   string
	}{
		{"no parse", 0, "00000000_00000000_00000000_00000000"},
		{"clean", 1, "00000000_00110000_00000000_00000000"},
		{"some problems", 2, "00000000_00100000_00000000_00000000"},
		{"big problems", 3, "00000000_00010000_00000000_00000000"},
	}
	for _, v := range testData {
		s := score{}
		assert.Equal(t, v.score, s.parsingQuality(v.quality).String(), v.desc)
	}
}

func TestCompareAuth(t *testing.T) {
	testData := []struct {
		desc, au1, au2, res string
	}{
		{"no match2", "L", "Banks", "false|false"},
		{"no match2", "Banks", "L", "false|true"},
		{"no match2", "Banks", "B", "true|false"},
		{"no match2", "Banks", "Banz", "false|true"},
		{"no match2", "Banks", "Banks", "true|false"},
	}
	for _, v := range testData {
		match, giveup := compareAuth(v.au1, v.au2)
		res := fmt.Sprintf("%v|%v", match, giveup)
		assert.Equal(t, v.res, res, v.desc)
	}
}

func TestAuthNormalize(t *testing.T) {
	testData := []struct {
		desc, auth, res string
	}{
		{"empty", "", ""},
		{"abbr1", "L.", "L"},
		{"abbr2", "Linn.", "Linn"},
		{"initial1", "A. Linn.", "Linn"},
		{"initial2", "A. B. Lin", "Lin"},
		{"initial3", "A. B.", ""},
		{"two words", "A. B. Koza Koza", "Koza Koza"},
	}
	for _, v := range testData {
		assert.Equal(t, v.res, authNormalize(v.auth), v.desc)
	}
}

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
		s := score{value: v.score}
		res := s.details()
		assert.Equal(t, v.rank, res.InfraSpecificRankScore, v.msg)
		assert.InDelta(t, res.FuzzyLessScore, v.fuzzy, 0.01, v.msg)
		assert.InDelta(t, res.CuratedDataScore, v.curat, 0.01, v.msg)
		assert.InDelta(t, res.AuthorMatchScore, v.auth, 0.01, v.msg)
		assert.Equal(t, v.accept, res.AcceptedNameScore, v.msg)
		assert.InDelta(t, res.ParsingQualityScore, v.pars, 0.01, v.msg)
	}
}
