package abbr_test

import (
	"math"
	"testing"

	"github.com/gnames/gnames/pkg/ent/abbr"
	alib "github.com/gnames/gnlib/ent/abbr"
	"github.com/stretchr/testify/assert"
)

func lookup(key, stem string) abbr.Lookup {
	return abbr.Lookup{
		Key:             key,
		Genus:           abbr.Epithets("x " + stem)[0],
		CanonicalStemID: stem,
		CanonicalStem:   stem,
		DsNum:           1,
	}
}

var lookups = []abbr.Lookup{
	lookup("c-elegans", "Caenorhabditis elegans"),
	lookup("c-briggs", "Caenorhabditis briggs"),
	lookup("c-remane", "Caenorhabditis remane"),
	lookup("c-elegans", "Cucumis elegans"),
	lookup("c-briggs", "Cucumis briggs"),
	lookup("c-elegans", "Cyperus elegans"),
	lookup("c-elegans", "Cyperus remane elegans"),
}

func TestCandidates(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		msg       string
		names     []string
		tolerance int
		genera    []string
		keyGenera map[string]int
	}{
		{
			msg:       "all match",
			names:     []string{"C. elegans", "C. briggsae", "C. remanei"},
			tolerance: 20,
			genera:    []string{"Caenorhabditis"},
			keyGenera: map[string]int{
				"c-elegans": 3, "c-briggs": 2, "c-remane": 1,
			},
		},
		{
			msg:       "tolerance",
			names:     []string{"C. elegans", "C. briggsae", "C. remanei"},
			tolerance: 34,
			genera:    []string{"Caenorhabditis", "Cucumis"},
		},
		{
			msg:       "prefix",
			names:     []string{"Cy. elegans"},
			tolerance: 20,
			genera:    []string{"Cyperus"},
			keyGenera: map[string]int{"c-elegans": 1},
		},
		{
			msg:       "one name",
			names:     []string{"C. elegans"},
			tolerance: 20,
			genera:    []string{"Caenorhabditis", "Cucumis", "Cyperus"},
		},
	}

	for _, v := range tests {
		q, err := abbr.NewQuery(alib.Input{
			Names:     names(v.names...),
			Tolerance: &v.tolerance,
		})
		assert.Nil(err, v.msg)
		cands, keyGenera := q.Candidates(lookups)
		var genera []string
		for _, c := range cands {
			genera = append(genera, c.Genus)
		}
		assert.Equal(v.genera, genera, v.msg)
		if v.keyGenera != nil {
			assert.Equal(v.keyGenera, keyGenera, v.msg)
		}
	}
}

func TestBestMatch(t *testing.T) {
	assert := assert.New(t)
	q, err := abbr.NewQuery(alib.Input{Names: names("C. elegans")})
	assert.Nil(err)
	cands, _ := q.Candidates(lookups)
	c := cands[2]
	assert.Equal("Cyperus", c.Genus)
	assert.Equal(1, len(c.Matches))
	// the full match wins over the epithet moved match
	assert.Equal("Cyperus elegans", c.Matches[0].Lookup.CanonicalStem)
	assert.Equal(alib.Full, c.Matches[0].MatchType)
}

func TestEvidence(t *testing.T) {
	assert := assert.New(t)
	assert.InDelta(-math.Log10(0.02), abbr.Evidence(10, 2, 1000), 1e-9)
	assert.Equal(0.0, abbr.Evidence(1000, 10, 1000))
	assert.Equal(abbr.MaxEvidence, abbr.Evidence(1, 1, 1_000_000_000))
	assert.Equal(0.0, abbr.Evidence(10, 2, 0))
}

