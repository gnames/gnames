package gnames

import (
	"testing"

	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/stretchr/testify/assert"
)

func TestMetaStats(t *testing.T) {
	assert := assert.New(t)
	ranks := "kingdom|phylum|class|order|family|genus|species"
	names := []vlib.Name{
		statsName("1", 1,
			"Animalia|Chordata|Aves|Strigiformes|Strigidae|Bubo|Bubo bubo", ranks),
		statsName("2", 1,
			"Animalia|Chordata|Aves|Passeriformes|Corvidae|Corvus|Corvus corax", ranks),
		statsName("3", 1,
			"Animalia|Chordata|Mammalia|Carnivora|Felidae|Puma|Puma concolor", ranks),
		statsName("4", 1,
			"Plantae|Tracheophyta|Magnoliopsida|Rosales|Moraceae|Morus|Morus alba", ranks),
		// names that are not from Catalogue of Life are ignored
		statsName("5", 11,
			"Animalia|Chordata|Mammalia|Primates|Hominidae|Homo|Homo sapiens", ranks),
	}

	res := meta(vlib.Input{WithStats: true}, names)
	assert.Equal(4, res.StatsNamesNum)
	assert.Equal("Animalia", res.Kingdom)
	assert.Equal("Aves", res.Class)
	assert.Equal(float32(0.5), res.ClassPercentage)
	assert.Equal([]vlib.Class{
		{ClassName: "Aves", NamesNumber: 2, Percentage: 0.5},
		{ClassName: "Magnoliopsida", NamesNumber: 1, Percentage: 0.25},
		{ClassName: "Mammalia", NamesNumber: 1, Percentage: 0.25},
	}, res.Classes)

	res = meta(vlib.Input{}, names)
	assert.Empty(res.Class)
	assert.Zero(res.ClassPercentage)
	assert.Nil(res.Classes)
}

func statsName(id string, dsID int, path, ranks string) vlib.Name {
	return vlib.Name{
		ID: id,
		BestResult: &vlib.ResultData{
			DataSourceID:        dsID,
			ClassificationPath:  path,
			ClassificationRanks: ranks,
		},
	}
}
