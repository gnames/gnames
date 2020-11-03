package rest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	vlib "github.com/gnames/gnlib/domain/entity/verifier"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// Run starts HTTP/1 service for scientific names verification.
func Run(s VerificationService) {
	log.Printf("Starting the HTTP API server on port %d.", s.Port())
	r := mux.NewRouter()

	r.HandleFunc("/ping",
		func(resp http.ResponseWriter, req *http.Request) {
			pingHTTP(resp, req, s)
		})

	r.HandleFunc("/version",
		func(resp http.ResponseWriter, req *http.Request) {
			getVersionHTTP(resp, req, s)
		})

	r.HandleFunc("/verification",
		func(resp http.ResponseWriter, req *http.Request) {
			verifyHTTP(resp, req, s)
		})

	r.HandleFunc("/data_sources/{id:[0-9]+}",
		func(resp http.ResponseWriter, req *http.Request) {
			vars := mux.Vars(req)
			id, err := strconv.Atoi(vars["id"])
			if err != nil {
				log.Warnf("Cannot convert DataSourceID %s: %s.", vars["id"], err)
			}
			getDataSourcesHTTP(resp, req, s, vlib.DataSourcesOpts{DataSourceID: id})
		}).Methods("GET", "POST")

	r.HandleFunc("/data_sources",
		func(resp http.ResponseWriter, req *http.Request) {
			getDataSourcesHTTP(resp, req, s, vlib.DataSourcesOpts{})
		}).Methods("GET", "POST")

	addr := fmt.Sprintf(":%d", s.Port())

	server := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 300 * time.Second,
		ReadTimeout:  300 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}

func pingHTTP(resp http.ResponseWriter, _ *http.Request,
	s VerificationService) {
	resp.Write([]byte(s.Ping()))
}

func getVersionHTTP(resp http.ResponseWriter, _ *http.Request,
	s VerificationService) {
	version := s.GetVersion()
	ver, err := s.Encode(version)
	if err != nil {
		log.Warnf("Cannot decode version: %s.", err)
	}
	resp.Write([]byte(ver))
}

func verifyHTTP(resp http.ResponseWriter, req *http.Request,
	s VerificationService) {
	var params vlib.VerifyParams
	var body []byte
	var err error

	if body, err = ioutil.ReadAll(req.Body); err != nil {
		log.Warnf("verifyHTTP: cannot read message from request : %v.", err)
		return
	}

	if err = s.Decode(body, &params); err != nil {
		log.Warnf("verifyHTTP: cannot decode message from request : %v.", err)
		return
	}

	verified := s.Verify(params)

	if out, err := s.Encode(verified); err == nil {
		resp.Write(out)
	} else {
		log.Warnf("MatchAry: Cannot encode response : %v.", err)
	}
}

func getDataSourcesHTTP(resp http.ResponseWriter, req *http.Request,
	s VerificationService, opts vlib.DataSourcesOpts) {
	dataSources := s.DataSources(opts)

	if out, err := s.Encode(dataSources); err == nil {
		resp.Write(out)
	} else {
		log.Warnf("MatchAry: Cannot encode response : %v.", err)
	}
}
