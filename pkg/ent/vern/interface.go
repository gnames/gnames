package vern

import (
	vlib "github.com/gnames/gnlib/ent/verifier"
)

// Vernaculars provides functions required to add vernacular
// names to the results of verification. The results are
// returned back in the same sequence they were received.
type Vernaculars interface {
	AddVernacularNames(
		vernLangs []string,
		names []vlib.Name,
	) ([]vlib.Name, error)
}
