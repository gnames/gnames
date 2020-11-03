package rest

import (
	"github.com/gnames/gnames"
	"github.com/gnames/gnlib/domain/entity/gn"
	vlib "github.com/gnames/gnlib/domain/entity/verifier"
	"github.com/gnames/gnlib/encode"
	log "github.com/sirupsen/logrus"
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

// Version returns Version of gnames project.
func (v VerifierHTTP) GetVersion() gn.Version {
	return gn.Version{
		Version: gnames.Version,
		Build:   gnames.Build,
	}
}

// Verify takes names-strings and options and returns verification result.
func (v VerifierHTTP) Verify(params vlib.VerifyParams) []*vlib.Verification {
	verif, err := v.gn.Verify(params)
	if err != nil {
		log.Warnf("Cannot verify names: %s", err)
	}
	return verif
}

// DataSources takes data-source id and opts and returns the data-source
// metadata.  If no id is provided, it returns metadata for all data-sources.
func (v VerifierHTTP) DataSources(opts vlib.DataSourcesOpts) []*vlib.DataSource {
	ds, err := v.gn.DataSources(opts)
	if err != nil {
		log.Warnf("VerifierHTTP cannot get data_sources: %s.", err)
	}
	return ds
}

// Port returns port of HTTP/1 service.
func (v VerifierHTTP) Port() int {
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
