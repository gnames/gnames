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
	log "github.com/sirupsen/logrus"
)

type gnames struct {
	cfg     config.Config
	vf      verifier.Verifier
	facet   facet.Facet
	matcher gnmatcher.GNmatcher
}

// NewGNames is a constructor that returns implmentation of GNames interface.
func NewGNames(
	cnf config.Config,
	vf verifier.Verifier,
	fc facet.Facet,
) GNames {
	return gnames{
		cfg:     cnf,
		vf:      vf,
		facet:   fc,
		matcher: matcher.NewGNmatcher(cnf.MatcherURL),
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
	log.Printf("Verifying %d name-strings.", len(input.NameStrings))
	namesRes := make([]*vlib.Name, len(input.NameStrings))

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
	matchRecords, err := g.vf.MatchRecords(ctx, matches)
	if err != nil {
		errString = err.Error()
	}

	for i, v := range matches {
		if mr, ok := matchRecords[v.ID]; ok {
			s := score.NewScore()
			s.SortResults(mr)
			item := vlib.Name{
				ID:             mr.InputID,
				Name:           mr.Input,
				MatchType:      mr.MatchType,
				Curation:       mr.Curation,
				DataSourcesNum: mr.DataSourcesNum,
				BestResult:     s.BestResult(mr),
				Results:        s.Results(input.DataSources, mr, input.WithAllMatches),
				Error:          errString,
			}
			if input.WithCapitalization {
				item.Name = input.NameStrings[i]
				item.ID = gnuuid.New(item.Name).String()
			}

			namesRes[i] = &item
		} else {
			log.Warnf("Cannot find record for '%s'.", v.Name)
		}
	}
	res := vlib.Output{Meta: meta(input, namesRes), Names: namesRes}
	return res, nil
}

func (g gnames) Search(
	ctx context.Context,
	inp search.Input,
) search.Output {
	inp.Query = inp.ToQuery()
	log.Printf("Searching '%s'.", inp.Query)

	res := search.Output{Meta: search.Meta{Input: inp}}
	matchRecords, err := g.facet.Search(ctx, inp)
	if err != nil {
		res.Error = err.Error()
	}
	res.NamesNumber = len(matchRecords)

	sortedCanonicals := sortCanonicals(matchRecords)
	resCans := make([]*vlib.Name, len(matchRecords))
	var dss []int
	var all bool
	if inp.WithAllResults {
		dss = []int{0}
		all = true
	}

	for i, v := range sortedCanonicals {
		mr := matchRecords[v]
		s := score.NewScore()
		s.SortResults(mr)
		item := vlib.Name{
			ID:         mr.InputID,
			Name:       mr.Input,
			MatchType:  mr.MatchType,
			BestResult: s.BestResult(mr),
			Results:    s.Results(dss, mr, all),
		}
		item.Curation = item.BestResult.Curation
		resCans[i] = &item
	}

	res.Names = resCans
	return res
}

func meta(input vlib.Input, names []*vlib.Name) vlib.Meta {
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

func sortCanonicals(mrs map[string]*verifier.MatchRecord) []string {
	res := make([]string, len(mrs))
	var count int
	for k := range mrs {
		res[count] = k
		count++
	}
	sort.Strings(res)
	return res
}
