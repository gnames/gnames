package rest

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	gnames "github.com/gnames/gnames/pkg"
	"github.com/gnames/gnames/pkg/ent/recon"
	"github.com/gnames/gnfmt"
	"github.com/gnames/gnlib/ent/reconciler"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnquery"
	"github.com/gnames/gnquery/ent/search"
	"github.com/gnames/gnuuid"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	nsqcfg "github.com/sfgrp/lognsq/config"
	"github.com/sfgrp/lognsq/ent/nsq"
	"github.com/sfgrp/lognsq/io/nsqio"
)

var (
	apiPath       = "/api/v1/"
	reconcileType = "ScientificNameString"
	reconcileID   = "name_string"
)

// Run starts HTTP/1 service on given port for scientific names verification.
func Run(gn gnames.GNames, port int) {
	log.Info().Int("port", port).Msg("Starting HTTP API server")
	e := echo.New()
	e.Use(middleware.Gzip())
	e.Use(middleware.CORS())

	loggerNSQ := setLogger(e, gn)
	if loggerNSQ != nil {
		defer loggerNSQ.Stop()
	}

	e.GET("/", info)
	e.GET("/api", info)
	e.GET("/api/", info)
	e.GET("/api/v1", info)
	e.GET("/api/v1/", info)
	e.GET(apiPath, info)
	e.GET(apiPath+"ping", ping)
	e.GET(apiPath+"version", ver(gn))
	e.GET(apiPath+"data_sources", dataSources(gn))
	e.GET(apiPath+"data_sources/:id", oneDataSource(gn))
	e.GET(apiPath+"name_strings", nameInfoGET(gn))
	e.GET(apiPath+"name_strings/", nameInfoGET(gn))
	e.GET(apiPath+"name_strings/:id", nameGET(gn))
	// same as verify, kept for backward compatibility
	e.POST(apiPath+"verifications", verificationPOST(gn))
	// same as verify, kept for backward compatibility
	e.GET(apiPath+"verifications/:names", verificationGET(gn))
	e.POST(apiPath+"verify", verificationPOST(gn))
	e.GET(apiPath+"verify/:names", verificationGET(gn))
	e.POST(apiPath+"search", searchPOST(gn))
	e.GET(apiPath+"search/:query", searchGET(gn))
	e.GET(apiPath+"reconcile", reconcileGET(gn))
	e.POST(apiPath+"reconcile", reconcilePOST(gn))
	e.GET(apiPath+"reconcile/properties", propertiesGET(gn))

	addr := fmt.Sprintf(":%d", port)
	s := &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
	}
	e.Logger.Fatal(e.StartServer(s))
}

func info(c echo.Context) error {
	return c.String(http.StatusOK,
		`The API is described at
https://apidoc.globalnames.org/gnames`)
}

func ping(c echo.Context) error {
	return c.String(http.StatusOK, "pong")
}

func ver(gn gnames.GNames) func(echo.Context) error {
	return func(c echo.Context) error {
		result := gn.GetVersion()
		return c.JSON(http.StatusOK, result)
	}
}

func dataSources(gn gnames.GNames) func(echo.Context) error {
	return func(c echo.Context) error {
		dataSources, err := gn.DataSources()
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, dataSources)
	}
}

func oneDataSource(gn gnames.GNames) func(echo.Context) error {
	return func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return err
		}
		dataSources, err := gn.DataSources(id)
		if err != nil {
			return err
		}
		if len(dataSources) == 0 {
			return fmt.Errorf("cannot find DataSource for id '%s'", idStr)
		}
		return c.JSON(http.StatusOK, dataSources[0])
	}
}

func reconcileGET(gn gnames.GNames) func(echo.Context) error {
	return func(c echo.Context) error {
		enc := gnfmt.GNjson{}
		if c.QueryParam("extend") != "" {
			res, err := extend(gn, c.QueryParam("extend"))
			if err != nil {
				return err
			}
			return c.JSON(http.StatusOK, res)
		}
		if c.QueryParam("queries") == "" {
			return manifest(c, gn)
		}
		var params map[string]reconciler.Query
		q, err := url.QueryUnescape(c.QueryParam("queries"))
		if err != nil {
			return err
		}
		err = enc.Decode([]byte(q), &params)
		if err != nil {
			return err
		}
		res, err := reconcile(gn, params)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, res)
	}
}

