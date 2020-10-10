package usecase

import (
	"github.com/gnames/gnames/domain/entity"
	"github.com/gnames/gnames/lib/format"
)

type Verifier interface {
	// Verify takes names-strings and options and returns verification result.
	Verify(entity.VerifyParams) []*entity.Verification

	// DataSources takes data-source id and opts and returns the data-source
	// metadata.  If no id is provided, it returns metadata for all data-sources.
	DataSources(entity.DataSourcesOpts) []*entity.DataSource
}

// Outputter interface us a uniform way to create an output of a datum
type Outputter interface {
	// FormattedOutput takes a record and returns a string representation of
	// the record accourding to supplied format.
	Output(record interface{}, f format.Format) string
}
