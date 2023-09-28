package score

import (
	"math"
	"slices"
	"strings"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

// authMatch covers all possible states of authors matching.
type authMatch int

var clt = collate.New(language.English, collate.Loose)

const (
	// noOverlap: authors do not overlap.
	noOverlap authMatch = iota
	// incomparable: cannot compare because author sets are empty.
	incomparable
	// noAuthVsAuth means that authors cannot be compared, and input has
	// no authorship, but matched name does provide authorship.
	noAuthVsAuth
	// overlap: there are some common authors, but both lists have unique authors.
	overlap
	// fullInclusion: one set of authors fully included into another set.
	fullInclusion
	// identical: both sets include all authors.
	identical
)

var authNorm = map[string]string{
	"Linné":   "Linn",
	"Linne":   "Linn",
	"Sokolov": "Sokoloff",
}

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
	auth2 = authorsNormalize(auth2)
	if len(auth1) == 0 && len(auth2) == 0 {
		return incomparable
	}
	if len(auth2) == 0 {
		return incomparable
	}
	if len(auth1) == 0 && len(auth2) > 0 {
		return noAuthVsAuth
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
	rsShort := []rune(short)
	rs := []rune(long)[0:len(rsShort)]
	short2 := string(rs)

	return clt.CompareString(short, short2) == 0, au1 < au2
}

// authorsNormalize normalizes a list of authors.
func authorsNormalize(auths []string) []string {
	res := make([]string, 0, len(auths))
	for _, v := range auths {
		auth := authNormalize(v)
		res = append(res, auth)
	}
	slices.Sort(res)
	return res
}

// authNormalize normalizes one author.
func authNormalize(auth string) string {
	words := strings.Split(auth, " ")
	l := len(words) - 1
	if l > 0 {
		words = authTrimAnnot(words)
		l = len(words) - 1
	}

	auth = words[l]
	if strings.HasPrefix(auth, "d'") {
		auth = auth[2:]
	}
	auth = strings.TrimRight(auth, ".")
	if v, ok := authNorm[auth]; ok {
		auth = v
	}
	return auth
}

func authTrimAnnot(auth []string) []string {
	el := []string{"fil.", "I", "II", "III"}
	l := len(auth)
	for i := range el {
		if auth[l-1] == el[i] {
			return auth[0 : l-1]
		}
	}
	return auth
}

// findYearsMatch determines how two years values relate to each other.
func findYearsMatch(y1, y2 int) yearMatch {
	if y1 == 0 || y2 == 0 {
		return notAvailable
	}
	diff := math.Abs(float64(y1) - float64(y2))
	if diff == 0 {
		return perfectMatch
	}
	if diff <= 2 {
		return approxMatch
	}
	return noMatch
}