func extend(gn gnames.GNames, q string) (reconciler.ExtendOutput, error) {
	var res reconciler.ExtendOutput
	enc := gnfmt.GNjson{}
	var params reconciler.ExtendQuery
	err := enc.Decode([]byte(q), &params)
	if err != nil {
		return res, err
	}
	return gn.ExtendReconcile(params)
}

func reconcilePOST(gn gnames.GNames) func(echo.Context) error {
	return func(c echo.Context) error {
		ctx, cancel := getContext(c)
		defer cancel()
		chErr := make(chan error)

		go func() {
			defer close(chErr)
			enc := gnfmt.GNjson{}

			var err error
			ext := []byte(c.FormValue("extend"))
			if len(ext) > 0 {
				var params reconciler.ExtendQuery
				var extRes reconciler.ExtendOutput
				err := enc.Decode([]byte(ext), &params)
				if err == nil {
					extRes, err = gn.ExtendReconcile(params)
				}
				if err == nil {
					err = c.JSON(http.StatusOK, extRes)
				}
				chErr <- err
			}

			var params map[string]reconciler.Query
			var res reconciler.Output
			q := []byte(c.FormValue("queries"))

			err = enc.Decode(q, &params)
			if err == nil {
				res, err = reconcile(gn, params)
			}

			if err == nil {
				err = c.JSON(http.StatusOK, res)
			}

			chErr <- err
		}()

		select {
		case <-ctx.Done():
			<-chErr
			return ctx.Err()
		case err := <-chErr:
			return err
		case <-time.After(6 * time.Minute):
			return errors.New("request took too long")
		}
	}
}

func reconcile(
	gn gnames.GNames,
	params map[string]reconciler.Query,
) (reconciler.Output, error) {
	var err error
	var res reconciler.Output
	var verified vlib.Output
	var names, ids []string
	for k, v := range params {
		ids = append(ids, k)
		names = append(names, v.Query)
	}
	inp := vlib.Input{
		NameStrings:    names,
		WithAllMatches: true,
	}
	verified, err = gn.Verify(context.Background(), inp)
	if err != nil {
		return res, err
	}
	res = gn.Reconcile(verified, params, ids)
	return res, nil
}

func propertiesGET(gn gnames.GNames) func(echo.Context) error {
	return func(c echo.Context) error {
		t, err := url.QueryUnescape(c.QueryParam("type"))
		if err != nil {
			return err
		}
		t = strings.TrimSpace(t)
		if t != reconcileID {
			t = reconcileID
		}
		res := properties(gn, t)

		return c.JSON(http.StatusOK, res)
	}
}

func properties(gn gnames.GNames, typ string) reconciler.PropertyOutput {
	return reconciler.PropertyOutput{
		Type: reconcileType,
		Properties: []reconciler.Property{
			recon.CanonicalForm.Property(),
			recon.CurrentName.Property(),
			recon.Classification.Property(),
			recon.DataSource.Property(),
			recon.AllDataSources.Property(),
			recon.OutlinkURL.Property(),
		},
	}
}

func manifest(c echo.Context, gn gnames.GNames) error {
	gnvURL := gn.GetConfig().WebPageURL
	gnamesURL := gn.GetConfig().GnamesHostURL
	types := []reconciler.Type{
		{
			ID:   reconcileID,
			Name: reconcileType,
		},
	}
	preview := reconciler.Preview{
		Width:  350,
		Height: 233,
		URL:    gnvURL + "/name_strings/widget/{{id}}",
	}

	view := reconciler.View{
		URL: gnvURL + "/name_strings/{{id}}?all_matches=true",
	}

	ext := reconciler.Extend{
		ProposeProperties: reconciler.ProposeProperties{
			ServiceURL:  gnamesURL,
			ServicePath: apiPath + "reconcile/properties",
		},
	}

	res := reconciler.Manifest{
		Versions:        []string{"0.2"},
		Name:            "GlobalNames",
		IdentifierSpace: "https://verifier.globalnames.org/api/v1/name_strings/",
		SchemaSpace:     "http://apidoc.globalnames.org/gnames#",
		DefaultTypes:    types,
		Preview:         preview,
		View:            view,
		BatchSize:       50,
		Extend:          ext,
	}
	return c.JSON(http.StatusOK, res)
}

