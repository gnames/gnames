package gnames

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/gnames/gnames/pkg/ent/abbr"
	alib "github.com/gnames/gnlib/ent/abbr"
)

// AbbrNames finds candidate genera for names with the same abbreviated
// genus. Errors in the input are returned in the metadata of the output.
// Errors of the database are returned as errors.
func (g gnames) AbbrNames(
	ctx context.Context,
	inp alib.Input,
) (alib.Output, error) {
	q, err := abbr.NewQuery(inp)
	if err == nil {
		inp, err = g.abbrGenusHints(ctx, inp)
		if err != nil {
			return alib.Output{Meta: abbrMeta(q)},
				fmt.Errorf("gnames.AbbrNames: %w", err)
		}
		q, err = abbr.NewQuery(inp)
	}
	res := alib.Output{Meta: abbrMeta(q), Genera: []alib.Genus{}}
	if err != nil {
		res.Error = err.Error()
		return res, nil
	}

	cands, prefixSize, err := g.abbrCandidates(ctx, q)
	if err != nil {
		return res, fmt.Errorf("gnames.AbbrNames: %w", err)
	}
	if q.Input.WithFeatures {
		res.PrefixSize = prefixSize
	}

	canonicals, err := g.abbrCanonicals(ctx, cands)
	if err != nil {
		return res, fmt.Errorf("gnames.AbbrNames: %w", err)
	}

	matched := make(map[*abbr.Unit]struct{})
	for _, c := range cands {
		res.Genera = append(res.Genera, abbrGenus(c, canonicals))
		for _, m := range c.Matches {
			matched[m.Unit] = struct{}{}
		}
	}

	for _, u := range q.Units {
		if _, ok := matched[u]; ok {
			res.MatchedNamesNumber += len(u.Names)
		}
	}
	res.UnmatchedNamesNumber = res.NamesNumber - res.MatchedNamesNumber
	return res, nil
}

// abbrGenusHints removes one-word genus hints that are not genera in
// the database. GNfinder sends any capitalized word near a name as a
// hint, and such words would take positions of real genera. Binomial
// and trinomial hints are kept without a check.
func (g gnames) abbrGenusHints(
	ctx context.Context,
	inp alib.Input,
) (alib.Input, error) {
	var hints []string
	for _, n := range inp.Names {
		for _, h := range n.GenusHints {
			if h = strings.TrimSpace(h); !isGenus(h) {
				continue
			}
			hints = append(hints, h)
		}
	}
	if len(hints) == 0 {
		return inp, nil
	}
	slices.Sort(hints)
	recs, err := g.ab.AbbrGenera(ctx, slices.Compact(hints))
	if err != nil {
		return inp, err
	}
	genera := make(map[string]struct{}, len(recs))
	for _, r := range recs {
		genera[r.Genus] = struct{}{}
	}

	names := make([]alib.Name, len(inp.Names))
	for i, n := range inp.Names {
		var hs []string
		for _, h := range n.GenusHints {
			h = strings.TrimSpace(h)
			if _, ok := genera[h]; ok || !isGenus(h) {
				hs = append(hs, h)
			}
		}
		n.GenusHints = hs
		names[i] = n
	}
	inp.Names = names
	return inp, nil
}

// isGenus tells if a hint is one word, not a binomial or trinomial
// name.
func isGenus(hint string) bool {
	return !strings.Contains(hint, " ")
}

