package rest

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gnames/gnames/model"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// Run starts HTTP/1 service for scientific names verification.
func Run(s model.VerificationService) {
	log.Printf("Starting the HTTP API server on port %d.", s.GetPort())
	r := mux.NewRouter()

	r.HandleFunc("/ping",
		func(resp http.ResponseWriter, req *http.Request) {
			pingHTTP(resp, req, s)
		})

	addr := fmt.Sprintf(":%d", s.GetPort())

	server := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 300 * time.Second,
		ReadTimeout:  300 * time.Second,
	}

	log.Fatal(server.ListenAndServe())
}

func pingHTTP(resp http.ResponseWriter, _ *http.Request,
	s model.VerificationService) {
	resp.Write([]byte(s.Ping()))
}

