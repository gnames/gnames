package matcher

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/gnames/gnlib/domain/entity/gn"
	mlib "github.com/gnames/gnlib/domain/entity/matcher"
	"github.com/gnames/gnlib/encode"
	log "github.com/sirupsen/logrus"
)

type matcherREST struct {
	url string
	enc encode.Encoder
}

func NewGNMatcher(url string) matcherREST {
	return matcherREST{url: url, enc: encode.GNjson{}}
}

func (mr matcherREST) GetVersion() gn.Version {
	var err error
	response := gn.Version{}
	resp, err := http.Get(mr.url + "version")
	if err != nil {
		log.Warnf("Cannot get gnmatcher version: %s.", err)
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("Cannot get gnmatcher version: %s.", err)
	}
	err = mr.enc.Decode(respBytes, &response)
	if err != nil {
		log.Warnf("Cannot get gnmatcher version: %s.", err)
	}
	return response
}

func (mr matcherREST) MatchNames(names []string) []*mlib.Match {
	var response []*mlib.Match
	req, err := mr.enc.Encode(names)
	if err != nil {
		log.Warnf("Cannot encode name-strings: %s.", err)
	}
	r := bytes.NewReader(req)
	resp, err := http.Post(mr.url+"match", "application/json", r)
	if err != nil {
		log.Warnf("Cannot get matches response: %s.", err)
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("Cannot read matches from response: %s.", err)
	}
	err = mr.enc.Decode(respBytes, &response)
	if err != nil {
		log.Warnf("Cannot decode matches: %s.", err)
	}
	return response
}
