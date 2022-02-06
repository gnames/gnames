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
	"github.com/gnames/gnquery"
	"github.com/gnames/gnquery/ent/search"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	nsqcfg "github.com/sfgrp/lognsq/config"
	"github.com/sfgrp/lognsq/ent/nsq"
	"github.com/sfgrp/lognsq/io/nsqio"
	log "github.com/sirupsen/logrus"
)

var apiPath = "/api/v0/"

// Run starts HTTP/1 service on given port for scientific names verification.
func Run(gn gnames.GNames, port int) {
	log.Printf("Starting the HTTP API server on port %d.", port)
	e := echo.New()
	e.Use(middleware.Gzip())
	e.Use(middleware.CORS())

	loggerNSQ := setLogger(e, gn)
	if loggerNSQ != nil {
		defer loggerNSQ.Stop()
	}

	e.GET(apiPath+"ping", ping())
	e.GET(apiPath+"version", ver(gn))
	e.GET(apiPath+"data_sources", dataSources(gn))
	e.GET(apiPath+"data_sources/:id", oneDataSource(gn))
	e.POST(apiPath+"verifications", verificationPOST(gn))
	e.GET(apiPath+"verifications/:names", verificationGET(gn))
	e.POST(apiPath+"search", searchPOST(gn))
	e.GET(apiPath+"search/:query", searchGET(gn))

	addr := fmt.Sprintf(":%d", port)
	s := &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
	}
	e.Logger.Fatal(e.StartServer(s))
}

func ping() func(echo.Context) error {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	}
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
		txContext := c.QueryParam("context") == "true"
		ctxThresholdStr := c.QueryParam("context_threshold")
		matches := c.QueryParam("all_matches") == "true"

		ctxThreshold, _ := strconv.ParseFloat(ctxThresholdStr, 64)
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
			WithContext:        txContext,
			ContextThreshold:   float32(ctxThreshold),
		}
		verified, err := gn.Verify(context.Background(), params)
		if err != nil {
			return err
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
	nsqAddr := g.WebLogsNsqdTCP()
	withLogs := g.WithWebLogs()

	if nsqAddr != "" {
		cfg := nsqcfg.Config{
			StderrLogs: withLogs,
			Topic:      "gnames",
			Address:    nsqAddr,
		}
		remote, err := nsqio.New(cfg)
		logCfg := middleware.DefaultLoggerConfig
		if err == nil {
			logCfg.Output = remote
		}
		e.Use(middleware.LoggerWithConfig(logCfg))
		if err != nil {
			log.Warn(err)
		}
		return remote
	} else if withLogs {
		e.Use(middleware.Logger())
		return nil
	}
	return nil
}
