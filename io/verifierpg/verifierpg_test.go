package verifierpg_test

import (
	"context"
	"testing"

	"github.com/gnames/gnames/config"
	"github.com/gnames/gnames/io/matcher"
	"github.com/gnames/gnames/io/verifierpg"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/stretchr/testify/assert"
)

const restURL = "http://:8080/api/v0/"

func TestVerifyPGExact(t *testing.T) {
	names := []string{
		"Not name",
		"Bubo bubo",
		"Pomatomus",
		"Pardosa moesta",
		"Plantago major var major",
		"Cytospora ribis mitovirus 2",
		"A-shaped rods",
		"Alb. alba",
		"Pisonia grandis",
		"Acacia vestita may",
		"Candidatus Aenigmarchaeum subterraneum",
	}

	cfg := config.New(config.OptMatcherURL(restURL))
	vpg := verifierpg.New(cfg)
	mtr := matcher.New(cfg.MatcherURL)
	matches := mtr.MatchNames(names)

	input := vlib.Input{}
	mrs, err := vpg.MatchRecords(context.Background(), matches, input)
	assert.Nil(t, err)
	assert.Equal(t, 11, len(mrs))
}
