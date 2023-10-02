package gnames

import (
	"context"
	"slices"

	"github.com/gnames/gnames/pkg/ent/verifier"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnquery/ent/search"
	"github.com/rs/zerolog/log"
)

func (g gnames) Search(
	ctx context.Context,
	input search.Input,
) search.Output {
	input.Query = input.ToQuery()
	log.Info().Str("action", "search").Str("query", input.Query)

	res := search.Output{Meta: search.Meta{Input: input}}
	matchRecords, err := g.facet.Search(ctx, input)
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

func sortNames(mrs map[string]*verifier.MatchRecord) []string {
	res := make([]string, len(mrs))
	var count int
	for k := range mrs {
		res[count] = k
		count++
	}
	slices.Sort(res)
	return res
}
