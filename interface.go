package gnames

import (
	"github.com/gnames/gnlib/domain/entity/gn"
	vlib "github.com/gnames/gnlib/domain/entity/verifier"
)

type GNames interface {
	gn.Versioner
	Verify(params vlib.VerifyParams) ([]*vlib.Verification, error)
	DataSources(ids ...int) ([]*vlib.DataSource, error)
}
