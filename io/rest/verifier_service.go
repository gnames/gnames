package rest

import (
	"github.com/gnames/gnames"
	"github.com/gnames/gnlib/domain/entity/gn"
	vlib "github.com/gnames/gnlib/domain/entity/verifier"
	"github.com/gnames/gnlib/encode"
)

type verifierService struct {
	gnames  gnames.GNames
	port    int
	encoder encode.Encoder
}

// NewVerifierService is a constructor for the implementation of the VerifierService interface.
func NewVerifierService(g gnames.GNames, port int, enc encode.Encoder) VerifierService {
	return verifierService{gnames: g, port: port, encoder: enc}
}

// Ping checks if the service is alive.
func (vs verifierService) Ping() string {
	return "pong"
}

// Version returns Version of gnames project.
func (vs verifierService) GetVersion() gn.Version {
	return vs.gnames.GetVersion()
}

// Verify takes names-strings and options and returns verification result.
func (vs verifierService) Verify(params vlib.VerifyParams) ([]*vlib.Verification, error) {
	return vs.gnames.Verify(params)
}

// DataSources takes data-source id and opts and returns the data-source
// metadata.  If no id is provided, it returns metadata for all data-sources.
func (vs verifierService) DataSources(ids ...int) ([]*vlib.DataSource, error) {
	return vs.gnames.DataSources(ids...)
}

// Port returns port of HTTP/1 service.
func (vs verifierService) Port() int {
	return vs.port
}

// Encode serializes an object.
func (vs verifierService) Encode(input interface{}) ([]byte, error) {
	return vs.encoder.Encode(input)
}

// Decode deserializes an object.
func (vs verifierService) Decode(input []byte, output interface{}) error {
	return vs.encoder.Decode(input, output)
}
