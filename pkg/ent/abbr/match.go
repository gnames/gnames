package abbr

import (
	"slices"
	"strings"

	alib "github.com/gnames/gnlib/ent/abbr"
)

// MatchType compares epithet stems of an input name with epithet stems
// of a database name. The last epithet identifies a name:
//
//   - Full: the same epithets.
//   - EpithetMoved: the same last epithet, other epithets differ.
//   - SpeciesLevel: the same species epithet, last epithets differ.
//
// Any other shared epithet is NoMatch. Both lists must be free of
// nominotypical repeats.
func MatchType(input, db []string) alib.MatchTypeValue {
	if len(input) == 0 || len(db) == 0 {
		return alib.NoMatch
	}
	switch {
	case slices.Equal(input, db):
		return alib.Full
	case input[len(input)-1] == db[len(db)-1]:
		return alib.EpithetMoved
	case input[0] == db[0]:
		return alib.SpeciesLevel
	}
	return alib.NoMatch
}

// Epithets returns epithets of a canonical form (stemmed or not) without
// nominotypical repeats: `Bubo bubo bubo` gives `bubo`.
func Epithets(canonical string) []string {
	words := strings.Fields(canonical)
	if len(words) < 2 {
		return nil
	}
	res := words[1:]
	for l := len(res); l > 1 && res[l-1] == res[l-2]; l-- {
		res = res[:l-1]
	}
	return res
}

// matchKey returns the key of the epithet that a match is based on. For
// a full match it is the rarest key of the name. For other matches it is
// the key of the shared epithet.
func matchKey(
	u *Unit,
	mt alib.MatchTypeValue,
	keyGenera map[string]int,
) string {
	switch mt {
	case alib.EpithetMoved:
		return u.Keys[len(u.Keys)-1]
	case alib.SpeciesLevel:
		return u.Keys[0]
	}
	res := u.Keys[0]
	for _, k := range u.Keys[1:] {
		if keyGenera[k] < keyGenera[res] {
			res = k
		}
	}
	return res
}

// ExactEpithets tells if epithets of a matched canonical form are the
// same as epithets of the input name, not only their stems. Only
// epithets that the match type compares are checked.
func ExactEpithets(
	input, canonical []string,
	mt alib.MatchTypeValue,
) bool {
	if len(input) == 0 || len(canonical) == 0 {
		return false
	}
	switch mt {
	case alib.Full:
		return slices.Equal(input, canonical)
	case alib.EpithetMoved:
		return input[len(input)-1] == canonical[len(canonical)-1]
	case alib.SpeciesLevel:
		return input[0] == canonical[0]
	}
	return false
}
