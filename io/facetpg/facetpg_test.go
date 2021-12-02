package facetpg_test

import (
	"context"
	"testing"

	"github.com/gnames/gnames/config"
	"github.com/gnames/gnames/io/facetpg"
	"github.com/gnames/gnquery"
	"github.com/stretchr/testify/assert"
)

func TestSearchPG(t *testing.T) {
	cfg := config.New()
	fct := facetpg.New(cfg)
	inp := gnquery.New().Parse("g:Bubo asp:bub. yr:1700- tx:Aves")

	res, err := fct.Search(context.Background(), inp)
	assert.Nil(t, err)
	assert.True(t, len(res) > 5)
}
