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
	params vlib.Input,
) (vlib.Output, error) {
	log.Printf("Verifying %d name-strings.", len(params.NameStrings))
	namesRes := make([]*vlib.Name, len(params.NameStrings))

	var matches []mlib.Match

	if params.WithCapitalization {
		names := make([]string, len(params.NameStrings))
		for i := range params.NameStrings {
			names[i] = str.CapitalizeName(params.NameStrings[i])
		}
		matches = g.matcher.MatchNames(names)
	} else {
		matches = g.matcher.MatchNames(params.NameStrings)
	}

	var errString string
	matchRecords, err := g.vf.MatchRecords(ctx, matches)
	if err != nil {
		errString = err.Error()
	}

	log.Printf("REC: %#v", matchRecords)

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
				Results:        s.Results(params.DataSources, mr, params.WithAllMatches),
				Error:          errString,
			}
			if params.WithCapitalization {
				item.Name = params.NameStrings[i]
				item.ID = gnuuid.New(item.Name).String()
			}

			namesRes[i] = &item
		} else {
			log.Warnf("Cannot find record for '%s'.", v.Name)
		}
	}
	log.Printf("%#v", params)
	res := vlib.Output{Meta: meta(params, namesRes), Names: namesRes}
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
		resCans[i] = &item
	}

	res.Names = resCans
	return res
}

func meta(params vlib.Input, names []*vlib.Name) vlib.Meta {
	allSources := len(params.DataSources) == 1 && params.DataSources[0] == 0
	hs := make([]gnctx.Hierarch, len(names))
	for i := range names {
		hs[i] = names[i]
	}
	c := gnctx.Context{Context: &gnctx.Clade{}, Kingdom: &gnctx.Clade{}}
	if params.WithContext {
		c = gnctx.New(hs, params.ContextThreshold)
	}
	res := vlib.Meta{
		NamesNumber:        len(params.NameStrings),
		WithAllSources:     allSources,
		WithAllMatches:     params.WithAllMatches,
		WithContext:        params.WithContext,
		WithCapitalization: params.WithCapitalization,
		ContextThreshold:   params.ContextThreshold,
		DataSources:        params.DataSources,
		Context:            c.Context.Name,
		ContextPercentage:  c.ContextPercentage,
		Kingdom:            c.Kingdom.Name,
		KingdomPercentage:  c.KingdomPercentage,
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
