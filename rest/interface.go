package rest

import (
	"github.com/gnames/gnames/domain/entity"
	"github.com/gnames/gnames/domain/usecase"
	"github.com/gnames/gnlib/encode"
)

type VerificationService interface {
	// Ping checks if the service is alive.
	Ping() string

	// Version returns Version of gnames project.
	Version() entity.Version

	// Port returns port of the HTTP/1 service.
	Port() int

	usecase.Verifier

	encode.Encoder
}
