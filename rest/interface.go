package rest

import (
	"github.com/gnames/gnlib/domain/entity/gn"
	vlib "github.com/gnames/gnlib/domain/entity/verifier"
	"github.com/gnames/gnlib/encode"
)

type VerificationService interface {
	// Ping checks if the service is alive.
	Ping() string

	// Versioner returns Version of gnames project.
	gn.Versioner

	// Port returns port of the HTTP/1 service.
	Port() int

	vlib.Verifier

	encode.Encoder
}
