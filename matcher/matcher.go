package matcher

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/gnames/gnames/encode"
	gnm "github.com/gnames/gnmatcher/model"
)

func MatchNames(names []string, url string) ([]*gnm.Match, error) {
	var response []*gnm.Match
	enc := encode.GNgob{}
	req, err := enc.Encode(names)
	if err != nil {
		return response, err
	}
	r := bytes.NewReader(req)
	resp, err := http.Post(url+"match", "application/x-binary", r)
	if err != nil {
		return response, err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response, err
	}
	err = enc.Decode(respBytes, &response)
	if err != nil {
		return response, err
	}
	return response, nil
}
