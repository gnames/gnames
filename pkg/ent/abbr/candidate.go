package abbr

import (
	"strings"

	alib "github.com/gnames/gnlib/ent/abbr"
)

// Candidate is a genus that matched at least one unit and is within the
// tolerance.
type Candidate struct {
	// Genus is the name of the genus.
	Genus string

	// Matches are matches of units, in the order of Query.Units.
	Matches []Match

	// Unmatched are units not matched by the genus.
	Unmatched []*Unit

	// Records are rows of `abbr_genera` for the genus. A homonym has
	// several rows.
	Records []GenusRecord

	// Record is the row whose classification is shown in the output.
	Record GenusRecord

	// Score is the score of the genus and its parts.
	Score alib.Score
}

// Match is the best match of a unit in a genus.
type Match struct {
	// Unit is the matched unit.
	Unit *Unit

	// Lookup is the matched row of `abbr_lookups`.
	Lookup Lookup

	// MatchType tells how the unit matched.
	MatchType alib.MatchTypeValue

	// Key is the key of the epithet the match is based on.
	Key string
}

// Candidates matches units of the query to lookup rows, groups matches by
// genus, and returns genera within the tolerance. The second value maps
// keys to the number of genera that have them, among genera that start
// with the prefix.
func (q Query) Candidates(
	lookups []Lookup,
) ([]*Candidate, map[string]int) {
	keyGenera := make(map[string]int)
	genusKeys := make(map[string]map[string][]Lookup)
	var genera []string
	for _, l := range lookups {
		if !strings.HasPrefix(l.Genus, q.Prefix) {
			continue
		}
		keys, ok := genusKeys[l.Genus]
		if !ok {
			keys = make(map[string][]Lookup)
			genusKeys[l.Genus] = keys
			genera = append(genera, l.Genus)
		}
		if _, ok := keys[l.Key]; !ok {
			keyGenera[l.Key]++
		}
		keys[l.Key] = append(keys[l.Key], l)
	}

	var res []*Candidate
	for _, g := range genera {
		c := q.candidate(g, genusKeys[g], keyGenera)
		if c != nil {
			res = append(res, c)
		}
	}
	return res, keyGenera
}

func (q Query) candidate(
	genus string,
	keys map[string][]Lookup,
	keyGenera map[string]int,
) *Candidate {
	res := &Candidate{Genus: genus}
	for _, u := range q.Units {
		m, ok := bestMatch(u, keys)
		if !ok {
			res.Unmatched = append(res.Unmatched, u)
			continue
		}
		m.Key = matchKey(u, m.MatchType, keyGenera)
		res.Matches = append(res.Matches, m)
	}

	matched := len(res.Matches)
	if matched == 0 {
		return nil
	}

	var allowed int
	if matched >= MinMatchedForTolerance {
		allowed = len(q.Units) * q.Tolerance / 100
	}
	if len(res.Unmatched) > allowed {
		return nil
	}
	return res
}

// bestMatch finds the best match of a unit among lookup rows of a genus.
// A better match type wins, then a name from more data sources, then the
// alphabetically first name.
func bestMatch(u *Unit, keys map[string][]Lookup) (Match, bool) {
	var res Match
	var found bool
	seen := make(map[string]struct{})
	for _, k := range u.Keys {
		for _, l := range keys[k] {
			if _, ok := seen[l.CanonicalStemID]; ok {
				continue
			}
			seen[l.CanonicalStemID] = struct{}{}

			mt := MatchType(u.Stems, Epithets(l.CanonicalStem))
			if mt == alib.NoMatch {
				continue
			}
			m := Match{Unit: u, Lookup: l, MatchType: mt}
			if !found || better(m, res) {
				res = m
				found = true
			}
		}
	}
	return res, found
}

func better(a, b Match) bool {
	if a.MatchType != b.MatchType {
		return a.MatchType > b.MatchType
	}
	if a.Lookup.DsNum != b.Lookup.DsNum {
		return a.Lookup.DsNum > b.Lookup.DsNum
	}
	return a.Lookup.CanonicalStem < b.Lookup.CanonicalStem
}
