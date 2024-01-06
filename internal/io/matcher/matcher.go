package matcher

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"

	"github.com/gnames/gnfmt"
	"github.com/gnames/gnlib/ent/gnvers"
	mlib "github.com/gnames/gnlib/ent/matcher"
	gnmatcher "github.com/gnames/gnmatcher/pkg"
	gnmcfg "github.com/gnames/gnmatcher/pkg/config"
)

type matcherREST struct {
	url string
	enc gnfmt.Encoder
}

// New creates an implementation of GNmatcher interface.
func New(url string) gnmatcher.GNmatcher {
	return matcherREST{url: url, enc: gnfmt.GNjson{}}
}

func (mr matcherREST) GetVersion() gnvers.Version {
	var err error
	response := gnvers.Version{}
	resp, err := http.Get(mr.url + "version")
	if err != nil {
		slog.Error("Cannot get gnmatcher version", "error", err)
	}
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Cannot get gnmatcher version", "error", err)
	}
	err = mr.enc.Decode(respBytes, &response)
	if err != nil {
		slog.Error("Cannot get gnmatcher version", "error", err)
	}
	return response
}

func (mr matcherREST) MatchNames(
	names []string,
	opts ...gnmcfg.Option,
) mlib.Output {
	var response mlib.Output
	cfg := gnmcfg.New(opts...)
	req, err := mr.enc.Encode(mlib.Input{
		Names:                   names,
		WithSpeciesGroup:        cfg.WithSpeciesGroup,
		WithUninomialFuzzyMatch: cfg.WithUninomialFuzzyMatch,
		DataSources:             cfg.DataSources,
	})
	if err != nil {
		slog.Error("Cannot encode name-strings", "error", err)
	}
	r := bytes.NewReader(req)
	resp, err := http.Post(mr.url+"matches", "application/json", r)
	if err != nil {
		slog.Error("Cannot get matches response", "error", err)
	}
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Cannot read matches from response", "error", err)
	}
	err = mr.enc.Decode(respBytes, &response)
	if err != nil {
		slog.Error("Cannot decode matches", "error", err)
	}
	return response
}

// GetConfig is a placeholder
func (mr matcherREST) GetConfig() gnmcfg.Config {
	return gnmcfg.New()
}
