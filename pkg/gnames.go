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
// When cfg.MatcherURL is empty, an embedded gnmatcher is initialised.
// When cfg.MatcherURL is set, an HTTP client to a remote service is used.
func New(
	cfg config.Config,
	vf verif.Verifier,
	vern vern.Vernaculars,
	sr srch.Searcher,
) (GNames, error) {
	var m gnmatcher.GNmatcher
	var err error
	if cfg.MatcherURL != "" {
		m = matcher.NewREST(cfg.MatcherURL)
	} else {
		m, err = matcher.NewLib(cfg)
		if err != nil {
			return nil, err
		}
	}
	return gnames{
		cfg:     cfg,
		vf:      vf,
		vern:    vern,
		sr:      sr,
		matcher: m,
	}, nil
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
