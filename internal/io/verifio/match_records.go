package verifio

import (
	"github.com/gnames/gnames/pkg/ent/verif"
	mlib "github.com/gnames/gnlib/ent/matcher"
	vlib "github.com/gnames/gnlib/ent/verifier"
)

// partitionMatches partitions matches into two categories:
// no match, match by canonical.
func partitionMatches(matches []mlib.Match) verif.MatchSplit {
	ms := verif.MatchSplit{
		NoMatch:   make([]*mlib.Match, 0, len(matches)),
		Virus:     make([]*mlib.Match, 0, len(matches)),
		Canonical: make([]*mlib.Match, 0, len(matches)),
	}
	for i := range matches {
		switch matches[i].MatchType {
		case vlib.NoMatch:
			ms.NoMatch = append(ms.NoMatch, &matches[i])
		case vlib.Virus:
			ms.Virus = append(ms.Virus, &matches[i])
		default:
			ms.Canonical = append(ms.Canonical, &matches[i])
		}
	}
	return ms
}
