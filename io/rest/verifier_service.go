package rest

import (
	"github.com/gnames/gnames"
	"github.com/gnames/gnfmt"
)

type verifierService struct {
	gnames.GNames
	gnfmt.Encoder
	port int
}

// NewVerifierService is a constructor for the implementation of the VerifierService interface.
func NewVerifierService(g gnames.GNames, port int, enc gnfmt.Encoder) VerifierService {
	return verifierService{GNames: g, port: port, Encoder: enc}
}

// Ping checks if the service is alive.
func (vs verifierService) Ping() string {
	return "pong"
}

// Port returns port of HTTP/1 service.
func (vs verifierService) Port() int {
	return vs.port
}
