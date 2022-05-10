package matcher

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/gnames/gnfmt"
	"github.com/gnames/gnlib/ent/gnvers"
	mlib "github.com/gnames/gnlib/ent/matcher"
	"github.com/gnames/gnmatcher"
	gnmcfg "github.com/gnames/gnmatcher/config"
	"github.com/rs/zerolog/log"
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
		log.Warn().Err(err).Msg("Cannot get gnmatcher version")
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot get gnmatcher version")
	}
	err = mr.enc.Decode(respBytes, &response)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot get gnmatcher version")
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
		Names:            names,
		WithSpeciesGroup: cfg.WithSpeciesGroup,
		DataSources:      cfg.DataSources,
	})
	if err != nil {
		log.Warn().Err(err).Msg("Cannot encode name-strings")
	}
	r := bytes.NewReader(req)
	resp, err := http.Post(mr.url+"matches", "application/json", r)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot get matches response")
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot read matches from response")
	}
	err = mr.enc.Decode(respBytes, &response)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot decode matches")
	}
	return response
}

// GetConfig is a placeholder
func (mr matcherREST) GetConfig() gnmcfg.Config {
	return gnmcfg.New()
}
