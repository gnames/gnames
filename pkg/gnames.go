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

// Option is a functional option for the gnames constructor.
type Option func(*gnames)

// WithMatcher injects a pre-built GNmatcher, bypassing the default
// initialization. Useful for testing or when sharing a single instance.
func WithMatcher(m gnmatcher.GNmatcher) Option {
	return func(g *gnames) {
		g.matcher = m
	}
}

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
// An optional WithMatcher option can be passed to inject a pre-built matcher.
func New(
	cfg config.Config,
	vf verif.Verifier,
	vern vern.Vernaculars,
	sr srch.Searcher,
	opts ...Option,
) (GNames, error) {
	g := &gnames{
		cfg:  cfg,
		vf:   vf,
		vern: vern,
		sr:   sr,
	}

	for _, opt := range opts {
		opt(g)
	}

	if g.matcher == nil {
		var err error
		if cfg.MatcherURL != "" {
			g.matcher = matcher.NewREST(cfg.MatcherURL)
		} else {
			g.matcher, err = matcher.NewLib(cfg)
			if err != nil {
				return nil, err
			}
		}
	}

	return *g, nil
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
