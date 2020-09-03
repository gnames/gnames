package rest

import (
	"github.com/gnames/gnames"
	"github.com/gnames/gnames/encode"
	"github.com/gnames/gnames/model"
)

type VerifierHTTP struct {
	gn      gnames.GNames
	encoder encode.Encoder
}

// NewNewVerifierHTTP is a constructor for VerifiNewVerifierHTTP.
func NewVerifierHTTP(gn gnames.GNames, enc encode.Encoder) VerifierHTTP {
	return VerifierHTTP{gn: gn, encoder: enc}
}

// Ping checks if the service is alive.
func (v VerifierHTTP) Ping() string {
	return "pong"
}

// GetVersion returns Version of gnames project.
func (v VerifierHTTP) GetVersion() model.Version {
	return model.Version{
		Version: gnames.Version,
		Build:   gnames.Build,
	}
}

// Verify takes names-strings and options and returns verification result.
func (v VerifierHTTP) Verify(vp model.VerifyParams) []model.Verification {
	verif := make([]model.Verification, len(vp.NameStrings))
	return verif
}

// GetDataSources takes data-source id and opts and returns the data-source
// metadata.  If no id is provided, it returns metadata for all data-sources.
func (v VerifierHTTP) GetDataSources(opts model.DataSourcesOpts) []model.DataSource {
	var ds []model.DataSource
	return ds
}

// GetPort returns port of HTTP/1 service.
func (v VerifierHTTP) GetPort() int {
	return v.gn.Config.GNport
}

// Encode serializes an object.
func (v VerifierHTTP) Encode(input interface{}) ([]byte, error) {
	return v.encoder.Encode(input)
}

// Decode deserializes an object.
func (v VerifierHTTP) Decode(input []byte, output interface{}) error {
	return v.encoder.Decode(input, output)
}
