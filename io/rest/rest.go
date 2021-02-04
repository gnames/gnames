package rest

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
)

const withLogs = true

// Run starts HTTP/1 service for scientific names verification.
func Run(vs VerifierService) {
	log.Printf("Starting the HTTP API server on port %d.", vs.Port())
	e := echo.New()
	e.Use(middleware.Gzip())
	e.Use(middleware.CORS())
	if withLogs {
		e.Use(middleware.Logger())
	}

	e.GET("/api/v1/ping", ping(vs))
	e.GET("/api/v1/version", ver(vs))
	e.GET("/api/v1/data_sources", dataSources(vs))
	e.GET("/api/v1/data_sources/:id", oneDataSource(vs))
	e.POST("/api/v1/verifications", verification(vs))
	e.GET("/api/v1/verifications/:names", getVerification(vs))

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
			return fmt.Errorf("Cannot find DataSource for id '%s'", idStr)
		}
		return c.JSON(http.StatusOK, dataSources[0])
	}
}

func verification(vs VerifierService) func(echo.Context) error {
	return func(c echo.Context) error {
		var params vlib.VerifyParams
		if err := c.Bind(&params); err != nil {
			return err
		}
		verified, err := vs.Verify(params)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, verified)
	}
}

func getVerification(vs VerifierService) func(echo.Context) error {
	return func(c echo.Context) error {
		nameStr, _ := url.QueryUnescape(c.Param("names"))
		names := strings.Split(nameStr, "|")
		var prefs []int
		prefsStr, _ := url.QueryUnescape(c.QueryParam("pref_sources"))
		for _, v := range strings.Split(prefsStr, "|") {
			if id, err := strconv.Atoi(v); err == nil {
				prefs = append(prefs, id)
			}
		}
		params := vlib.VerifyParams{
			NameStrings:      names,
			PreferredSources: prefs,
		}
		verified, err := vs.Verify(params)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, verified)
	}
}
