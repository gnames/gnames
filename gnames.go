package gnames

import (
	"github.com/gnames/gnames/config"
	"github.com/gnames/gnames/entity/score"
	"github.com/gnames/gnames/entity/verifier"
	"github.com/gnames/gnames/io/matcher"
	"github.com/gnames/gnlib/domain/entity/gn"
	vlib "github.com/gnames/gnlib/domain/entity/verifier"
	"github.com/gnames/gnmatcher"
	log "github.com/sirupsen/logrus"
)

type gnames struct {
	cfg     config.Config
	vf      verifier.Verifier
	matcher gnmatcher.GNMatcher
}

// NewGNames is a constructor that returns implmentation of GNames interface.
func NewGNames(cnf config.Config, vf verifier.Verifier) GNames {
	return gnames{
		cfg:     cnf,
		vf:      vf,
		matcher: matcher.NewGNMatcher(cnf.MatcherURL),
	}
}

func (g gnames) GetVersion() gn.Version {
	return gn.Version{
		Version: Version,
		Build:   Build,
	}
}

func (g gnames) Verify(params vlib.VerifyParams) ([]*vlib.Verification, error) {
	log.Printf("Verifying %d name-strings.", len(params.NameStrings))
	res := make([]*vlib.Verification, len(params.NameStrings))

	matches := g.matcher.MatchNames(params.NameStrings)

	var errString string
	matchRecords, err := g.vf.MatchRecords(matches)
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
