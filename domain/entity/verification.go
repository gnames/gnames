package entity

//Version is the output for Version method.
type Version struct {
	// Version of gnames project.
	Version string `json:"version"`
	// Build is the date and time of the project compilation.
	Build string `json:"build"`
}

// VerifyParams are options/parameters for the Verify method.
type VerifyParams struct {
	// NameStrings is a list of name-strings to verify.
	NameStrings []string `json:"nameStrings"`
	// PreferredSources contain DataSources IDs whos matches will be returned
	// becides the best result. See PreferredResults field in Verirication.
	PreferredSources []int `json:"preferredSources"`
	// WithVernaculars indicates if corresponding vernacular results will be
	// returned as well.
	WithVernaculars bool `json:"withVernaculars"`
}

// Verification is a result returned by Verify method.
type Verification struct {
	// InputID is a UUIDv5 generated out of the Input string.
	InputID string `json:"input_id"`
	// Input is a verified name-string
	Input string `json:"input"`
	// MatchType is best available match.
	MatchType
	// BestResult is the best result according to GNames scoring.
	BestResult *ResultData `json:"bestResult,omitempty"`

	// PreferredResults contain all detected matches from preverred data sources
	// provided by user.
	PreferredResults []*ResultData `json:"preferredResults,omitempty"`

	// DataSourcesNum is a number of data sources that matched an
	// input name-string.
	DataSourcesNum int `json:"dataSourcesNum"`

	// CurationLevel estimates reliability of matched data sources. If
	// matches are returned by at least one manually curated data source, or by
	// automatically curated data source, or only by sources that are not
	// significantly manually curated.
	CurationLevel `json:"-"`

	// CurationLevelString is a string representation of CurationLevel
	CurationLevelString string `json:"curationLevel"`

	// Error provides an error message, if any. If error is not empty, the match
	// failed because of a bug in the service.
	Error string `json:"error"`
}

// ResultData are returned data of the "best" or "preferred" result of
// name verification.
type ResultData struct {
	// DataSourceID is the ID of a matched DataSource.
	DataSourceID int `json:"dataSourceId"`

	// Shortened/abbreviated title of the data source.
	DataSrouceTitleShort string `json:"dataSourceTitleShort"`

	// CurationLevel of the data source.
	CurationLevel `json:"-"`

	// Curation is a string representation of CurationLevel
	CurationLevelString string `json:"curationLevel"`

	// ID from a data source. We try our best to return ID that correspond to
	// global_id of a DataSource. If such ID is not provided, this ID will be
	// auto-generated.  Auto-generated IDs will have 'gn_' prefix.
	ID string `json:"id"`

	// EntryDate is a timestamp created on entry of the data.
	EntryDate string `json:"entryDate"`

	// Score determines how well the match did work. It is used to determine
	// best match overall, and best match for every data-source.
	Score uint32 `json:"-"`

	// GlobalID that is exposed globally by a DataSource.
	GlobalID string `json:"globalId,omitempty"`

	// LocalID used by a DataSource internally.
	LocalID string `json:"localId,omitempty"`

	// Outlink to the record in the DataSource.
	Outlink string `json:"outlink,omitempty"`

	// MatchedName is a name-string from the DataSource that was matched
	// by GNames algorithm.
	MatchedName string `json:"matchedName"`

	// MatchCardinality is the cardinality of returned name:
	// 0 - No match, virus or hybrid formula,
	// 1 - Uninomial, 2 - Binomial, 3 - trinomial etc.
	MatchedCardinality int `json:"matchedCardinality"`

	// MatchedCanonicalSimple is a simplified canonicl form without ranks for
	// names lower than species, and with ommited hybrid signs for named hybrids.
	// Quite often simple canonical is the same as full canonical. Hybrid signs
	// are preserved for hybrid formulas.
	MatchedCanonicalSimple string `json:"matchedCanonicalSimple,omitempty"`

	// MatchedCanonicalFull is a canonical form that preserves hybrid signs
	// and infraspecific ranks.
	MatchedCanonicalFull string `json:"matchedCanonicalFull,omitempty"`

	// MatchedAuthors is a list of authors mentioned in the name.
	MatchedAuthors []string `json:"-"`

	// MatchedYear is a year mentioned in the name. Multiple years or
	// approximate years are ignored.
	MatchedYear int `json:"-"`

	// CurrentRecordID is the id of currently accepted name given by
	// the data-source.
	CurrentRecordID string `json:"currentRecordID"`

	// CurrentName is a currently accepted name (it is only provided by
	// DataSources with taxonomic data).
	CurrentName string `json:"currentName"`

	// CurrentCardinality is a cardinality of the accepted name.
	// It might differ from the matched name cardinality.
	CurrentCardinality int `json:"currentCardinality"`

	// CurrentCanonicalSimple is a canonical form for the currently accepted name.
	CurrentCanonicalSimple string `json:"currentCanonicalSimple"`

	// CurrentCanonicalFull is a full version of canonicall form for the
	// currently accepted name.
	CurrentCanonicalFull string `json:"currentCanonicalFull"`

	// IsSynonym is true if there is an indication in the DataSource that the
	// name is not a currently accepted name for one or another reason.
	IsSynonym bool `json:"isSynonym"`

	// ClassificationPath to the name (if provided by the DataSource).
	// Classification path consists of a hierarchy of name-strings.
	ClassificationPath string `json:"classificationPath,omitempty"`

	// ClassificationRanks of the classification path. They follow the
	// same order as the classification path.
	ClassificationRanks string `json:"classificationRanks,omitempty"`

	// ClassificationIDs of the names-strings. They always correspond to
	// the "id" field.
	ClassificationIDs string `json:"classificationIds,omitempty"`

	// EditDistance is a Levenshtein edit distance between canonical form of the
	// input name-string and the matched canonical form. If match type is
	// "EXACT", edit-distance will be 0.
	EditDistance int `json:"editDistance"`

	// StemEditDistance is a Levenshtein edit distance after removing suffixes
	// from specific epithets from canonical forms.
	StemEditDistance int `json:"stemEditDistance"`

	//MatchType describes what kind of a match happened to a name-string.
	MatchType `json:"matchType"`

	// Vernacular names that correspond to the matched name. (Will be implemented
	// later)
	Vernaculars []Vernacular `json:"vernaculars,omitempty"`
}

// Vernacular name
type Vernacular struct {
	Name string `json:"name"`
	// Language of the name, hopefully in ISO form.
	Language string `json:"language,omitempty"`
	// Locality is geographic places where the name is used.
	Locality string `json:"locality,omitempty"`
}