func TestScore(t *testing.T) {
	assert := assert.New(t)
	bird := abbr.GenusRecord{
		Genus:               "Morus",
		Kingdom:             "Animalia",
		EpithetsNum:         10,
		NamesNum:            5,
		Classification:      "Animalia|Chordata|Aves|Suliformes|Sulidae|Morus",
		ClassificationRanks: "kingdom|phylum|class|order|family|genus",
		DataSourceID:        13,
		Known:               true,
	}
	plant := abbr.GenusRecord{
		Genus:               "Morus",
		Kingdom:             "Plantae",
		EpithetsNum:         10,
		NamesNum:            280,
		Classification:      "Plantae|Tracheophyta|Magnoliopsida|Rosales|Morus",
		ClassificationRanks: "kingdom|phylum|class|order|genus",
		DataSourceID:        13,
		Known:               true,
	}
	noClass := abbr.GenusRecord{
		Genus:               "Morus",
		EpithetsNum:         10,
		Classification:      "Rosales|Moraceae|Morus",
		ClassificationRanks: "order|family|genus",
		DataSourceID:        13,
		Known:               true,
	}
	unknown := abbr.GenusRecord{Genus: "Morus", EpithetsNum: 10}

	tests := []struct {
		msg            string
		recs           []abbr.GenusRecord
		class, taxon   string
		classScore     float64
		taxonScore     float64
		kingdom        string
		hints          float64
		hintsFirstName []string
	}{
		{"no context", []abbr.GenusRecord{bird, plant}, "", "",
			0, 0, "Plantae", 0, nil},
		{"bird", []abbr.GenusRecord{bird, plant}, "Aves", "Chordata",
			0.5, 0.5, "Animalia", 0, nil},
		{"plant", []abbr.GenusRecord{bird, plant}, "Magnoliopsida", "",
			0.5, 0, "Plantae", 0, nil},
		{"other class", []abbr.GenusRecord{bird, plant}, "Insecta", "Insecta",
			-0.5, -0.5, "Plantae", 0, nil},
		{"no class rank", []abbr.GenusRecord{noClass}, "Insecta", "Insecta",
			0, -0.5, "", 0, nil},
		{"unknown", []abbr.GenusRecord{unknown}, "Insecta", "Insecta",
			0, 0, "", 0, nil},
		{"hint", []abbr.GenusRecord{plant}, "", "",
			0, 0, "Plantae", 1.5, []string{"Mus", "Cyclops", "Morus"}},
		{"name hint", []abbr.GenusRecord{plant}, "", "",
			0, 0, "Plantae", 2.25, []string{"Mus", "Morus nigra"}},
		{"same name hint", []abbr.GenusRecord{plant}, "", "",
			0, 0, "Plantae", 3, []string{"Mus musculus", "Morus albus"}},
		{"genus, then same name", []abbr.GenusRecord{plant}, "", "",
			0, 0, "Plantae", 3, []string{"Mus", "Morus", "Morus alba"}},
	}

	for _, v := range tests {
		q, err := abbr.NewQuery(alib.Input{
			Names: []alib.Name{
				{Name: "M. alba", GenusHints: v.hintsFirstName},
				{Name: "M. nigra"},
				{Name: "M. rubra"},
			},
			DominantClass: v.class,
			MainTaxon:     v.taxon,
		})
		assert.Nil(err, v.msg)
		c := &abbr.Candidate{
			Genus: "Morus",
			Matches: []abbr.Match{
				{Unit: q.Units[0], MatchType: alib.Full, Key: "m-alb"},
				{Unit: q.Units[1], MatchType: alib.EpithetMoved, Key: "m-nigr"},
			},
			Unmatched: []*abbr.Unit{q.Units[2]},
			Records:   v.recs,
		}
		s := abbr.Scorer{
			Query:      q,
			KeyGenera:  map[string]int{"m-alb": 10, "m-nigr": 1},
			PrefixSize: 100_000,
		}
		s.Score(c)

		// alba: 10*10/100000 -> 3; nigra: 10*1/100000 -> 4 * 0.8.
		// The genus prior adds the size (1) to alba, where it adds most.
		assert.InDelta(7.2, c.Score.MatchedNames, 1e-9, v.msg)
		assert.Equal(-1.0, c.Score.UnmatchedNames, v.msg)
		assert.Equal(v.hints, c.Score.GenusHints, v.msg)
		assert.Equal(v.classScore, c.Score.DominantClass, v.msg)
		assert.Equal(v.taxonScore, c.Score.MainTaxon, v.msg)
		assert.Equal(v.kingdom, c.Record.Kingdom, v.msg)
		if v.recs[0].Known {
			assert.Equal(0.0, c.Score.UnknownGenus, v.msg)
		} else {
			assert.Equal(abbr.UnknownGenusWeight, c.Score.UnknownGenus, v.msg)
		}
		sum := c.Score.MatchedNames + c.Score.GenusHints +
			c.Score.UnmatchedNames + c.Score.DominantClass +
			c.Score.MainTaxon + c.Score.UnknownGenus + c.Score.Popularity
		assert.InDelta(sum, c.Score.Total, 1e-9, v.msg)
	}
}

