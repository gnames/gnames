package facet

import (
	"context"

	"github.com/gnames/gnames/ent/verifier"
	"github.com/gnames/gnquery/ent/search"
)

type Facet interface {
	Search(context.Context, search.Input) (map[string]*verifier.MatchRecord, error)
}
