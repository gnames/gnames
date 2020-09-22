package score

import (
	"sort"

	"github.com/gnames/gnames/data"
	"github.com/gnames/gnames/domain/entity"
	// log "github.com/sirupsen/logrus"
)

// Calculate goes through entity.ResultData aggregated by a match and
// assigns each of them a score accoring to the scoring algorithms.
func Calculate(mr *data.MatchRecord) {
	for _, rd := range mr.MatchResults {
		score := Score{}
		score = score.rank(mr.CanonicalFull, rd.MatchedCanonicalFull,
			mr.Cardinality, rd.MatchedCardinality).
			curation(rd.DataSourceID, rd.CurationLevel).
			auth(mr.Authors, rd.MatchedAuthors, mr.Year, rd.MatchedYear).
			accepted(rd.ID, rd.CurrentRecordID).
			fuzzy(rd.EditDistance)

		rd.Score = score.Value
	}
	// sort according to the score
	mrs := mr.MatchResults
	sort.SliceStable(mrs, func(i, j int) bool {
		return mrs[i].Score > mrs[j].Score
	})
	mr.Sorted = true
}

// BestResult returns the highest runked entity.ResultData according to
// scoring algorithm.
func BestResult(mr *data.MatchRecord) *entity.ResultData {
	if mr.MatchResults == nil {
		var br *entity.ResultData
		return br
	}

	if !mr.Sorted {
		Calculate(mr)
	}
	return mr.MatchResults[0]
}

// PreferredResults returns the best scoring entity.ResultData for each of
// the preffered data-source. From 0 to 1 results per data-source are allowed.
func PreferredResults(
	sources []int,
	mr *data.MatchRecord) []*entity.ResultData {

	if mr.MatchResults == nil || len(mr.MatchResults) == 0 {
		var res []*entity.ResultData
		return res
	}

	if !mr.Sorted {
		Calculate(mr)
	}
	// maps a data-source ID to corresponding result data.
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