// abbrCandidates finds, scores and sorts candidate genera, and keeps
// only the best of them, up to GeneraLimit. It also returns the size of
// all genera with the prefix, which is 0 if nothing matched.
func (g gnames) abbrCandidates(
	ctx context.Context,
	q abbr.Query,
) ([]*abbr.Candidate, int, error) {
	keys := q.Keys()
	if len(keys) == 0 {
		return nil, 0, nil
	}

	lookups, err := g.ab.AbbrLookups(ctx, keys)
	if err != nil {
		return nil, 0, err
	}

	cands, keyGenera := q.Candidates(lookups)
	if len(cands) == 0 {
		return nil, 0, nil
	}

	genera := make([]string, len(cands))
	for i, c := range cands {
		genera[i] = c.Genus
	}
	recs, err := g.ab.AbbrGenera(ctx, genera)
	if err != nil {
		return nil, 0, err
	}
	recsMap := make(map[string][]abbr.GenusRecord)
	for _, r := range recs {
		recsMap[r.Genus] = append(recsMap[r.Genus], r)
	}

	prefixSize, err := g.ab.AbbrPrefixSize(ctx, q.Prefix)
	if err != nil {
		return nil, 0, err
	}

	s := abbr.Scorer{Query: q, KeyGenera: keyGenera, PrefixSize: prefixSize}
	for _, c := range cands {
		c.Records = recsMap[c.Genus]
		s.Score(c)
	}
	abbr.Sort(cands)

	if len(cands) > q.GeneraLimit {
		cands = cands[:q.GeneraLimit]
	}

	if q.Input.WithFeatures {
		for _, c := range cands {
			f := s.Features(c)
			c.Features = &f
		}
	}
	return cands, prefixSize, nil
}

// abbrCanonicals returns canonical forms of matched names of the
// candidates.
func (g gnames) abbrCanonicals(
	ctx context.Context,
	cands []*abbr.Candidate,
) (map[string][]string, error) {
	var ids []string
	for _, c := range cands {
		for _, m := range c.Matches {
			ids = append(ids, m.Lookup.CanonicalStemID)
		}
	}
	if len(ids) == 0 {
		return nil, nil
	}
	slices.Sort(ids)
	return g.ab.AbbrCanonicals(ctx, slices.Compact(ids))
}

func abbrMeta(q abbr.Query) alib.Meta {
	names := q.Input.Names
	if names == nil {
		names = []alib.Name{}
	}
	return alib.Meta{
		Names:         names,
		Abbreviation:  q.Abbreviation,
		MainTaxon:     q.Input.MainTaxon,
		DominantClass: q.Input.DominantClass,
		Tolerance:     q.Tolerance,
		GeneraLimit:   q.GeneraLimit,
		WithFeatures:  q.Input.WithFeatures,
		NamesNumber:   len(names),
	}
}

func abbrGenus(c *abbr.Candidate, canonicals map[string][]string) alib.Genus {
	res := alib.Genus{
		Name:                c.Genus,
		ClassificationPath:  c.Record.Classification,
		ClassificationRanks: c.Record.ClassificationRanks,
		EpithetsNumber:      c.Record.EpithetsNum,
		Score:               c.Score,
		MatchedNames:        []alib.MatchedName{},
		Features:            c.Features,
	}

	for _, m := range c.Matches {
		can, exact := abbrCanonical(m, canonicals[m.Lookup.CanonicalStemID])
		for _, n := range m.Unit.Names {
			res.MatchedNames = append(res.MatchedNames, alib.MatchedName{
				ID:               n.ID,
				Name:             n.Name,
				MatchedCanonical: can,
				MatchType:        m.MatchType,
				StemMatch:        !exact,
			})
		}
	}

	for _, u := range c.Unmatched {
		for _, n := range u.Names {
			res.UnmatchedNames = append(res.UnmatchedNames, alib.Name{
				ID:   n.ID,
				Name: n.Name,
			})
		}
	}
	return res
}

// abbrCanonical picks a canonical form of a match. Several canonical
// forms can share a stem (`Carex alba`, `Carex albus`). The form with
// the same epithets as the input wins, then the alphabetically first.
// The boolean tells if epithets are the same, not only their stems.
func abbrCanonical(m abbr.Match, canonicals []string) (string, bool) {
	if len(canonicals) == 0 {
		return m.Lookup.CanonicalStem, false
	}
	canonicals = slices.Clone(canonicals)
	slices.Sort(canonicals)
	for _, c := range canonicals {
		if abbr.ExactEpithets(m.Unit.Epithets, abbr.Epithets(c), m.MatchType) {
			return c, true
		}
	}
	return canonicals[0], false
}