func TestGenusPrior(t *testing.T) {
	assert := assert.New(t)
	q, err := abbr.NewQuery(alib.Input{
		Names: []alib.Name{{Name: "C. elegans"}, {Name: "C. briggsae"}},
	})
	assert.Nil(err)
	s := abbr.Scorer{
		Query:      q,
		KeyGenera:  map[string]int{"c-elegans": 1000, "c-briggs": 1},
		PrefixSize: 1_000_000,
	}
	score := func(size int, units int) float64 {
		c := &abbr.Candidate{
			Genus:   "Caenorhabditis",
			Records: []abbr.GenusRecord{{EpithetsNum: size, Known: true}},
		}
		keys := []string{"c-elegans", "c-briggs"}
		for i := range units {
			c.Matches = append(c.Matches, abbr.Match{
				Unit: q.Units[i], MatchType: alib.Full, Key: keys[i],
			})
		}
		s.Score(c)
		return c.Score.MatchedNames
	}

	// With one name the size cancels, even for a genus larger than the
	// chance of the match.
	assert.Equal(3.0, score(1, 1))
	assert.Equal(3.0, score(100, 1))
	assert.Equal(3.0, score(10_000, 1))

	// The second name pays the size (2): briggsae 6 - 2, elegans 3 - 2,
	// and the prior adds the size once.
	assert.Equal(7.0, score(100, 2))
}

func TestPopularity(t *testing.T) {
	assert := assert.New(t)
	q, err := abbr.NewQuery(alib.Input{
		Names: []alib.Name{{Name: "C. elegans"}, {Name: "C. briggsae"}},
	})
	assert.Nil(err)
	s := abbr.Scorer{Query: q, PrefixSize: 1_000_000}

	c := &abbr.Candidate{
		Genus:   "Caenorhabditis",
		Records: []abbr.GenusRecord{{EpithetsNum: 79, Known: true}},
		Matches: []abbr.Match{
			{Unit: q.Units[0], MatchType: alib.Full, Lookup: abbr.Lookup{DsNum: 100}},
			{Unit: q.Units[1], MatchType: alib.Full, Lookup: abbr.Lookup{DsNum: 1}},
		},
	}
	s.Score(c)
	// mean of log10(100) and log10(1), times 0.5
	assert.Equal(0.5, c.Score.Popularity)

	c.Matches = nil
	s.Score(c)
	assert.Equal(0.0, c.Score.Popularity)
}

// scoreFromFeatures recalculates the total score from features, as a
// client that tunes weights does.
func scoreFromFeatures(f alib.Features, prefixSize int) float64 {
	matchWeights := map[alib.MatchTypeValue]float64{
		alib.Full: 1.0, alib.EpithetMoved: 0.8, alib.SpeciesLevel: 0.6,
	}
	// With GenusPriorWeight 1 the anchor pays no size, as a genus of size 1.
	part := 1

	var matched, gain, pop, hints float64
	for _, m := range f.Matches {
		w := matchWeights[m.MatchType]
		ev := abbr.Evidence(f.GenusSize, m.KeyGenera, prefixSize) * w
		matched += ev
		gain = max(gain, abbr.Evidence(part, m.KeyGenera, prefixSize)*w-ev)
		pop += math.Log10(float64(max(m.DsNum, 1)))
		if m.HintPosition > 0 {
			hints += abbr.HintWeight * m.HintFactor / float64(m.HintPosition)
		}
	}
	res := matched + gain + hints +
		abbr.UnmatchedWeight*float64(f.UnmatchedUnits) +
		abbr.ClassWeight*float64(f.DominantClassFit) +
		abbr.MainTaxonWeight*float64(f.MainTaxonFit)
	if len(f.Matches) > 0 {
		res += abbr.PopularityWeight * pop / float64(len(f.Matches))
	}
	if !f.Known {
		res += abbr.UnknownGenusWeight
	}
	return res
}

