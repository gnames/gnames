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

	"github.com/gnames/gnames"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	nsqcfg "github.com/sfgrp/lognsq/config"
	"github.com/sfgrp/lognsq/ent/nsq"
	"github.com/sfgrp/lognsq/io/nsqio"
)

// Run starts HTTP/1 service for scientific names verification.
func Run(vs VerifierService) {
	log.Info().Msgf("Starting the HTTP API server on port %d.", vs.Port())
	e := echo.New()
	e.Use(middleware.Gzip())
	e.Use(middleware.CORS())

	loggerNSQ := setLogger(e, vs)
	if loggerNSQ != nil {
		defer loggerNSQ.Stop()
	}

	e.GET("/api/v1/ping", ping(vs))
	e.GET("/api/v1/version", ver(vs))
	e.GET("/api/v1/data_sources", dataSources(vs))
	e.GET("/api/v1/data_sources/:id", oneDataSource(vs))
	e.POST("/api/v1/verifications", verificationPOST(vs))
	e.GET("/api/v1/verifications/:names", verificationGET(vs))

	addr := fmt.Sprintf(":%d", vs.Port())
	s := &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
	}
	e.Logger.Fatal(e.StartServer(s))
}

func ping(vs VerifierService) func(echo.Context) error {
	return func(c echo.Context) error {
		result := vs.Ping()
		return c.String(http.StatusOK, result)
	}
}

func ver(vs VerifierService) func(echo.Context) error {
	return func(c echo.Context) error {
		result := vs.GetVersion()
		return c.JSON(http.StatusOK, result)
	}
}

func dataSources(vs VerifierService) func(echo.Context) error {
	return func(c echo.Context) error {
		dataSources, err := vs.DataSources()
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, dataSources)
	}
}

func oneDataSource(vs VerifierService) func(echo.Context) error {
	return func(c echo.Context) error {
		idStr := c.Param("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return err
		}
		dataSources, err := vs.DataSources(id)
		if err != nil {
			return err
		}
		if len(dataSources) == 0 {
			return fmt.Errorf("cannot find DataSource for id '%s'", idStr)
		}
		return c.JSON(http.StatusOK, dataSources[0])
	}
}

func verificationPOST(vs VerifierService) func(echo.Context) error {
	return func(c echo.Context) error {
		ctx, cancel := getContext(c)
		defer cancel()
		chErr := make(chan error)

		go func() {
			defer close(chErr)

			var err error
			var verified []*vlib.Verification
			var params vlib.VerifyParams

			err = c.Bind(&params)

			if err == nil {
				verified, err = vs.Verify(ctx, params)
			}

			if err == nil {
				err = c.JSON(http.StatusOK, verified)
			}

			if l := len(params.NameStrings); l > 0 {
				log.Info().
					Int("namesNum", l).
					Str("example", params.NameStrings[0]).
					Str("parsedBy", "REST API").
					Str("method", "GET").
					Msg("Verification")
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

func verificationGET(vs VerifierService) func(echo.Context) error {
	return func(c echo.Context) error {
		nameStr, _ := url.QueryUnescape(c.Param("names"))
		names := strings.Split(nameStr, "|")
		var prefs []int
		prefsStr, _ := url.QueryUnescape(c.QueryParam("pref_sources"))
		capitalize := c.QueryParam("capitalize") == "true"
		for _, v := range strings.Split(prefsStr, "|") {
			if id, err := strconv.Atoi(v); err == nil {
				prefs = append(prefs, id)
			}
		}
		matches := c.QueryParam("all_matches") == "true"
		params := vlib.VerifyParams{
			NameStrings:        names,
			PreferredSources:   prefs,
			WithCapitalization: capitalize,
			WithAllMatches:     matches,
		}
		verified, err := vs.Verify(context.Background(), params)
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
			Topic:      "gnames-api-v1",
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
