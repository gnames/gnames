package gnames

import (
	"github.com/gnames/gnlib/domain/entity/gn"
	vlib "github.com/gnames/gnlib/domain/entity/verifier"
)

// GNames is the main use-case interface of the app. Its purpose to provide
// metadata of registered DataSources and provide functionality for
// verification (resolution/reconciliation) of name-strings to
// known to gnames scientific names, as well as providing data where
// these names occur.
type GNames interface {
	gn.Versioner
	Verify(params vlib.VerifyParams) ([]*vlib.Verification, error)
	DataSources(ids ...int) ([]*vlib.DataSource, error)
}
