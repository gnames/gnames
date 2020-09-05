package gnames

import (
	"database/sql"

	"github.com/gnames/gnames/config"
	"github.com/gnames/gnames/database"
	"github.com/gnames/gnames/matcher"
	"github.com/gnames/gnames/model"
	log "github.com/sirupsen/logrus"
)

type GNames struct {
	Config config.Config
	DB     *sql.DB
}

func NewGNames(cnf config.Config) GNames {
	db := database.NewDB(cnf)
	return GNames{Config: cnf, DB: db}
}

func (gn GNames) Verify(params model.VerifyParams) []*model.Verification {
	log.Printf("Verifying %d name-strings.", len(params.NameStrings))
	res := make([]*model.Verification, len(params.NameStrings))
	matches, err := matcher.MatchNames(params.NameStrings, gn.Config.MatcherURL)
	errString := ""
	if err != nil {
		errString = err.Error()
	}
	for i, v := range matches {
		item := model.Verification{
			InputID: v.ID,
			Input:   v.Name,
			Error:   errString,
		}
		res[i] = &item
	}
	return res
}

func (gn GNames) GetDataSources(opts model.DataSourcesOpts) ([]*model.DataSource, error) {
	if opts.DataSourceID > 0 {
		log.Printf("Getting data source with ID %d.", opts.DataSourceID)
		return database.GetDataSource(gn.DB, opts.DataSourceID)
	}
	log.Println("Getting all data sources.")
	return database.GetDataSources(gn.DB)
}
