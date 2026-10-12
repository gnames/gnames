package abbr_test

import (
	"strings"
	"testing"

	"github.com/gnames/gnames/pkg/ent/abbr"
	alib "github.com/gnames/gnlib/ent/abbr"
	"github.com/stretchr/testify/assert"
)

func TestMatchType(t *testing.T) {
	tests := []struct {
		input, db string
		res       alib.MatchTypeValue
	}{
		{"tigris altaica", "Panthera tigris altaica", alib.Full},
		{"bubo", "Bubo bubo bubo", alib.Full},
		{"tigris altaica", "Panthera altaica", alib.EpithetMoved},
		{"altaica", "Panthera tigris altaica", alib.EpithetMoved},
		{"tigris altaica", "Panthera pardus altaica", alib.EpithetMoved},
		{"tigris altaica", "Panthera tigris", alib.SpeciesLevel},
		{"tigris", "Panthera tigris altaica", alib.SpeciesLevel},
		{"tigris altaica", "Panthera tigris amurensis", alib.SpeciesLevel},
		{"tigris altaica", "Panthera onca tigris", alib.NoMatch},
		{"tigris", "Panthera", alib.NoMatch},
		{"", "Panthera tigris", alib.NoMatch},
	}
	for _, v := range tests {
		res := abbr.MatchType(strings.Fields(v.input), abbr.Epithets(v.db))
		assert.Equal(t, v.res, res, v.input+" vs "+v.db)
	}
}

func TestEpithets(t *testing.T) {
	tests := []struct {
		can string
		res []string
	}{
		{"Bubo", nil},
		{"Bubo bubo", []string{"bubo"}},
		{"Bubo bubo bubo", []string{"bubo"}},
		{"Aus bus bus bus", []string{"bus"}},
		{"Aus bus cus", []string{"bus", "cus"}},
		{"Aus bus cus cus", []string{"bus", "cus"}},
	}
	for _, v := range tests {
		assert.Equal(t, v.res, abbr.Epithets(v.can), v.can)
	}
}

func TestExactEpithets(t *testing.T) {
	tests := []struct {
		input, can string
		mt         alib.MatchTypeValue
		res        bool
	}{
		{"alba", "Carex alba", alib.Full, true},
		{"alba", "Carex albus", alib.Full, false},
		{"alba rosea", "Carex rosea", alib.EpithetMoved, true},
		{"alba rosea", "Carex roseus", alib.EpithetMoved, false},
		{"alba rosea", "Carex alba", alib.SpeciesLevel, true},
		{"alba rosea", "Carex albus", alib.SpeciesLevel, false},
	}
	for _, v := range tests {
		res := abbr.ExactEpithets(
			strings.Fields(v.input), abbr.Epithets(v.can), v.mt,
		)
		assert.Equal(t, v.res, res, v.input+" vs "+v.can)
	}
}
