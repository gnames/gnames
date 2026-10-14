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

		// alba: 10*10/100000 -> 3; nigra: 10*1/100000 -> 4 * 0.8
		assert.InDelta(6.2, c.Score.MatchedNames, 1e-9, v.msg)
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
			c.Score.MainTaxon + c.Score.UnknownGenus
		assert.InDelta(sum, c.Score.Total, 1e-9, v.msg)
	}
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
