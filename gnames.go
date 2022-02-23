package gnames

import (
	"context"
	"sort"
	"unicode"

	"github.com/gnames/gnames/config"
	"github.com/gnames/gnames/ent/facet"
	"github.com/gnames/gnames/ent/score"
	"github.com/gnames/gnames/ent/verifier"
	"github.com/gnames/gnames/io/matcher"
	gnctx "github.com/gnames/gnlib/ent/context"
	"github.com/gnames/gnlib/ent/gnvers"
	mlib "github.com/gnames/gnlib/ent/matcher"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnmatcher"
	"github.com/gnames/gnparser/ent/str"
	"github.com/gnames/gnquery/ent/search"
	"github.com/gnames/gnuuid"
	"github.com/rs/zerolog/log"
)

type gnames struct {
	cfg     config.Config
	vf      verifier.Verifier
	facet   facet.Facet
	matcher gnmatcher.GNmatcher
}

// NewGNames is a constructor that returns implmentation of GNames interface.
func NewGNames(
	cfg config.Config,
	vf verifier.Verifier,
	fc facet.Facet,
) GNames {
	return gnames{
		cfg:     cfg,
		vf:      vf,
		facet:   fc,
		matcher: matcher.New(cfg.MatcherURL),
	}
}

func (g gnames) GetVersion() gnvers.Version {
	return gnvers.Version{
		Version: Version,
		Build:   Build,
	}
}

func (g gnames) Verify(
	ctx context.Context,
	input vlib.Input,
) (vlib.Output, error) {
	namesNum := len(input.NameStrings)
	if namesNum > 0 {
		log.Info().Str("action", "verification").
			Int("namesNum", len(input.NameStrings)).
			Str("example", input.NameStrings[0])
	}
	namesRes := make([]vlib.Name, len(input.NameStrings))

	var matches []mlib.Match

	if input.WithCapitalization {
		names := make([]string, len(input.NameStrings))
		for i := range input.NameStrings {
			names[i] = str.CapitalizeName(input.NameStrings[i])
		}
		matches = g.matcher.MatchNames(names)
	} else {
		matches = g.matcher.MatchNames(input.NameStrings)
	}

	var errString string
	matchRecords, err := g.vf.MatchRecords(ctx, matches, input)
	if err != nil {
		errString = err.Error()
	}

	for i, v := range matches {
		if mr, ok := matchRecords[v.ID]; ok {
			namesRes[i] = outputName(mr, input.WithAllMatches)
			namesRes[i].Error = errString
			if input.WithCapitalization {
				namesRes[i].Name = input.NameStrings[i]
				namesRes[i].ID = gnuuid.New(namesRes[i].Name).String()
			}
		} else {
			log.Warn().Msgf("Cannot find record for '%s'.", v.Name)
		}
	}
	res := vlib.Output{Meta: meta(input, namesRes), Names: namesRes}
	return res, nil
}

func overloadTxt(mr *verifier.MatchRecord) string {
	if !mr.Overload {
		return ""
	}
	return "Too many records (possibly strains), some results are truncated"
}

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

func outputName(mr *verifier.MatchRecord, allMatches bool) vlib.Name {
	s := score.New()
	s.SortResults(mr)
	item := vlib.Name{
		ID:               mr.ID,
		Name:             mr.Name,
		DataSourcesNum:   mr.DataSourcesNum,
		Cardinality:      mr.Cardinality,
		OverloadDetected: overloadTxt(mr),
	}
	bestResult := s.BestResult(mr)
	if bestResult != nil {
		item.Curation = bestResult.Curation
		item.MatchType = bestResult.MatchType
	} else {
		item.OverloadDetected = ""
	}

	if allMatches {
		item.Results = s.Results(mr)
	} else {
		item.BestResult = bestResult
	}
	return item
}

func meta(input vlib.Input, names []vlib.Name) vlib.Meta {
	allSources := len(input.DataSources) == 1 && input.DataSources[0] == 0
	hs := make([]gnctx.Hierarch, len(names))
	for i := range names {
		hs[i] = names[i]
	}
	var c gnctx.Context
	var ks []vlib.Kingdom

	if input.WithContext {
		c = gnctx.New(hs, input.ContextThreshold)
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
		NamesNumber:        len(input.NameStrings),
		WithAllSources:     allSources,
		WithAllMatches:     input.WithAllMatches,
		WithContext:        input.WithContext,
		WithCapitalization: input.WithCapitalization,
		ContextThreshold:   input.ContextThreshold,
		DataSources:        input.DataSources,
		Context:            c.Context.Name,
		ContextPercentage:  c.ContextPercentage,
		Kingdom:            c.Kingdom.Name,
		KingdomPercentage:  c.KingdomPercentage,
		Kingdoms:           ks,
	}
	return res
}

func (g gnames) DataSources(ids ...int) ([]*vlib.DataSource, error) {
	return g.vf.DataSources(ids...)
}

func FirstUpperCase(name string) string {
	runes := []rune(name)
	if len(runes) < 2 {
		return name
	}

	one := runes[0]
	two := runes[1]
	if unicode.IsUpper(one) || !unicode.IsLetter(one) {
		return name
	}
	if one == 'x' && (two == ' ' || unicode.IsUpper(two)) {
		return name
	}
	runes[0] = unicode.ToUpper(one)
	return string(runes)
}

func sortNames(mrs map[string]*verifier.MatchRecord) []string {
	res := make([]string, len(mrs))
	var count int
	for k := range mrs {
		res[count] = k
		count++
	}
	sort.Strings(res)
	return res
}

func (g gnames) GetConfig() config.Config {
	return g.cfg
}
