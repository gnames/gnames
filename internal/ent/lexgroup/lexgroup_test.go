package lexgroup_test

import (
	"os"
	"testing"

	"github.com/gnames/gnames/internal/ent/lexgroup"
	"github.com/gnames/gnfmt"
	"github.com/gnames/gnlib/ent/verifier"
	"github.com/stretchr/testify/assert"
)

var enc = gnfmt.GNjson{}

func TestLexGroupEmpty(t *testing.T) {
	assert := assert.New(t)
	name := verifier.Name{}
	grps := lexgroup.NameToLexicalGroups(name)
	assert.Equal(0, len(grps))
}

func TestLexGroupBestResult(t *testing.T) {
	assert := assert.New(t)
	txt, err := os.ReadFile("../../testdata/lexgroup0.json")
	assert.Nil(err)
	_ = txt
	var n verifier.Name
	err = enc.Decode(txt, &n)
	assert.Nil(err)
	grps := lexgroup.NameToLexicalGroups(n)
	assert.Equal(1, len(grps))
	assert.Equal("Bubo bubo (Linnaeus, 1758)", grps[0].Name)
	assert.Equal([]string{"Bubo bubo (Linnaeus, 1758)"},
		grps[0].LexicalVariants)
}

func TestLexGroupVirus(t *testing.T) {
	assert := assert.New(t)
	txt, err := os.ReadFile("../../testdata/lexgroup1.json")
	assert.Nil(err)
	var n verifier.Name
	err = enc.Decode(txt, &n)
	assert.Nil(err)
	grps := lexgroup.NameToLexicalGroups(n)
	assert.Equal(1, len(grps))
	assert.Equal("Tobacco mosaic virus", grps[0].Name)

	names := []string{
		"Tobacco mosaic virus",
		"Tobacco mosaic virus (strain O)",
		"Tobacco mosaic virus (vulgare)",
		"Tobacco mosaic virus group",
		"Tobacco mosaic virus strain 06",
		"Tobacco mosaic virus strain B395A",
		"Tobacco mosaic virus strain Dahlemense",
		"Tobacco mosaic virus strain ER",
		"Tobacco mosaic virus strain HR",
		"Tobacco mosaic virus strain Kokubu",
		"Tobacco mosaic virus strain Korean",
		"Tobacco mosaic virus strain Ohio V",
		"Tobacco mosaic virus strain OM",
		"Tobacco mosaic virus strain Rakkyo",
		"Tobacco mosaic virus strain tomatoL",
		"Tobacco mosaic virus strain tomato/L",
	}
	assert.Equal(names, grps[0].LexicalVariants)
}

func TestLexGroupCarex(t *testing.T) {
	assert := assert.New(t)
	txt, err := os.ReadFile("../../testdata/lexgroup2.json")
	assert.Nil(err)
	var n verifier.Name
	err = enc.Decode(txt, &n)
	assert.Nil(err)
	grps := lexgroup.NameToLexicalGroups(n)
	assert.Equal(3, len(grps))
}

func TestLexGroupStrongylapsis(t *testing.T) {
	assert := assert.New(t)
	txt, err := os.ReadFile("../../testdata/lexgroup3.json")
	assert.Nil(err)
	var n verifier.Name
	err = enc.Decode(txt, &n)
	assert.Nil(err)
	grps := lexgroup.NameToLexicalGroups(n)
	assert.Equal(41, len(grps))
}
