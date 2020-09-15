package score

import (
	"sort"
	"strings"
)

// authMatch covers all possible states of authors matching.
type authMatch int

const (
	// noOverlap: authors do not overlap.
	noOverlap authMatch = iota
	// uncomparable: cannot compare because one or both author sets are empty.
	uncomparable
	// overlap: there are some common authors, but both lists have unique authors.
	overlap
	// fullInclusion: one set of authors fully included into another set.
	fullInclusion
	// identical: both sets include all authors.
	identical
)

// yearMatch covers possible states of matching years.
type yearMatch int

const (
	// noMatch: years are too different.
	noMatch yearMatch = iota
	// notAvailable: one or both years are not given.
	notAvailable
	// approxMatch: years differ slightly.
	approxMatch
	// perfectMatch: years match exactly.
	perfectMatch
)

// auth takes two lists of authors, and their corresponding years and
// tries to match them to each other. The score is decided on how well
// authors and years did match.
// The score takes 3 bits and ranges from 0 to 7.
func (s Score) auth(auth1, auth2 []string, year1, year2 int) Score {
	shift := 25
	years := findYearsMatch(year1, year2)
	authors := findAuthMatch(auth1, auth2)
	var i uint32 = 0

	if authors == identical {
		switch years {
		case perfectMatch:
			i = 0b111 //7
		case approxMatch:
			i = 0b110 //6
		case notAvailable:
			i = 0b101 //5
		case noMatch:
			i = 0b011 //3
		}
	} else if authors == fullInclusion {
		switch years {
		case perfectMatch:
			i = 0b110 //6
		case approxMatch:
			i = 0b101 //5
		case notAvailable:
			i = 0b100 //4
		case noMatch:
			i = 0b010 //2
		}
	} else if authors == overlap {
		switch years {
		case noMatch:
			i = 0b000 //0
		default:
			i = 0b001 //1
		}
	} else if authors == uncomparable {
		i = 0b100 //4
	}
	s.Value = (s.Value | i<<shift)
	return s
}

func max(i1, i2 uint32) uint32 {
	if i1 >= i2 {
		return i1
	}
	return i2
}

// findAuthMatch determines how much two slices of author strings relate
// to each other.
func findAuthMatch(auth1, auth2 []string) authMatch {
	auth1 = authorsNormalize(auth1)
	if len(auth1) == 0 {
		return uncomparable
	}
	auth2 = authorsNormalize(auth2)
	if len(auth2) == 0 {
		return uncomparable
	}
	return findAuthOverlap(auth1, auth2)
}

// findAuthOverlap determines how much two slices of author strings overlap
// each other.
func findAuthOverlap(auth1, auth2 []string) authMatch {
	var nomatchLong, nomatchShort []string
	short, long := auth1, auth2
	if len(auth2) < len(auth1) {
		short, long = auth2, auth1
	}

	cursor := 0
	giveup := false
	match := false
OUTER:
	for _, v := range long {
		for i, vv := range short {
			if i < cursor {
				continue
			}
			match, giveup = compareAuth(v, vv)
			if match {
				nomatchShort = append(nomatchShort, short[cursor:i]...)
				cursor = i + 1
				continue OUTER
			}
			if giveup {
				if !match {
					nomatchLong = append(nomatchLong, v)
				}
				giveup = false
				continue OUTER
			}
		}
		nomatchLong = append(nomatchLong, v)
	}

	nomatchShort = append(nomatchShort, short[cursor:]...)

	if len(nomatchLong)+len(nomatchShort) == 0 {
		return identical
	}

	if len(nomatchLong) == 0 || len(nomatchShort) == 0 {
		return fullInclusion
	}

	if len(nomatchLong) < len(long) && len(nomatchShort) < len(short) {
		return overlap
	}

	return noOverlap
}

// compareAuth matches two authors to one another.
func compareAuth(au1, au2 string) (bool, bool) {
	short, long := au1, au2
	if len(au1) > len(au2) {
		short, long = au2, au1
	}
	return strings.HasPrefix(long, short), au1 < au2
}

// authorsNormalize normalizes a list of authors.
func authorsNormalize(auths []string) []string {
	res := make([]string, 0, len(auths))
	for _, v := range auths {
		auth := authNormalize(v)
		if auth == "" {
			continue
		}
		res = append(res, auth)
	}
	sort.Strings(res)
	return res
}

// authNormalize normalizes one author.
func authNormalize(auth string) string {
	words := strings.Split(auth, " ")
	if len(words) == 1 {
		return strings.TrimRight(words[0], ".")
	}

	res := make([]string, 0, len(words))
	for _, v := range words {
		if v[len(v)-1] == '.' && len(v) == 2 {
			continue
		}
		res = append(res, strings.TrimRight(v, "."))
	}
	return strings.Join(res, " ")
}

// findYearsMatch determines how two years values relate to each other.
func findYearsMatch(y1, y2 int) yearMatch {
	if y1 == 0 || y2 == 0 {
		return notAvailable
	}
	diff := y1 - y2
	if diff == 0 {
		return perfectMatch
	}
	if diff == -1 || diff == 1 {
		return approxMatch
	}
	return noMatch
}
