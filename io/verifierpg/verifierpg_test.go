package verifierpg_test

import (
	"context"
	"testing"

	"github.com/gnames/gnames/config"
	"github.com/gnames/gnames/io/matcher"
	"github.com/gnames/gnames/io/verifierpg"
	"github.com/stretchr/testify/assert"
)

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
	cfg := config.New()
	vpg := verifierpg.New(cfg)
	mtr := matcher.NewGNmatcher(cfg.MatcherURL)
	matches := mtr.MatchNames(names)

	mrs, err := vpg.MatchRecords(context.Background(), matches)
	assert.Nil(t, err)
	assert.Equal(t, len(mrs), 11)
}
