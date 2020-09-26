package gnuuid

import (
	u "github.com/google/uuid"
)

// Domain is UUIDv5 seed for `globalnames.org` domain identifiers. It was
// generated originally with
//
// u.NewSHA1(u.NameSpaceDNS, []byte("globalnames.org"))
var (
	GNDomain = u.Must(u.Parse("90181196-fecf-5082-a4c1-411d4f314cda"))
	Nil      = u.Nil
)

// Creates new UUIDv5 identifier for globalnames
func New(name string) u.UUID {
	return u.NewSHA1(GNDomain, []byte(name))
}
