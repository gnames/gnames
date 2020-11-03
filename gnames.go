package gnames

import (
	"github.com/gnames/gnames/config"
	"github.com/gnames/gnames/data"
	"github.com/gnames/gnames/matcher"
	"github.com/gnames/gnames/score"
	mlib "github.com/gnames/gnlib/domain/entity/matcher"
	vlib "github.com/gnames/gnlib/domain/entity/verifier"
	log "github.com/sirupsen/logrus"
)

type GNames struct {
	Config config.Config
	data.DataGrabber
	mlib.Matcher
}

func NewGNames(cnf config.Config, dg data.DataGrabber) GNames {
	return GNames{
		Config:      cnf,
		DataGrabber: dg,
		Matcher:     matcher.NewMatcherREST(cnf.MatcherURL),
	}
}

func (gn GNames) Verify(params vlib.VerifyParams) ([]*vlib.Verification, error) {
	log.Printf("Verifying %d name-strings.", len(params.NameStrings))
	res := make([]*vlib.Verification, len(params.NameStrings))

	matches := gn.Matcher.MatchAry(params.NameStrings)

	var errString string
	matchRecords, err := gn.DataGrabber.MatchRecords(matches)
	if err != nil {
		errString = err.Error()
	}

	for i, v := range matches {
		if mr, ok := matchRecords[v.ID]; ok {
			score.Calculate(mr)
			item := vlib.Verification{
				InputID:             mr.InputID,
				Input:               mr.Input,
				MatchType:           mr.MatchType,
				CurationLevel:       mr.CurationLevel,
				CurationLevelString: mr.CurationLevel.String(),
				DataSourcesNum:      mr.DataSourcesNum,
				BestResult:          score.BestResult(mr),
				PreferredResults:    score.PreferredResults(params.PreferredSources, mr),
				Error:               errString,
			}

			res[i] = &item
		} else {
			log.Warnf("Cannot find record for '%s'.", v.Name)
		}
	}
	return res, nil
}

func (gn GNames) DataSources(opts vlib.DataSourcesOpts) ([]*vlib.DataSource, error) {
	log.Printf("Getting data source with ID %d.", opts.DataSourceID)
	dsID := opts.DataSourceID
	nullDsID := data.NullInt{Int: dsID, Valid: dsID > 0}
	return gn.DataGrabber.DataSources(nullDsID)
}
