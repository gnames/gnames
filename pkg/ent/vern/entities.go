package vern

// Record contains data required for finding vernacular names for a
// particular taxon.
type Record struct {
	// DataSourceID contains ID for the required DataSource.
	DataSourceID int
	// RecordID contains ID of the required taxon record.
	RecordID string

	// CurrentRecordID contains ID of the current taxon record.
	CurrentRecordID string
}
