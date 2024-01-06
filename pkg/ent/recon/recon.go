package recon

import "github.com/gnames/gnlib/ent/reconciler"

// GnamesProperty is used for extending the reconciliation process.
// These properties can be obtained for matched reconciliation results and
// add additional information for matches.
//
// Some properties are used as filters to add constrains to the reconciliation.
type GnamesProperty int

const (
	// Unknown is a placeholder "nil" property.
	Unknown GnamesProperty = iota

	// HigherTaxon is a filter property. It restricts reconciliation results
	// to ones that include given HigherTaxon, for example `Mollusca` or
	// `Plantae`.
	HigherTaxon

	// DataSourceIDs is a filter property. It restricts reconciliation results
	// to the given data-sources. The data-sources are expressed through
	// their IDs.
	DataSourceIDs

	// CanonicalForm provides the canonical form of matched name.
	CanonicalForm

	// CurrenName provides currently accepted name of the matched taxon
	// according to the data-source where match happened.
	CurrentName

	// Classification provides a classification path to matched taxon according
	// to the corresponding data-source.
	Classification

	// DataSource provides information about the matched data-source.
	DataSource

	// AllDataSources provides information for all data-sources where match
	// happened.
	AllDataSources

	// OutlinkURL provides a URL to the matched name or taxon at the site of
	// the data-source.
	OutlinkURL
)

// NewProp create a new GnamesProperty out of given string. If string cannot
// be matched to any known properties, the `Unknown` property is returned.
func NewProp(id string) GnamesProperty {
	switch id {
	case "higher_taxon":
		return HigherTaxon
	case "data_source_ids":
		return DataSourceIDs
	case "canonical_form":
		return CanonicalForm
	case "current_name":
		return CurrentName
	case "classification":
		return Classification
	case "data_source":
		return DataSource
	case "all_data_sources":
		return AllDataSources
	case "outlink_url":
		return OutlinkURL

	default:
		return Unknown
	}
}

func (p GnamesProperty) Property() reconciler.Property {
	var id, name string
	switch p {
	case HigherTaxon:
		id = "higher_taxon"
		name = "HigherTaxon"
	case DataSourceIDs:
		id = "data_source_ids"
		name = "DataSourceIds"
	case CanonicalForm:
		id = "canonical_form"
		name = "CanonicalForm"
	case CurrentName:
		id = "current_name"
		name = "CurrentName"
	case Classification:
		id = "classification"
		name = "Classification"
	case DataSource:
		id = "data_source"
		name = "DataSource"
	case AllDataSources:
		id = "all_data_sources"
		name = "AllDataSources"
	case OutlinkURL:
		id = "outlink_url"
		name = "OutlinkURL"
	default:
		id = "unknown"
		name = "UnknownProperty"
	}
	return reconciler.Property{ID: id, Name: name}
}
