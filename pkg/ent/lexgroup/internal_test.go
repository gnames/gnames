package lexgroup

import (
	"testing"

	"github.com/gnames/gnparser"
	"github.com/stretchr/testify/assert"
)

var p = gnparser.New(
	gnparser.NewConfig(
		gnparser.OptWithDetails(true),
	),
)

func TestGetAuthors(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		msg    string
		name   string
		isNil  bool
		isComb bool
		orig   string
		comb   string
	}{
		{"no au", "Bubo bubo", true, false, "", ""},
		{"no parse", "123", true, false, "", ""},
		{"virus", "Tobacco Mosaic Virus", true, false, "", ""},
		{"simple", "Pardosa moesta Banks, 1892 ", false, false, "B", ""},
		{"combo", "Carex scirpoidea Michx. subsp. convoluta (Kük.) D.A. Dunlop", false, true, "K", "D"},
		{"multiple", "Navicula rhomboides var. lineolata (Ehrenberg) Cleve & Möller, 1879", false, true, "E", "CM"},
	}

	for _, v := range tests {
		parsed := p.ParseName(v.name)
		res := getAuthors(parsed)
		assert.Equal(v.isNil, res == nil)
		if res == nil {
			continue
		}
		assert.Equal(v.isComb, res.isCombination)
		assert.Equal(v.orig, res.orig)
		assert.Equal(v.comb, res.comb)
	}
}
