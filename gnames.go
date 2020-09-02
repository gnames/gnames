package gnames

import (
	"github.com/gnames/gnames/config"
	"github.com/gnames/gnames/model"
)

type GNames struct {
	Config config.Config
}

func NewGNames(cnf config.Config) GNames {
	return GNames{Config: cnf}
}

func (gn GNames) Verify(names model.VerifyParams) []model.Verification {
	var vs []model.Verification
	return vs
}
