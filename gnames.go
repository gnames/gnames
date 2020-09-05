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
	Config      config.Config
	DataSources map[int]*model.DataSource
	DB          *sql.DB
}

func NewGNames(cnf config.Config) GNames {
	db := database.NewDB(cnf)
	gn := GNames{Config: cnf, DB: db}
	gn.DataSources = gn.getDataSourcesMap()
	return gn
}

func (gn GNames) Verify(params model.VerifyParams) []*model.Verification {
	log.Printf("Verifying %d name-strings.", len(params.NameStrings))
	res := make([]*model.Verification, len(params.NameStrings))

	matches, err := matcher.MatchNames(params.NameStrings, gn.Config.MatcherURL)
	errString := ""
	if err != nil {
		errString = err.Error()
	}

	matchRecords, err := database.MatchRecords(gn.DB, gn.DataSources, matches)
	if err != nil {
		errString = err.Error()
	}

	for i, v := range matches {
		mr := matchRecords[database.InputUUID(v.ID)]
		item := model.Verification{
			InputID:        v.ID,
			Input:          v.Name,
			MatchType:      mr.MatchType,
			CurationLevel:  mr.CurationLevel,
			DataSourcesNum: mr.DataSourcesNum,
			BestResult:     bestResult(mr.ResultData),
			Error:          errString,
		}

		res[i] = &item
	}
	for _, v := range res {
		log.Printf("BM: %+v", v.BestResult)
	}
	return res
}

func bestResult(rds []*model.ResultData) *model.ResultData {
	var rd *model.ResultData
	if len(rds) > 0 {
		rd = rds[0]
	}
	return rd
}

func (gn GNames) GetDataSources(opts model.DataSourcesOpts) ([]*model.DataSource, error) {
	if opts.DataSourceID > 0 {
		log.Printf("Getting data source with ID %d.", opts.DataSourceID)
		return database.GetDataSource(gn.DB, opts.DataSourceID)
	}
	log.Println("Getting all data sources.")
	return database.GetDataSources(gn.DB)
}

func (gn GNames) getDataSourcesMap() map[int]*model.DataSource {
	res := make(map[int]*model.DataSource)
	dss, err := gn.GetDataSources(model.DataSourcesOpts{})
	if err != nil {
		log.Fatalf("Cannot init DataSources data: %s", err)
	}
	for _, ds := range dss {
		res[ds.ID] = ds
	}
	return res
}