func nameInfoGET(gn gnames.GNames) func(echo.Context) error {
	return func(c echo.Context) error {
		txt := `
The "Scientific Name-String" represents a scientific biological name. The
generation of scientific names is governed by codes of nomenclature, including:

The International Code of Zoological Nomenclature
The International Code of Nomenclature for algae, fungi, and plants
The International Code of Nomenclature for Cultivated Plants
The International Code of Nomenclature of Prokaryotes
The International Code of Virus Classification and Nomenclature

However, these codes do not provide universal, strict rules on how names must
be spelled out. As a result, a scientific name can be represented by various
name-strings.

Examples of different representations for "Carex scirpoidea var.
convoluta" are:

Carex scirpoidea var. convoluta
Carex scirpoidea var. convoluta Kük.
Carex scirpoidea Michx. var. convoluta Kükenth.
Carex scirpoidea var. convoluta Kükenthal

It's important to note that the code for viruses does not follow binomial
nomenclature.

This endpoint resolves a specific spelling of a scientific name to data-source
records where that name-string was used. Appending the URL with the name-string
GlobalNames identifier, or the string itself will provide more details of the
"best match".
		
It is also possible to expand search to all relevant results, and
filter out results to particular data-sources by providing their IDs.

Example:

https://verifier.globalnames.org/api/v1/name_strings/0eeccd70-eaf2-5c51-ad8b-46cfb3db1645?all_matches=true&data_sources=1,11

The list of DataSource IDs can be found at

https://verifier.globalnames.org/api/v1/data_sources

or

https://verifier.globalnames.org/data_sources
`
		return c.String(http.StatusOK, txt)
	}
}

func nameGET(gn gnames.GNames) func(echo.Context) error {
	return func(c echo.Context) error {
		idStr := c.Param("id")
		if idStr == "" {
			err := errors.New("empty id input")
			return err
		}

		if _, err := uuid.Parse(idStr); err != nil {
			idStr = gnuuid.New(idStr).String()
		}
		var ds []int
		dsStr, _ := url.QueryUnescape(c.QueryParam("data_sources"))
		matches := c.QueryParam("all_matches") == "true"
		for _, v := range strings.Split(dsStr, ",") {
			if id, err := strconv.Atoi(v); err == nil {
				ds = append(ds, id)
			}
		}
		params := vlib.NameStringInput{
			ID:             idStr,
			DataSources:    ds,
			WithAllMatches: matches,
		}

		name, err := gn.NameByID(params, false)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, name)
	}
}

func verificationPOST(gn gnames.GNames) func(echo.Context) error {
	return func(c echo.Context) error {
		ctx, cancel := getContext(c)
		defer cancel()
		chErr := make(chan error)

		go func() {
			defer close(chErr)

			var err error
			var verified vlib.Output
			var params vlib.Input

			err = c.Bind(&params)

			if err == nil {
				verified, err = gn.Verify(ctx, params)
			}

			if l := len(params.NameStrings); l > 0 {
				log.Info().
					Int("namesNum", l).
					Str("example", params.NameStrings[0]).
					Str("parsedBy", "REST API").
					Str("method", "POST").
					Msg("Verification")
			}

			if err == nil {
				err = c.JSON(http.StatusOK, verified)
			}

			chErr <- err
		}()

		select {
		case <-ctx.Done():
			<-chErr
			return ctx.Err()
		case err := <-chErr:
			return err
		case <-time.After(6 * time.Minute):
			return errors.New("request took too long")
		}
	}
}