func TestFeatures(t *testing.T) {
	assert := assert.New(t)
	bird := abbr.GenusRecord{
		EpithetsNum:         12,
		Classification:      "Animalia|Chordata|Aves|Suliformes|Sulidae|Morus",
		ClassificationRanks: "kingdom|phylum|class|order|family|genus",
		Known:               true,
	}
	plant := abbr.GenusRecord{
		EpithetsNum:         30,
		Classification:      "Plantae|Tracheophyta|Magnoliopsida|Rosales|Morus",
		ClassificationRanks: "kingdom|phylum|class|order|genus",
		Known:               true,
	}
	unknown := abbr.GenusRecord{EpithetsNum: 3}

	tests := []struct {
		msg          string
		recs         []abbr.GenusRecord
		class, taxon string
		classFit     int
		taxonFit     int
		hints        []string
		hintPos      int
		hintFactor   float64
	}{
		{"no context", []abbr.GenusRecord{bird, plant}, "", "", 0, 0, nil, 0, 0},
		{"bird", []abbr.GenusRecord{bird, plant}, "Aves", "Chordata", 1, 1,
			[]string{"Mus", "Morus"}, 2, 1},
		{"other class", []abbr.GenusRecord{plant}, "Insecta", "Insecta",
			-1, -1, []string{"Morus"}, 1, 1},
		{"name hint", []abbr.GenusRecord{plant}, "", "", 0, 0,
			[]string{"Morus nigra"}, 1, 1.5},
		{"same name hint", []abbr.GenusRecord{plant}, "", "", 0, 0,
			[]string{"Mus", "Morus alba"}, 2, 2},
		{"unknown", []abbr.GenusRecord{unknown}, "Aves", "Aves", 0, 0, nil, 0, 0},
	}

	for _, v := range tests {
		q, err := abbr.NewQuery(alib.Input{
			Names: []alib.Name{
				{ID: "1", Name: "M. alba", GenusHints: v.hints},
				{ID: "2", Name: "M. nigra"},
				{ID: "3", Name: "M. alba"},
				{ID: "4", Name: "M. rubra"},
			},
			DominantClass: v.class,
			MainTaxon:     v.taxon,
		})
		assert.Nil(err, v.msg)
		c := &abbr.Candidate{
			Genus: "Morus",
			Matches: []abbr.Match{
				{Unit: q.Units[0], MatchType: alib.Full, Key: "m-alb",
					Lookup: abbr.Lookup{DsNum: 20}},
				{Unit: q.Units[1], MatchType: alib.EpithetMoved, Key: "m-nigr",
					Lookup: abbr.Lookup{DsNum: 3}},
			},
			Unmatched: []*abbr.Unit{q.Units[2]},
			Records:   v.recs,
		}
		s := abbr.Scorer{
			Query:      q,
			KeyGenera:  map[string]int{"m-alb": 10, "m-nigr": 1},
			PrefixSize: 100_000,
		}
		s.Score(c)
		f := s.Features(c)

		assert.Equal(genusSizeOf(v.recs), f.GenusSize, v.msg)
		assert.Equal(v.classFit, f.DominantClassFit, v.msg)
		assert.Equal(v.taxonFit, f.MainTaxonFit, v.msg)
		assert.Equal(v.recs[0].Known, f.Known, v.msg)
		assert.Equal(1, f.UnmatchedUnits, v.msg)
		assert.Equal(2, len(f.Matches), v.msg)

		alba := f.Matches[0]
		assert.Equal("M. alba", alba.Name, v.msg)
		assert.Equal([]string{"1", "3"}, alba.IDs, v.msg)
		assert.Equal(2, alba.NamesNumber, v.msg)
		assert.Equal(alib.Full, alba.MatchType, v.msg)
		assert.Equal("m-alb", alba.Key, v.msg)
		assert.Equal(10, alba.KeyGenera, v.msg)
		assert.Equal(20, alba.DsNum, v.msg)
		assert.Equal(v.hintPos, alba.HintPosition, v.msg)
		assert.Equal(v.hintFactor, alba.HintFactor, v.msg)
		assert.Equal(0, f.Matches[1].HintPosition, v.msg)

		// Features are enough to recalculate the score. Parts of the
		// score are rounded, so the totals can differ a little.
		assert.InDelta(c.Score.Total, scoreFromFeatures(f, s.PrefixSize), 0.01, v.msg)
	}
}

func genusSizeOf(recs []abbr.GenusRecord) int {
	res := 1
	for _, r := range recs {
		res = max(res, r.EpithetsNum)
	}
	return res
}

func TestSort(t *testing.T) {
	assert := assert.New(t)
	cs := []*abbr.Candidate{
		{Genus: "B", Score: alib.Score{Total: 1}},
		{Genus: "C", Score: alib.Score{Total: 2}},
		{Genus: "A", Score: alib.Score{Total: 1}, Matches: make([]abbr.Match, 2)},
		{Genus: "D", Score: alib.Score{Total: 1}},
	}
	abbr.Sort(cs)
	var res []string
	for _, c := range cs {
		res = append(res, c.Genus)
	}
	assert.Equal([]string{"C", "A", "B", "D"}, res)
}
