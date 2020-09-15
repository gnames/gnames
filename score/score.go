// Package score provides a single number that allows to sort matching results
// according to their relevancy.
// Score contains 32bits. They are distributed between different criteria the
// following way (from the left):

// xx000000_00000000_00000000_00000000: matching rank for ranked infraspecies
// 00 - rank does not match
//    `Aus bus var. cus` vs `Aus bus f. cus`
// 01 - match is unknown
//    `Aus bus cus` vs `Aus bus f. cus`
//    `Aus bus` vs `Aus bus`
// 10 - rank matches
//    `Aus bus var. cus` vs `Aus bus var. cus`
//
// 00xx0000_00000000_00000000_00000000: curation
// 00 - uncurated sources only
// 01 - auto-curated sources
// 10 - human-curated sources
// 11 - Catalogue of Life
//
// 0000xx00_00000000_00000000_00000000: matching authorship
// 00 - authorship does not match
//    `Aus bus Linn.` vs `Aus bus Banks`
// 01 - match is unknown
//    `Aus bus Linn` vs `Aus bus`
// 10 - authorship or year matches
//    `Aus bus L.` vs `Aus bus Linn.`
// 11 - authorship and year matches (one year difference is accepted)
//     `Aus bus L. 1765` vs `Aus bus Linn. 1766`
//
package score

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gnames/gnames/data"
	"github.com/gnames/gnames/domain/entity"
)

type Score struct {
	Value uint32
}

func (s Score) String() string {
	return fmt.Sprintf("%032b: %d", s.Value, s.Value)
}

func (s Score) Calculate(mr *data.MatchRecord, rd *entity.ResultData) uint32 {
	score := s.rank(mr.CanonicalFull, rd.MatchedCanonicalFull,
		mr.Cardinality, rd.MatchedCardinality).
		curation(rd.DataSourceID, rd.CurationLevel).
		auth(mr.Authors, rd.MatchedAuthors, mr.Year, rd.MatchedYear)
	return score.Value
}

func (s Score) Sort(mr *data.MatchRecord) {
	mrs := mr.MatchResults
	sort.SliceStable(mrs, func(i, j int) bool {
		return mrs[i].Score < mrs[j].Score
	})
	mr.Sorted = true
}

func (s Score) BestResult(mr *data.MatchRecord) *entity.ResultData {
	if mr.MatchResults == nil {
		var br *entity.ResultData
		return br
	}

	if !mr.Sorted {
		s.Sort(mr)
	}
	return mr.MatchResults[0]
}

func (s Score) PreferredResults(
	sources []int,
	mr *data.MatchRecord) []*entity.ResultData {

	if mr.MatchResults == nil || len(mr.MatchResults) == 0 {
		var res []*entity.ResultData
		return res
	}

	if !mr.Sorted {
		s.Sort(mr)
	}
	sourceMap := make(map[int]*entity.ResultData)
	for _, v := range sources {
		sourceMap[v] = nil
	}
	for _, v := range mr.MatchResults {
		if datum, ok := sourceMap[v.DataSourceID]; ok && datum == nil {
			sourceMap[v.DataSourceID] = v
		}
	}
	res := make([]*entity.ResultData, 0, len(sources))
	for _, v := range sourceMap {
		if v != nil {
			res = append(res, v)
		}
	}
	return res
}

// rank checks if infraspecific canonical forms contain the same ranks. If they
// do the score is 2, if comparison cannot be done the score is 1, and if the
// ranks are different, the score is 0. 2 bits, shift 30
func (s Score) rank(can1, can2 string, card1, card2 int) Score {
	shift := 30
	i := s.Value

	if card1 < 3 || card1 != card2 ||
		!strings.Contains(can1, ".") || !strings.Contains(can2, ".") {
		s.Value = (i | uint32(0b01<<shift))
		return s
	}

	if can1 == can2 {
		s.Value = (i | uint32(0b10<<shift))
	}
	return s
}

// Sort by curation level of data-sources
func (s Score) curation(dataSourceID int, curationLevel entity.CurationLevel) Score {
	shift := 28
	i := uint32(curationLevel)
	if dataSourceID == 1 {
		i = 3
	}
	s.Value = (s.Value | uint32(i<<shift))
	return s
}
