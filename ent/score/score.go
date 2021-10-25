package score

import (
	"fmt"
	"sort"

	"github.com/gnames/gnames/ent/verifier"
	vlib "github.com/gnames/gnlib/ent/verifier"
)

// NewScore returns an implementation of the Score interface.
func NewScore() Score {
	return score{}
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

// SortResults goes through vlib.ResultData aggregated by a match and
// assigns each of them a score accoring to the scoring algorithms.
func (s score) SortResults(mr *verifier.MatchRecord) {
	for _, rd := range mr.MatchResults {
		s = s.rank(mr.CanonicalFull, rd.MatchedCanonicalFull,
			mr.Cardinality, rd.MatchedCardinality).
			fuzzy(rd.EditDistance).
			curation(rd.DataSourceID, rd.Curation).
			auth(mr.Authors, rd.MatchedAuthors, mr.Year, rd.MatchedYear).
			accepted(rd.RecordID, rd.CurrentRecordID).
			parsingQuality(rd.ParsingQuality)
		rd.Score = s.value
		s.value = 0
	}
	// Sort (in reverse) according to the score. First element has
	// the highest score, the last has the lowest.
	mrs := mr.MatchResults
	sort.SliceStable(mrs, func(i, j int) bool {
		return mrs[i].Score > mrs[j].Score
	})
	mr.Sorted = true
}

// BestResult returns the highest runked vlib.ResultData according to
// scoring algorithm.
func (s score) BestResult(mr *verifier.MatchRecord) *vlib.ResultData {
	if mr.MatchResults == nil {
		return nil
	}

	if !mr.Sorted {
		s.SortResults(mr)
	}
	return mr.MatchResults[0]
}

// PreferredResults returns the best scoring vlib.ResultData for each of
// the preffered data-source. From 0 to 1 results per data-source are allowed.
func (s score) PreferredResults(
	sources []int,
	mr *verifier.MatchRecord,
	allMatches bool,
) []*vlib.ResultData {
	if mr.MatchResults == nil || len(mr.MatchResults) == 0 {
		return nil
	}

	allSources := len(sources) == 1 && sources[0] == 0

	if !mr.Sorted {
		s.SortResults(mr)
	}
	// maps a data-source ID to corresponding result data.
	sourceMap := make(map[int][]*vlib.ResultData)
	for _, v := range sources {
		sourceMap[v] = nil
	}
	var resLen int
	for _, v := range mr.MatchResults {
		dsID := v.DataSourceID
		if data, ok := sourceMap[dsID]; (ok || allSources) && (allMatches || data == nil) {
			resLen++
			sourceMap[dsID] = append(sourceMap[dsID], v)
		}
	}
	res := make([]*vlib.ResultData, 0, resLen)
	for _, v := range sourceMap {
		if v != nil {
			res = append(res, v...)
		}
	}
	return res
}
