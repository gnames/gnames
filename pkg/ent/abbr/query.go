package abbr

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	alib "github.com/gnames/gnlib/ent/abbr"
	"github.com/gnames/gnparser/ent/stemmer"
)

const (
	// DefaultTolerance is the percentage of names that may be unmatched
	// for a genus, if the input does not set it.
	DefaultTolerance = 20

	// DefaultGeneraLimit is the number of candidate genera in the result,
	// if the input does not set it.
	DefaultGeneraLimit = 10

	// MaxGeneraLimit is the largest allowed GeneraLimit.
	MaxGeneraLimit = 100

	// MaxNames is the largest number of names in a request.
	MaxNames = 5_000

	// MaxGenusHints is the number of genus hints of a name that are used.
	// Other hints are ignored.
	MaxGenusHints = 5

	// MinMatchedForTolerance is the number of matched names a genus needs
	// before its unmatched names are tolerated.
	MinMatchedForTolerance = 2
)

// abbrRe matches a genus abbreviation: a capital letter, up to two more
// letters, and a period.
var abbrRe = regexp.MustCompile(`^\p{Lu}\p{Ll}{0,2}\.$`)

// epithetRe matches an epithet: lowercase letters, maybe with hyphens.
var epithetRe = regexp.MustCompile(`^\p{Ll}[\p{Ll}-]*$`)

// Query is a parsed and validated input.
type Query struct {
	// Input is the original input.
	Input alib.Input

	// Abbreviation is the genus abbreviation of all names, for example
	// `Cae.`.
	Abbreviation string

	// Prefix is the abbreviation without the period. Candidate genera
	// start with it.
	Prefix string

	// Tolerance is the percentage of names that may be unmatched for a
	// genus.
	Tolerance int

	// GeneraLimit is the maximum number of candidate genera.
	GeneraLimit int

	// Units are distinct names of the input, by their epithet stems.
	Units []*Unit
}

// Unit is a group of input names with the same epithet stems. The same
// name often occurs many times in a text. It counts once for matching
// and scoring.
type Unit struct {
	// Names are the input names of the unit, in input order.
	Names []alib.Name

	// Epithets are normalized epithets after removal of nominotypical
	// repeats.
	Epithets []string

	// Stems are stems of Epithets.
	Stems []string

	// Keys are lookup keys of Stems.
	Keys []string

	// Hints maps hinted genera to their best (lowest) position, starting
	// from 1.
	Hints map[string]int

	// HintFactors maps hinted genera to the multiplier of their hint
	// weight. A binomial or trinomial hint is stronger than a genus.
	HintFactors map[string]float64
}

// NewQuery validates the input, parses its names and groups them into
// units. All names must have the same capitalized genus abbreviation
// that ends with a period, otherwise it returns an error.
func NewQuery(inp alib.Input) (Query, error) {
	res := Query{
		Input:       inp,
		Tolerance:   DefaultTolerance,
		GeneraLimit: DefaultGeneraLimit,
	}

	if inp.Tolerance != nil {
		res.Tolerance = min(max(*inp.Tolerance, 0), 100)
	}
	if inp.GeneraLimit > 0 {
		res.GeneraLimit = min(inp.GeneraLimit, MaxGeneraLimit)
	}

	if len(inp.Names) == 0 {
		return res, errors.New("no names in the request")
	}
	if len(inp.Names) > MaxNames {
		return res, fmt.Errorf(
			"too many names: %d, the limit is %d",
			len(inp.Names), MaxNames,
		)
	}

	for _, v := range inp.Names {
		abbr, err := abbreviation(v.Name)
		if err != nil {
			return res, err
		}
		if res.Abbreviation == "" {
			res.Abbreviation = abbr
			continue
		}
		if abbr != res.Abbreviation {
			return res, fmt.Errorf(
				"names have different genus abbreviations: '%s' and '%s'",
				res.Abbreviation, abbr,
			)
		}
	}
	res.Prefix = strings.TrimSuffix(res.Abbreviation, ".")

	units := make(map[string]*Unit)
	for _, v := range inp.Names {
		epithets := splitEpithets(v.Name)
		u := newUnit(res.Abbreviation, epithets)
		id := strings.Join(u.Stems, " ")
		if _, ok := units[id]; !ok {
			units[id] = u
			res.Units = append(res.Units, u)
		}
		u = units[id]
		u.Names = append(u.Names, v)
		u.addHints(res.Prefix, v.GenusHints)
	}

	return res, nil
}

