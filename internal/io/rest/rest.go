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

var apiPath = "/api/v1/"

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
	e.GET(apiPath+"name_strings/:id", nameGET(gn))
	e.POST(apiPath+"verifications", verificationPOST(gn))
	e.GET(apiPath+"verifications/:names", verificationGET(gn))
	e.POST(apiPath+"search", searchPOST(gn))
	e.GET(apiPath+"search/:query", searchGET(gn))
	e.GET(apiPath+"reconcile", manifestGET())

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

func manifestGET() func(echo.Context) error {
	types := []reconciler.TypeDesc{
		{
			ID:   "globalnames.org/name_string",
			Name: "NameString",
		},
	}
	return func(c echo.Context) error {
		res := reconciler.Manifest{
			Versions:        []string{"0.2"},
			Name:            "GlobalNames",
			IdentifierSpace: "https://verifier.globalnames.org/api/v1/name_strings/",
			// TODO: change to complete URL
			SchemaSpace:  "http://apidoc.globalnames.org/gnames",
			DefaultTypes: types,
		}
		return c.JSON(http.StatusOK, res)
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
		for _, v := range strings.Split(dsStr, "|") {
			if id, err := strconv.Atoi(v); err == nil {
				ds = append(ds, id)
			}
		}
		params := vlib.NameStringInput{
			ID:             idStr,
			DataSources:    ds,
			WithAllMatches: matches,
		}

		name, err := gn.NameByID(params)
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
					Str("method", "GET").
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
			NameStrings:        names,
			DataSources:        ds,
			WithCapitalization: capitalize,
			WithAllMatches:     matches,
			WithStats:          stats,
			WithSpeciesGroup:   spGrp,
			MainTaxonThreshold: float32(mainTxnThreshold),
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
