package abbr

import (
	"cmp"
	"math"
	"slices"
	"strings"

	alib "github.com/gnames/gnlib/ent/abbr"
)

// Score weights. A score is measured in units of evidence: one unit makes
// a chance match ten times less likely. The weights are starting values,
// to be tuned on abbreviated names from BHL.
const (
	// MaxEvidence caps the evidence of one matched name, so a single odd
	// match cannot decide alone.
	MaxEvidence = 6.0

	// HintWeight is the bonus for a confirmed genus hint at position 1.
	// A hint at position n gives HintWeight/n.
	HintWeight = 3.0

	// NameHintFactor multiplies the weight of a binomial or trinomial
	// hint. Such a hint shows that the genus is in the text.
	NameHintFactor = 1.5

	// SameNameHintFactor multiplies the weight of a binomial or
	// trinomial hint with the same last epithet as the name. The text
	// most likely spelled out the name before.
	SameNameHintFactor = 2.0

	// UnmatchedWeight is the penalty for each unmatched name.
	UnmatchedWeight = -1.0

	// ClassWeight is the bonus for a genus in the dominant class, and the
	// penalty for a genus in another class.
	ClassWeight = 0.5

	// MainTaxonWeight is the bonus for a genus in the main taxon, and the
	// penalty for a genus outside of it.
	MainTaxonWeight = 0.5

	// UnknownGenusWeight is the penalty for a genus that is not in IRMNG
	// and has no classification. It is set high on purpose: such genera
	// are usually misspellings, and with one name they get the most
	// evidence. A real unknown genus needs several names or a hint.
	UnknownGenusWeight = -5.0

	// GenusPriorWeight is the share of the genus size that the anchor
	// match does not pay. A genus is mentioned in proportion to its size,
	// so with the full prior (1) the size cancels for one matched name,
	// and only further names pay it. With 0 every name pays the size.
	GenusPriorWeight = 1.0

	// PopularityWeight is the weight of the log10 of the number of data
	// sources with a matched name. The mean over matched names is used,
	// so popularity breaks ties without rewarding more matches again.
	PopularityWeight = 0.5
)

// matchWeights reduce the evidence of matches that are not the name from
// the text.
var matchWeights = map[alib.MatchTypeValue]float64{
	alib.Full:         1.0,
	alib.EpithetMoved: 0.8,
	alib.SpeciesLevel: 0.6,
}

// Evidence is the evidence that a matched epithet did not appear in the
// genus by chance. The chance is estimated as the size of the genus
// times the number of genera with the key, divided by the size of all
// genera with the prefix.
func Evidence(genusSize, keyGenera, prefixSize int) float64 {
	return evidence(logSize(genusSize), keyGenera, prefixSize)
}

// evidence is Evidence with the log10 of the genus size.
func evidence(logSize float64, keyGenera, prefixSize int) float64 {
	if prefixSize == 0 {
		return 0
	}
	rarity := math.Log10(float64(prefixSize) / float64(max(keyGenera, 1)))
	return min(max(rarity-logSize, 0), MaxEvidence)
}

func logSize(genusSize int) float64 {
	return math.Log10(float64(max(genusSize, 1)))
}

// Scorer calculates scores of candidates.
type Scorer struct {
	Query      Query
	KeyGenera  map[string]int
	PrefixSize int
}

// Score sets the classification and the score of a candidate.
// Records must be set before.
func (s Scorer) Score(c *Candidate) {
	size := genusSize(c.Records)

	var res alib.Score
	res.MatchedNames = s.matched(c.Matches, size)
	res.Popularity = popularity(c.Matches)
	for _, m := range c.Matches {
		if pos, ok := m.Unit.Hints[c.Genus]; ok {
			res.GenusHints += HintWeight * m.Unit.HintFactors[c.Genus] /
				float64(pos)
		}
	}
	if n := len(c.Unmatched); n > 0 {
		res.UnmatchedNames = UnmatchedWeight * float64(n)
	}

	c.Record = s.record(c.Records)
	res.DominantClass = ClassWeight *
		fit(c.Records, s.Query.Input.DominantClass, true)
	res.MainTaxon = MainTaxonWeight *
		fit(c.Records, s.Query.Input.MainTaxon, false)
	if !known(c.Records) {
		res.UnknownGenus = UnknownGenusWeight
	}

	res.MatchedNames = round(res.MatchedNames)
	res.GenusHints = round(res.GenusHints)
	res.UnmatchedNames = round(res.UnmatchedNames)
	res.DominantClass = round(res.DominantClass)
	res.MainTaxon = round(res.MainTaxon)
	res.Popularity = round(res.Popularity)
	res.Total = round(res.MatchedNames + res.GenusHints +
		res.UnmatchedNames + res.DominantClass + res.MainTaxon +
		res.UnknownGenus + res.Popularity)
	c.Score = res
}

