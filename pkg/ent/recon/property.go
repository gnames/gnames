package recon

import "github.com/gnames/gnlib/ent/reconciler"

type GnamesProperty int

const (
	Unknown GnamesProperty = iota
	HigherTaxon
	DataSourceIDs
	CurrentName
	Classification
	DataSource
	OutlinkURL
)

func NewProp(id string) GnamesProperty {
	switch id {
	case "higher_taxon":
		return HigherTaxon
	case "data_source_ids":
		return DataSourceIDs
	case "current_name":
		return CurrentName
	case "classification":
		return Classification
	case "data_source":
		return DataSource
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
	case CurrentName:
		id = "current_name"
		name = "CurrentName"
	case Classification:
		id = "classification"
		name = "Classification"
	case DataSource:
		id = "data_source"
		name = "DataSource"
	case OutlinkURL:
		id = "outlink_url"
		name = "OutlinkURL"
	default:
		id = "unknown"
		name = "UnknownProperty"
	}
	return reconciler.Property{ID: id, Name: name}
}
