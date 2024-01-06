package score

import (
	"cmp"
	"fmt"
	"math"
	"slices"

	"github.com/gnames/gnames/pkg/ent/verif"
	vlib "github.com/gnames/gnlib/ent/verifier"
)

// New returns an implementation of the Score interface.
func New(value ...uint32) Score {
	res := score{}
	if len(value) == 1 {
		res.value = value[0]
	}
	return res
}

// String returns a string representation of a score as a set of bits with
// every byte (8 bits) separated by an underscore.
func (s score) String() string {
	str := fmt.Sprintf("%032b", s.value)
	res := make([]byte, 35)
	offset := 0
	for i, v := range []byte(str) {
		res[i+offset] = v
		if (i+1)%8 == 0 && (i+1)%32 != 0 {
			offset++
			res[i+offset] = '_'
		}
	}
	return string(res)
}

// sortScore represents the score as a float64 number.
func (s score) sortScore() float64 {
	return math.Log10(float64(s.value))
}

// SortResults goes through vlib.ResultData aggregated by a match and
// assigns each of them a score accoring to the scoring algorithms.
func (s score) SortResults(mr *verif.MatchRecord) {
	for _, rd := range mr.MatchResults {
		s = s.cardinality(mr.Cardinality, rd.MatchedCardinality).
			rank(mr.CanonicalFull, rd.MatchedCanonicalFull,
				mr.Cardinality, rd.MatchedCardinality).
			fuzzy(rd.EditDistance).
			curation(rd.DataSourceID, rd.Curation).
			auth(mr.Authors, rd.MatchedAuthors, mr.Year, rd.MatchedYear).
			accepted(rd.RecordID, rd.CurrentRecordID).
			parsingQuality(rd.ParsingQuality)
		rd.SortScore = s.sortScore()
		rd.ScoreDetails = s.details()
		s.value = 0
	}
	// Sort (in reverse) according to the score. First element has
	// the highest score, the last has the lowest.
	mrs := mr.MatchResults
	slices.SortStableFunc(mrs, func(a, b *vlib.ResultData) int {
		return cmp.Compare(b.SortScore, a.SortScore)
	})
	mr.Sorted = true
}

// Results returns the best scoring vlib.ResultData for each of
// the preffered data-source. From 0 to 1 results per data-source are allowed.
func (s score) Results(
	mr *verif.MatchRecord,
) []*vlib.ResultData {

	if !mr.Sorted {
		s.SortResults(mr)
	}

	return mr.MatchResults
}

// ScoreDetails converts the scoreinteger to human-readable ScoreDetails.
func (s score) details() vlib.ScoreDetails {
	res := vlib.ScoreDetails{
		CardinalityScore:       s.cardinalityVal(),
		InfraSpecificRankScore: s.rankVal(),
		FuzzyLessScore:         s.fuzzyVal(),
		CuratedDataScore:       s.curationVal(),
		AuthorMatchScore:       s.authVal(),
		AcceptedNameScore:      s.acceptedVal(),
		ParsingQualityScore:    s.parsingQualityVal(),
	}
	return res
}
