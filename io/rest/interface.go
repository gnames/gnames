package rest

import (
	"github.com/gnames/gnames"
	"github.com/gnames/gnlib/domain/entity/gn"
	"github.com/gnames/gnlib/encode"
)

// VerifierService interface is the API behing RESTful service.
type VerifierService interface {
	// Ping checks if the service is alive.
	Ping() string

	// Versioner returns Version of gnames project.
	gn.Versioner

	// Port returns port of the HTTP/1 service.
	Port() int

	gnames.GNames

	encode.Encoder
}
