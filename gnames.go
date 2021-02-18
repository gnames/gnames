package gnames

import (
	"context"

	"github.com/gnames/gnames/config"
	"github.com/gnames/gnames/ent/score"
	"github.com/gnames/gnames/ent/verifier"
	"github.com/gnames/gnames/io/matcher"
	"github.com/gnames/gnlib/ent/gnvers"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnmatcher"
	log "github.com/sirupsen/logrus"
)

type gnames struct {
	cfg     config.Config
	vf      verifier.Verifier
	matcher gnmatcher.GNmatcher
}

// NewGNames is a constructor that returns implmentation of GNames interface.
func NewGNames(cnf config.Config, vf verifier.Verifier) GNames {
	return gnames{
		cfg:     cnf,
		vf:      vf,
		matcher: matcher.NewGNmatcher(cnf.MatcherURL),
	}
}

func (g gnames) GetVersion() gnvers.Version {
	return gnvers.Version{
		Version: Version,
		Build:   Build,
	}
}

func (g gnames) Verify(
	ctx context.Context,
	params vlib.VerifyParams,
) ([]*vlib.Verification, error) {
	log.Printf("Verifying %d name-strings.", len(params.NameStrings))
	res := make([]*vlib.Verification, len(params.NameStrings))

	matches := g.matcher.MatchNames(params.NameStrings)

	var errString string
	matchRecords, err := g.vf.MatchRecords(ctx, matches)
	if err != nil {
		errString = err.Error()
	}

	for i, v := range matches {
		if mr, ok := matchRecords[v.ID]; ok {
			s := score.NewScore()
			s.SortResults(mr)
			item := vlib.Verification{
				InputID:          mr.InputID,
				Input:            mr.Input,
				MatchType:        mr.MatchType,
				Curation:         mr.Curation,
				DataSourcesNum:   mr.DataSourcesNum,
				BestResult:       s.BestResult(mr),
				PreferredResults: s.PreferredResults(params.PreferredSources, mr),
				Error:            errString,
			}

			res[i] = &item
		} else {
			log.Warnf("Cannot find record for '%s'.", v.Name)
		}
	}
	return res, nil
}

func (g gnames) DataSources(ids ...int) ([]*vlib.DataSource, error) {
	return g.vf.DataSources(ids...)
}
