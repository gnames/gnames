package rest

import (
	"github.com/gnames/gnames"
	"github.com/gnames/gnfmt"
)

// VerifierService interface is the API behing RESTful service.
type VerifierService interface {
	// Ping checks if the service is alive.
	Ping() string

	// Port returns port of the HTTP/1 service.
	Port() int

	gnames.GNames

	gnfmt.Encoder
}
