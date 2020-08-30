package model

type GNamesService interface {
	// Ping checks if the service is alive.
	Ping() string

	// GetVersion returns Version of gnames project.
	GetVersion() Version

	// Verify takes names-strings and options and returns verification result.
	Verify(VerifyOpts) []Verification

	// GetDataSources takes data-source id and opts and returns the data-source
	// metadata.  If no id is provided, it returns metadata for all data-sources.
	GetDataSources(DataSourcesOpts) []DataSource
}
