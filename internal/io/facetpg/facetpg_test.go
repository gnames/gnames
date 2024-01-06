package facetpg_test

import (
	"context"
	"testing"

	"github.com/gnames/gnames/internal/io/facetpg"
	"github.com/gnames/gnames/pkg/config"
	"github.com/gnames/gnquery"
	"github.com/stretchr/testify/assert"
)

func TestSearchPG(t *testing.T) {
	assert := assert.New(t)
	cfg := config.New()
	fct, err := facetpg.New(cfg)
	assert.Nil(err)
	inp := gnquery.New().Parse("g:Bubo asp:bub. yr:1700- tx:Aves")

	res, err := fct.Search(context.Background(), inp)
	assert.Nil(err)
	assert.True(len(res) > 5)
}
