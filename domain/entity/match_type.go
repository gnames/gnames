package entity

// MatchType describes how a name-string matched a name in gnames database.
type MatchType int

const (
	// NoMatch means that verification failed.
	NoMatch MatchType = iota

	// Exact means either canonical form, or the whole name-string matched
	// perfectlly.
	Exact

	// Fuzzy means that matches were not exact due to similarity of name-strings,
	// OCR or typing errors. Take these results with more suspition than
	// Exact matches. Fuzzy match is never done on uninomials due to the
	// high rate of false positives.
	Fuzzy

	// PartialExact: GNames failed to match full name string. Now the match
	// happened by removing either middle species epithets, or by choppping the
	// 'tail' words of the input name-string canonical form.
	PartialExact

	// PartialFuzzy is the same as PartialExact, but also the match was not
	// exact. We never do fuzzy matches for uninomials, due to high rate of false
	// positives.
	PartialFuzzy
)

var mapMatchType = map[int]string{
	0: "NoMatch",
	1: "Exact",
	2: "Fuzzy",
	3: "PartialExact",
	4: "PartialFuzzy",
}

func (mt MatchType) String() string {
	if match, ok := mapMatchType[int(mt)]; ok {
		return match
	} else {
		return "N/A"
	}
}
