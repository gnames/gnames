package gnames

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/gnames/gnames/internal/ent/facet"
	"github.com/gnames/gnames/internal/ent/lexgroup"
	"github.com/gnames/gnames/internal/ent/score"
	"github.com/gnames/gnames/internal/ent/verifier"
	"github.com/gnames/gnames/internal/io/matcher"
	"github.com/gnames/gnames/pkg/config"
	"github.com/gnames/gnlib/ent/gnvers"
	mlib "github.com/gnames/gnlib/ent/matcher"
	"github.com/gnames/gnlib/ent/reconciler"
	vlib "github.com/gnames/gnlib/ent/verifier"
	gnmatcher "github.com/gnames/gnmatcher/pkg"
	gncfg "github.com/gnames/gnmatcher/pkg/config"
	"github.com/gnames/gnparser/ent/str"
	"github.com/gnames/gnquery/ent/search"
	"github.com/gnames/gnstats/ent/stats"
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

	var matchOut mlib.Output
	var opts []gncfg.Option
	if input.WithSpeciesGroup {
		opts = append(opts, gncfg.OptWithSpeciesGroup(true))
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

	var errString string
	matchRecords, err := g.vf.MatchRecords(ctx, matchOut.Matches, input)
	if err != nil {
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
			log.Warn().Msgf("Cannot find record for '%s'.", v.Name)
		}
	}
	res := vlib.Output{Meta: meta(input, namesRes), Names: namesRes}
	return res, nil
}

func (g gnames) Reconcile(
	verif vlib.Output,
	qs map[string]reconciler.Query,
	ids []string,
) reconciler.Output {
	res := reconciler.Output(make(map[string]reconciler.ReconciliationResult))

	for i, v := range verif.Names {
		prs := qs[ids[i]].Properties
		lgs := lexgroup.NameToLexicalGroups(v)
		lgs = filterLexGrpByProperties(lgs, prs)
		var rcs []reconciler.ReconciliationCandidate

		for i, lg := range lgs {
			cur := 0.95
			first := 0.9
			auth := 0.9
			exact := 0.70
			complete := 0.4
			oneCode := 0.80

			if i == 0 {
				first = 1
			}

			if lg.Data[0].MatchType == vlib.Exact {
				exact = 1
				complete = 1
			} else if lg.Data[0].MatchType == vlib.Fuzzy {
				complete = 1
			} else if lg.Data[0].MatchType == vlib.PartialExact {
				exact = 1
				if lg.Data[0].MatchedCardinality == 1 {
					complete = 0.2
				}
			} else if lg.Data[0].MatchType == vlib.PartialFuzzy {
				if lg.Data[0].MatchedCardinality == 1 {
					complete = 0.2
				}
			}

			if lg.Data[0].ScoreDetails.AuthorMatchScore > 0.2 {
				auth = 1
			}
			if lg.Data[0].ScoreDetails.CuratedDataScore > 0 {
				cur = 1
			}
			if len(lg.NomCodes) < 2 {
				oneCode = 1
			}

			score := first * auth * cur * oneCode * exact * complete

			features := []reconciler.Feature{
				{ID: "first_result", Value: first},
				{ID: "exact_match", Value: exact},
				{ID: "complete_canonical_match", Value: complete},
				{ID: "authors_compatible", Value: auth},
				{ID: "has_curation_process", Value: cur},
				{ID: "single_nomenclatural_code", Value: oneCode},
			}

			rc := reconciler.ReconciliationCandidate{
				ID:       lg.ID,
				Score:    score,
				Match:    score == 1,
				Features: features,
				Name:     lg.Name,
			}
			rcs = append(rcs, rc)
		}
		res[ids[i]] = reconciler.ReconciliationResult{
			Result: rcs,
		}
	}
	return res
}

func (g gnames) NameByID(
	params vlib.NameStringInput,
) (vlib.NameStringOutput, error) {
	var res vlib.NameStringOutput
	mr, err := g.vf.NameByID(params)
	if err != nil {
		return res, err
	}
	meta := vlib.NameStringMeta{
		ID:             params.ID,
		DataSources:    params.DataSources,
		WithAllMatches: params.WithAllMatches,
	}
	res.NameStringMeta = meta

	if mr == nil {
		return res, nil
	}

	name := outputName(mr, params.WithAllMatches)
	res.Name = &name
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
		DataSourcesIDs:   mr.DataSourcesIDs,
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

func filterLexGrpByProperties(
	lgs []lexgroup.LexicalGroup,
	prs []reconciler.Property,
) []lexgroup.LexicalGroup {
	var res []lexgroup.LexicalGroup
	if len(prs) == 0 {
		return lgs
	}
	for i := range lgs {
		grp := filterGroup(lgs[i], prs)
		if len(grp.Data) > 0 {
			res = append(res, grp)
		}
	}
	return res
}

func filterGroup(
	lg lexgroup.LexicalGroup,
	prs []reconciler.Property,
) lexgroup.LexicalGroup {
	var res lexgroup.LexicalGroup
	fs := make(map[string]string)
	for i := range prs {
		pid := strings.ToLower(prs[i].PID)
		fs[pid] = prs[i].Value
	}
	if taxon, ok := fs["higher-taxon"]; ok {
		res = filterByTaxon(lg, taxon)
	} else {
		res = lg
	}
	if idStr, ok := fs["data-source-ids"]; ok {
		ids := getDataSourceIDs(idStr)
		if len(ids) > 0 {
			res = filterByDataSource(lg, ids)
		}
	}
	return res
}

func getDataSourceIDs(s string) map[int]struct{} {
	res := make(map[int]struct{})
	elements := strings.Split(s, ",")
	for _, v := range elements {
		v = strings.TrimSpace(v)
		id, err := strconv.Atoi(v)
		if err == nil {
			res[id] = struct{}{}
		}
	}
	return res
}

func filterByTaxon(
	lg lexgroup.LexicalGroup,
	taxon string,
) lexgroup.LexicalGroup {
	var res lexgroup.LexicalGroup
	var ds []*vlib.ResultData
	d := lg.Data
	for i := range d {
		if d[i].ClassificationPath == "" {
			continue
		}
		if strings.Index(d[i].ClassificationPath, taxon) > -1 {
			ds = append(ds, d[i])
		}
	}
	if len(ds) > 0 {
		res = lexgroup.New(ds[0])
		res.Data = ds
	}
	return res
}

func filterByDataSource(
	lg lexgroup.LexicalGroup,
	ids map[int]struct{},
) lexgroup.LexicalGroup {
	var res lexgroup.LexicalGroup
	var ds []*vlib.ResultData

	d := lg.Data
	for i := range d {
		if _, ok := ids[d[i].DataSourceID]; ok {
			ds = append(ds, d[i])
		}
	}

	if len(ds) > 0 {
		res = lexgroup.New(ds[0])
		res.Data = ds
	}
	return res
}