func verificationGET(gn gnames.GNames) func(echo.Context) error {
	return func(c echo.Context) error {
		nameStr, _ := url.QueryUnescape(c.Param("names"))
		names := strings.Split(nameStr, "|")
		dsStr, _ := url.QueryUnescape(c.QueryParam("data_sources"))
		capitalize := c.QueryParam("capitalize") == "true"
		spGrp := c.QueryParam("species_group") == "true"
		stats := c.QueryParam("stats") == "true"
		fuzzyUni := c.QueryParam("fuzzy_uninomial") == "true"
		mainTxnThresholdStr := c.QueryParam("main_taxon_threshold")
		matches := c.QueryParam("all_matches") == "true"

		mainTxnThreshold, _ := strconv.ParseFloat(mainTxnThresholdStr, 64)
		var ds []int
		for _, v := range strings.Split(dsStr, "|") {
			if id, err := strconv.Atoi(v); err == nil {
				ds = append(ds, id)
			}
		}

		params := vlib.Input{
			NameStrings:             names,
			DataSources:             ds,
			WithCapitalization:      capitalize,
			WithAllMatches:          matches,
			WithStats:               stats,
			WithSpeciesGroup:        spGrp,
			WithUninomialFuzzyMatch: fuzzyUni,
			MainTaxonThreshold:      float32(mainTxnThreshold),
		}
		verified, err := gn.Verify(context.Background(), params)
		if err != nil {
			return err
		}
		if l := len(names); l > 0 {
			log.Info().
				Int("namesNum", l).
				Str("example", names[0]).
				Str("parsedBy", "REST API").
				Str("method", "GET").
				Msg("Verification")
		}
		return c.JSON(http.StatusOK, verified)
	}
}

func searchGET(gn gnames.GNames) func(echo.Context) error {
	return func(c echo.Context) error {
		q, _ := url.QueryUnescape(c.Param("query"))
		gnq := gnquery.New()
		inp := gnq.Parse(q)
		res := gn.Search(context.Background(), inp)

		log.Info().
			Str("query", q).
			Str("parsedBy", "REST API").
			Str("method", "GET").
			Msg("Search")

		return c.JSON(http.StatusOK, res)
	}
}

func searchPOST(gn gnames.GNames) func(echo.Context) error {
	return func(c echo.Context) error {
		ctx, cancel := getContext(c)
		defer cancel()
		chErr := make(chan error)

		go func() {
			defer close(chErr)

			var err error
			var res search.Output
			var params search.Input

			err = c.Bind(&params)

			params = gnquery.New().Process(params)

			if err == nil {
				res = gn.Search(ctx, params)
				log.Info().
					Str("query", params.Query).
					Str("parsedBy", "REST API").
					Str("method", "POST").
					Msg("Search")
				err = c.JSON(http.StatusOK, res)
			}

			chErr <- err
		}()

		select {
		case <-ctx.Done():
			<-chErr
			return ctx.Err()
		case err := <-chErr:
			return err
		case <-time.After(6 * time.Minute):
			return errors.New("request took too long")
		}
	}
}

func getContext(c echo.Context) (ctx context.Context, cancel func()) {
	ctx = c.Request().Context()
	ctx, cancel = context.WithTimeout(ctx, 5*time.Minute)
	return ctx, cancel
}

func setLogger(e *echo.Echo, g gnames.GNames) nsq.NSQ {
	cfg := g.GetConfig()
	nsqAddr := cfg.NsqdTCPAddress
	withLogs := cfg.WithWebLogs
	contains := cfg.NsqdContainsFilter
	regex := cfg.NsqdRegexFilter

	if nsqAddr != "" {
		cfg := nsqcfg.Config{
			StderrLogs: withLogs,
			Topic:      "gnames",
			Address:    nsqAddr,
			Contains:   contains,
			Regex:      regex,
		}
		remote, err := nsqio.New(cfg)
		logCfg := middleware.DefaultLoggerConfig
		if err == nil {
			logCfg.Output = remote
			// set app logger too
			log.Logger = log.Output(remote)
		}
		e.Use(middleware.LoggerWithConfig(logCfg))
		if err != nil {
			log.Warn().Err(err)
		}
		return remote
	} else if withLogs {
		e.Use(middleware.Logger())
		return nil
	}
	return nil
}