// Keys returns distinct lookup keys of all units.
func (q Query) Keys() []string {
	var res []string
	for _, u := range q.Units {
		res = append(res, u.Keys...)
	}
	slices.Sort(res)
	return slices.Compact(res)
}

// abbreviation returns the genus abbreviation of a name, or an error if
// the name does not start with one.
func abbreviation(name string) (string, error) {
	words := strings.Fields(name)
	if len(words) == 0 {
		return "", errors.New("empty name in the request")
	}
	if !abbrRe.MatchString(words[0]) {
		return "", fmt.Errorf(
			"name '%s' does not start with a genus abbreviation "+
				"like 'C.' or 'Cae.'",
			name,
		)
	}
	return words[0], nil
}

// splitEpithets returns epithets of a name. Names are not parsed: they
// come from GNfinder, which already normalizes them. Words with a period
// (rank markers, abbreviated epithets like `P. t. altaica`) and hybrid
// signs are skipped. Epithets end at the first other word, such as an
// author.
func splitEpithets(name string) []string {
	var res []string
	for _, w := range strings.Fields(name)[1:] {
		if w == "×" || strings.HasSuffix(w, ".") {
			continue
		}
		if !epithetRe.MatchString(w) {
			break
		}
		res = append(res, w)
	}
	return res
}

// newUnit creates a unit from epithets of a name. Nominotypical repeats
// are removed: `B. bubo bubo` becomes `B. bubo`.
func newUnit(abbr string, epithets []string) *Unit {
	res := &Unit{
		Hints:       make(map[string]int),
		HintFactors: make(map[string]float64),
	}
	for _, e := range epithets {
		key := stemmer.AbbrKey(abbr, e)
		if key == "" {
			continue
		}
		res.Epithets = append(res.Epithets, e)
		res.Keys = append(res.Keys, key)
		res.Stems = append(res.Stems, key[strings.Index(key, "-")+1:])
	}

	for l := len(res.Stems); l > 1 && res.Stems[l-1] == res.Stems[l-2]; {
		l--
		res.Epithets = res.Epithets[:l]
		res.Keys = res.Keys[:l]
		res.Stems = res.Stems[:l]
	}
	return res
}

// addHints adds genus hints of a name to the unit. A hint can be a
// genus or a binomial or trinomial name, which hints at its genus.
// Hints that do not start with the prefix are ignored. Positions count
// distinct genera, so several names of one genus take one position.
// The factor of a genus comes from its strongest hint.
func (u *Unit) addHints(prefix string, hints []string) {
	var pos int
	seen := make(map[string]struct{})
	for _, h := range hints {
		words := strings.Fields(h)
		if len(words) == 0 || !strings.HasPrefix(words[0], prefix) {
			continue
		}
		genus := words[0]
		if _, ok := seen[genus]; !ok {
			if pos == MaxGenusHints {
				continue
			}
			seen[genus] = struct{}{}
			pos++
			if p, ok := u.Hints[genus]; !ok || pos < p {
				u.Hints[genus] = pos
			}
		}
		u.HintFactors[genus] = max(u.HintFactors[genus], u.hintFactor(h))
	}
}

// hintFactor returns the multiplier of the weight of a hint. A name
// with the same last epithet as the unit is the strongest hint, then
// any other name, then a genus.
func (u *Unit) hintFactor(hint string) float64 {
	hu := newUnit(hint, splitEpithets(hint))
	l := len(hu.Stems)
	switch {
	case l == 0:
		return 1
	case len(u.Stems) > 0 && hu.Stems[l-1] == u.Stems[len(u.Stems)-1]:
		return SameNameHintFactor
	default:
		return NameHintFactor
	}
}
