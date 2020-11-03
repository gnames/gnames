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

type MatcherREST struct {
	URL string
	Enc encode.Encoder
}

func NewMatcherREST(url string) MatcherREST {
	return MatcherREST{URL: url, Enc: encode.GNgob{}}
}

func (mr MatcherREST) GetVersion() gn.Version {
	var err error
	response := gn.Version{}
	var req []byte
	r := bytes.NewReader(req)
	resp, err := http.Post(mr.URL+"version", "application/x-binary", r)
	if err != nil {
		log.Warnf("Cannot get gnmatcher version: %s.", err)
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("Cannot get gnmatcher version: %s.", err)
	}
	err = mr.Enc.Decode(respBytes, &response)
	if err != nil {
		log.Warnf("Cannot get gnmatcher version: %s.", err)
	}
	return response
}

func (mr MatcherREST) MatchAry(names []string) []*mlib.Match {
	var response []*mlib.Match
	enc := encode.GNgob{}
	req, err := enc.Encode(names)
	if err != nil {
		log.Warnf("Cannot encode name-strings: %s.", err)
	}
	r := bytes.NewReader(req)
	resp, err := http.Post(mr.URL+"match", "application/x-binary", r)
	if err != nil {
		log.Warnf("Cannot get matches response: %s.", err)
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Warnf("Cannot read matches from response: %s.", err)
	}
	err = enc.Decode(respBytes, &response)
	if err != nil {
		log.Warnf("Cannot decode matches: %s.", err)
	}
	return response
}
