package gnames

import (
	"github.com/gnames/gnames/internal/io/matcher"
	"github.com/gnames/gnames/pkg/config"
	"github.com/gnames/gnames/pkg/ent/srch"
	"github.com/gnames/gnames/pkg/ent/verif"
	"github.com/gnames/gnames/pkg/ent/vern"
	"github.com/gnames/gnlib/ent/gnvers"
	gnmatcher "github.com/gnames/gnmatcher/pkg"
)

type gnames struct {
	cfg     config.Config
	vf      verif.Verifier
	vern    vern.Vernaculars
	sr      srch.Searcher
	matcher gnmatcher.GNmatcher
}

// New is a constructor that returns implmentation of GNames interface.
func New(
	cfg config.Config,
	vf verif.Verifier,
	vern vern.Vernaculars,
	sr srch.Searcher,
) GNames {
	return gnames{
		cfg:     cfg,
		vf:      vf,
		vern:    vern,
		sr:      sr,
		matcher: matcher.New(cfg.MatcherURL),
	}
}

func (g gnames) GetVersion() gnvers.Version {
	return gnvers.Version{
		Version: Version,
		Build:   Build,
	}
}

func (g gnames) GetConfig() config.Config {
	return g.cfg
}
