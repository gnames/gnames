package abbr_test

import (
	"testing"

	"github.com/gnames/gnames/pkg/ent/abbr"
	alib "github.com/gnames/gnlib/ent/abbr"
	"github.com/stretchr/testify/assert"
)

func names(ss ...string) []alib.Name {
	res := make([]alib.Name, len(ss))
	for i, s := range ss {
		res[i] = alib.Name{Name: s}
	}
	return res
}

func TestNewQueryErrors(t *testing.T) {
	tests := []struct {
		msg   string
		names []string
	}{
		{"no names", nil},
		{"empty name", []string{" "}},
		{"full genus", []string{"C. elegans", "Caenorhabditis briggsae"}},
		{"no period", []string{"C elegans"}},
		{"lowercase", []string{"c. elegans"}},
		{"long abbreviation", []string{"Caen. elegans"}},
		{"different", []string{"C. elegans", "Cae. briggsae"}},
	}
	for _, v := range tests {
		_, err := abbr.NewQuery(alib.Input{Names: names(v.names...)})
		assert.NotNil(t, err, v.msg)
	}
}

func TestNewQueryUnits(t *testing.T) {
	tests := []struct {
		msg, name string
		epithets  []string
		keys      []string
	}{
		{"binomial", "C. elegans", []string{"elegans"}, []string{"c-elegans"}},
		{"author", "C. elegans Maupas, 1900", []string{"elegans"},
			[]string{"c-elegans"}},
		{"trinomial", "P. tigris altaica", []string{"tigris", "altaica"},
			[]string{"p-tigr", "p-altaic"}},
		{"rank", "Ca. alba var. rosea", []string{"alba", "rosea"},
			[]string{"c-alb", "c-rose"}},
		{"nominotypical", "B. bubo bubo", []string{"bubo"},
			[]string{"b-bub"}},
		{"abbreviated epithet", "P. t. altaica", []string{"altaica"},
			[]string{"p-altaic"}},
		{"hybrid", "M. × piperita", []string{"piperita"},
			[]string{"m-piperit"}},
		{"hyphen", "C. novae-angliae", []string{"novae-angliae"},
			[]string{"c-nouae-angli"}},
		{"no epithet", "P. ", []string(nil), []string(nil)},
		{"capitalized epithet", "P. Tigris", []string(nil), []string(nil)},
	}
	for _, v := range tests {
		q, err := abbr.NewQuery(alib.Input{Names: names(v.name)})
		assert.Nil(t, err, v.msg)
		assert.Equal(t, 1, len(q.Units), v.msg)
		assert.Equal(t, v.epithets, q.Units[0].Epithets, v.msg)
		assert.Equal(t, v.keys, q.Units[0].Keys, v.msg)
	}
}

func TestNewQueryDefaults(t *testing.T) {
	assert := assert.New(t)
	q, err := abbr.NewQuery(alib.Input{Names: names("C. elegans")})
	assert.Nil(err)
	assert.Equal("C.", q.Abbreviation)
	assert.Equal("C", q.Prefix)
	assert.Equal(abbr.DefaultTolerance, q.Tolerance)
	assert.Equal(abbr.DefaultGeneraLimit, q.GeneraLimit)

	tol := 150
	q, err = abbr.NewQuery(alib.Input{
		Names:       names("Cae. elegans"),
		Tolerance:   &tol,
		GeneraLimit: 1000,
	})
	assert.Nil(err)
	assert.Equal("Cae", q.Prefix)
	assert.Equal(100, q.Tolerance)
	assert.Equal(abbr.MaxGeneraLimit, q.GeneraLimit)

	tol = 0
	q, err = abbr.NewQuery(alib.Input{Names: names("C. elegans"), Tolerance: &tol})
	assert.Nil(err)
	assert.Equal(0, q.Tolerance)
}

func TestNewQueryGrouping(t *testing.T) {
	assert := assert.New(t)
	inp := alib.Input{Names: []alib.Name{
		{ID: "1", Name: "C. elegans", GenusHints: []string{
			"Mus", "Cyclops", "Caenorhabditis",
		}},
		{ID: "2", Name: "C. briggsae"},
		{ID: "3", Name: "C. elegans", GenusHints: []string{"Caenorhabditis"}},
	}}
	q, err := abbr.NewQuery(inp)
	assert.Nil(err)
	assert.Equal(2, len(q.Units))
	u := q.Units[0]
	assert.Equal(2, len(u.Names))
	assert.Equal("3", u.Names[1].ID)
	assert.Equal(map[string]int{"Cyclops": 1, "Caenorhabditis": 1}, u.Hints)
	assert.Equal([]string{"c-briggs", "c-elegans"}, q.Keys())
}
