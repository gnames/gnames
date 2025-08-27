package gnames

import (
	"context"
	"fmt"
	"log/slog"
	"slices"

	"github.com/gnames/gnames/pkg/ent/score"
	"github.com/gnames/gnames/pkg/ent/verif"
	mlib "github.com/gnames/gnlib/ent/matcher"
	vlib "github.com/gnames/gnlib/ent/verifier"
	gncfg "github.com/gnames/gnmatcher/pkg/config"
	"github.com/gnames/gnparser/ent/str"
	"github.com/gnames/gnstats/ent/stats"
	"github.com/gnames/gnuuid"
)

func (g gnames) DataSources(ids ...int) []*vlib.DataSource {
	return g.vf.DataSources(ids...)
}

func (g gnames) Verify(
	ctx context.Context,
	input vlib.Input,
) (vlib.Output, error) {
	var errString string

	// trim names input to 50 when vernaculars are given
	if len(input.Vernaculars) > 0 && len(input.NameStrings) > 50 {
		input.NameStrings = input.NameStrings[:50]
	}

	namesRes := make([]vlib.Name, len(input.NameStrings))

	matchRecords, matchOut, err := g.getMatchRecords(ctx, input)
	if err != nil {
		// TODO fix this
		errString = err.Error()
	}

	for i, v := range matchOut.Matches {
		if mr, ok := matchRecords[v.ID]; ok {
			namesRes[i] = outputName(mr, input.WithAllMatches)
			namesRes[i].Error = errString
			if input.WithCapitalization {
				namesRes[i].Name = input.NameStrings[i]
				namesRes[i].ID = gnuuid.New(namesRes[i].Name).String()
			}
		} else {
			slog.Warn("Cannot find record for name", "name", v.Name)
		}
	}
	if len(input.Vernaculars) > 0 {
		namesRes, err = g.vern.AddVernacularNames(input.Vernaculars, namesRes)
		if err != nil {
			// TODO fix this
			errString = err.Error()
		}
	}
	res := vlib.Output{Meta: meta(input, namesRes), Names: namesRes}
	return res, nil
}

func outputName(mr *verif.MatchRecord, allMatches bool) vlib.Name {
	s := score.New()
	s.SortResults(mr)
	item := vlib.Name{
		ID:                 mr.ID,
		Name:               mr.Name,
		DataSourcesNum:     mr.DataSourcesNum,
		DataSourcesDetails: mr.DataSourcesDetails,
		Cardinality:        mr.Cardinality,
		OverloadDetected:   overloadTxt(mr),
	}

	results := s.Results(mr)
	if len(results) == 0 {
		item.OverloadDetected = ""
		return item
	}

	bestResult := results[0]
	item.Curation = bestResult.Curation
	item.MatchType = bestResult.MatchType

	if allMatches {
		item.Results = results
		return item
	}

	item.BestResult = bestResult
	item.DataSourcesIDs = getDataSourcesIDs(results)
	item.DataSourcesNum = len(item.DataSourcesIDs)
	return item
}

func meta(input vlib.Input, names []vlib.Name) vlib.Meta {
	allSources := len(input.DataSources) == 1 && input.DataSources[0] == 0
	hs := make([]stats.Hierarchy, 0, len(names))
	ids := make(map[string]struct{})
	for i := range names {
		if _, ok := ids[names[i].ID]; ok {
			continue
		}
		if names[i].BestResult == nil || names[i].BestResult.DataSourceID != 1 {
			continue
		}

		ids[names[i].ID] = struct{}{}
		hs = append(hs, names[i])
	}
	var c stats.Stats
	var ks []vlib.Kingdom

	if input.WithStats {
		c = stats.New(hs, input.MainTaxonThreshold)
		ks = make([]vlib.Kingdom, len(c.Kingdoms))
		for i, v := range c.Kingdoms {
			ks[i] = vlib.Kingdom{
				KingdomName: v.Name,
				NamesNumber: v.NamesNum,
				Percentage:  v.Percentage,
			}
		}
	}
	res := vlib.Meta{
		NamesNumber:             len(input.NameStrings),
		WithAllSources:          allSources,
		WithAllMatches:          input.WithAllMatches,
		WithStats:               input.WithStats,
		WithCapitalization:      input.WithCapitalization,
		WithSpeciesGroup:        input.WithSpeciesGroup,
		WithRelaxedFuzzyMatch:   input.WithRelaxedFuzzyMatch,
		WithUninomialFuzzyMatch: input.WithUninomialFuzzyMatch,
		MainTaxonThreshold:      input.MainTaxonThreshold,
		DataSources:             input.DataSources,
		MainTaxon:               c.MainTaxon.Name,
		MainTaxonPercentage:     c.MainTaxonPercentage,
		StatsNamesNum:           len(hs),
		Kingdom:                 c.Kingdom.Name,
		KingdomPercentage:       c.KingdomPercentage,
		Kingdoms:                ks,
	}
	return res
}

func (g gnames) getMatchRecords(
	ctx context.Context,
	input vlib.Input,
) (map[string]*verif.MatchRecord, mlib.Output, error) {

	namesNum := len(input.NameStrings)
	if namesNum > 0 {
		slog.Info("Verifying",
			slog.Int("namesNum", len(input.NameStrings)),
			slog.String("example", input.NameStrings[0]),
			slog.Bool("withAllMatches", input.WithAllMatches),
		)
	}

	var matchOut mlib.Output
	var opts []gncfg.Option
	if input.WithSpeciesGroup {
		opts = append(opts, gncfg.OptWithSpeciesGroup(true))
	}
	if input.WithRelaxedFuzzyMatch {
		opts = append(opts, gncfg.OptWithRelaxedFuzzyMatch(true))
	}
	if input.WithUninomialFuzzyMatch {
		opts = append(opts, gncfg.OptWithUninomialFuzzyMatch(true))
	}
	if len(input.DataSources) > 0 {
		opts = append(opts, gncfg.OptDataSources(input.DataSources))
	}

	if input.WithCapitalization {
		names := make([]string, len(input.NameStrings))
		for i := range input.NameStrings {
			names[i] = str.CapitalizeName(input.NameStrings[i])
		}
		matchOut = g.matcher.MatchNames(names, opts...)
	} else {
		matchOut = g.matcher.MatchNames(input.NameStrings, opts...)
	}

	mRec, err := g.vf.MatchRecords(ctx, matchOut.Matches, input)
	if err != nil {
		return mRec, matchOut, fmt.Errorf("gnames.getMatchRecords: cannot match records: %w", err)
	}

	return mRec, matchOut, nil
}

func overloadTxt(mr *verif.MatchRecord) string {
	if !mr.Overload {
		return ""
	}
	return "Too many records (possibly strains), some results are truncated"
}

func getDataSourcesIDs(rs []*vlib.ResultData) []int {
	resMap := make(map[int]struct{})
	for _, v := range rs {
		resMap[v.DataSourceID] = struct{}{}
	}
	res := make([]int, len(resMap))
	var count int
	for k := range resMap {
		res[count] = k
		count++
	}
	slices.Sort(res)
	return res
}
