package gnames

import (
	"github.com/gnames/gnames/config"
	"github.com/gnames/gnames/data"
	"github.com/gnames/gnames/domain/entity"
	"github.com/gnames/gnames/matcher"
	"github.com/gnames/gnames/score"
	gnmu "github.com/gnames/gnmatcher/domain/usecase"
	log "github.com/sirupsen/logrus"
)

type GNames struct {
	Config config.Config
	data.DataGrabber
	gnmu.Matcher
}

func NewGNames(cnf config.Config, dg data.DataGrabber) GNames {
	return GNames{
		Config:      cnf,
		DataGrabber: dg,
		Matcher:     matcher.NewMatcherREST(cnf.MatcherURL),
	}
}

func (gn GNames) Verify(params entity.VerifyParams) ([]*entity.Verification, error) {
	log.Printf("Verifying %d name-strings.", len(params.NameStrings))
	res := make([]*entity.Verification, len(params.NameStrings))

	matches := gn.Matcher.MatchAry(params.NameStrings)

	var errString string
	matchRecords, err := gn.DataGrabber.MatchRecords(matches)
	if err != nil {
		errString = err.Error()
	}

	for i, v := range matches {
		if mr, ok := matchRecords[v.ID]; ok {
			score.Calculate(mr)
			item := entity.Verification{
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
			log.Warnf("Cannot find %+v record.", v)
		}
	}
	return res, nil
}

func (gn GNames) DataSources(opts entity.DataSourcesOpts) ([]*entity.DataSource, error) {
	log.Printf("Getting data source with ID %d.", opts.DataSourceID)
	dsID := opts.DataSourceID
	nullDsID := data.NullInt{Int: dsID, Valid: dsID > 0}
	return gn.DataGrabber.DataSources(nullDsID)
}
