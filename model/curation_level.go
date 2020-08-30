package model

// CurationLevel tells if matched result was returned by at least one
// DataSource in the following categories.
type CurationLevel int

const (
	// NotCurated means that all DataSources where the name-string was matched
	// are not curated sufficiently.
	NotCurated CurationLevel = iota

	// Curated means that at least one DataSource is marked as sufficiently
	// curated. It does not mean that the particular match was manually checked
	// though.
	Curated

	// AutoCurated means that at least one of the returned DataSources invested
	// significantly in curating their data by scripts.
	AutoCurated
)

var mapCurationLevel = map[int]string{
	0: "NOT_CURATED",
	1: "CURATED",
	2: "AUTO_CURATED",
}

func (c CurationLevel) String() string {
	if match, ok := mapCurationLevel[int(c)]; ok {
		return match
	} else {
		return "N/A"
	}
}

