package gnames

import (
	"context"
	"log/slog"
	"slices"

	"github.com/gnames/gnames/pkg/ent/verif"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnquery/ent/search"
)

// Search finds scientific names that match the provided partial
// information. For example, it can handle cases where the genus is
// abbreviated or only part of the specific epithet is known.
// It can also utilize year and year range information to narrow
// down the search.
func (g gnames) Search(
	ctx context.Context,
	input search.Input,
) search.Output {
	input.Query = input.ToQuery()
	slog.Info("Search", "query", input.Query)

	res := search.Output{Meta: search.Meta{Input: input}}
	matchRecords, err := g.sr.AdvancedSearch(ctx, input)
	if err != nil {
		res.Error = err.Error()
	}
	res.NamesNumber = len(matchRecords)

	sortedNames := sortNames(matchRecords)
	resNames := make([]vlib.Name, len(matchRecords))

	for i, v := range sortedNames {
		mr := matchRecords[v]
		resNames[i] = outputName(mr, input.WithAllMatches)
	}

	res.Names = resNames
	return res
}

func sortNames(mrs map[string]*verif.MatchRecord) []string {
	res := make([]string, len(mrs))
	var count int
	for k := range mrs {
		res[count] = k
		count++
	}
	slices.Sort(res)
	return res
}