// Features returns raw values the score of a candidate is calculated
// from, so a client can recalculate the score with other weights.
// Records must be set before.
func (s Scorer) Features(c *Candidate) alib.Features {
	res := alib.Features{
		GenusSize:        genusSize(c.Records),
		DominantClassFit: int(fit(c.Records, s.Query.Input.DominantClass, true)),
		MainTaxonFit:     int(fit(c.Records, s.Query.Input.MainTaxon, false)),
		Known:            known(c.Records),
		UnmatchedUnits:   len(c.Unmatched),
		Matches:          make([]alib.MatchFeatures, 0, len(c.Matches)),
	}
	for _, m := range c.Matches {
		mf := alib.MatchFeatures{
			Name:         m.Unit.Names[0].Name,
			NamesNumber:  len(m.Unit.Names),
			MatchType:    m.MatchType,
			Key:          m.Key,
			KeyGenera:    s.KeyGenera[m.Key],
			DsNum:        m.Lookup.DsNum,
			HintPosition: m.Unit.Hints[c.Genus],
			HintFactor:   m.Unit.HintFactors[c.Genus],
		}
		for _, n := range m.Unit.Names {
			if n.ID != "" {
				mf.IDs = append(mf.IDs, n.ID)
			}
		}
		res.Matches = append(res.Matches, mf)
	}
	return res
}

// genusSize returns the size of a genus for the score: the largest size
// among its homonyms, at least 1.
func genusSize(recs []GenusRecord) int {
	res := 1
	for _, r := range recs {
		res = max(res, r.EpithetsNum)
	}
	return res
}

// matched returns the evidence of matched names. The anchor match pays
// only a part of the genus size, according to the genus prior. The
// anchor is the match where the prior adds most, so the result does not
// depend on the order of names.
func (s Scorer) matched(ms []Match, size int) float64 {
	full := logSize(size)
	part := full * (1 - GenusPriorWeight)

	var res, gain float64
	for _, m := range ms {
		w := matchWeights[m.MatchType]
		kg := s.KeyGenera[m.Key]
		ev := evidence(full, kg, s.PrefixSize) * w
		res += ev
		gain = max(gain, evidence(part, kg, s.PrefixSize)*w-ev)
	}
	return res + gain
}

// popularity returns the weighted mean of log10 of the number of data
// sources of matched names.
func popularity(ms []Match) float64 {
	if len(ms) == 0 {
		return 0
	}
	var res float64
	for _, m := range ms {
		res += math.Log10(float64(max(m.Lookup.DsNum, 1)))
	}
	return PopularityWeight * res / float64(len(ms))
}

// record picks the row whose classification is shown: the one that fits
// the dominant class and the main taxon best, then the one with most
// names.
func (s Scorer) record(recs []GenusRecord) GenusRecord {
	var res GenusRecord
	var best float64
	for i, r := range recs {
		rr := []GenusRecord{r}
		f := fit(rr, s.Query.Input.DominantClass, true) +
			fit(rr, s.Query.Input.MainTaxon, false)
		if i == 0 || f > best || (f == best && r.NamesNum > res.NamesNum) {
			res = r
			best = f
		}
	}
	return res
}

// fit returns 1 if any classification has the taxon, -1 if classifications
// are known and none has it, and 0 if the taxon or classifications are
// unknown. For a class, a classification is known only if it has the
// class rank.
func fit(recs []GenusRecord, taxon string, isClass bool) float64 {
	if taxon == "" {
		return 0
	}

	var known bool
	for _, r := range recs {
		if r.Classification == "" {
			continue
		}
		if slices.Contains(strings.Split(r.Classification, "|"), taxon) {
			return 1
		}
		if !isClass ||
			slices.Contains(strings.Split(r.ClassificationRanks, "|"), "class") {
			known = true
		}
	}
	if known {
		return -1
	}
	return 0
}

// known tells if any row of the genus is known. A genus missing from
// abbr_genera is unknown.
func known(recs []GenusRecord) bool {
	for _, r := range recs {
		if r.Known {
			return true
		}
	}
	return false
}

// Sort sorts candidates by total score, then by the number of matched
// names, then by genus.
func Sort(cs []*Candidate) {
	slices.SortStableFunc(cs, func(a, b *Candidate) int {
		if c := cmp.Compare(b.Score.Total, a.Score.Total); c != 0 {
			return c
		}
		if c := cmp.Compare(len(b.Matches), len(a.Matches)); c != 0 {
			return c
		}
		return cmp.Compare(a.Genus, b.Genus)
	})
}

func round(f float64) float64 {
	return math.Round(f*1000) / 1000
}
