package rest

import (
	"fmt"
	"net/http"
	"strconv"

	vlib "github.com/gnames/gnlib/domain/entity/verifier"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
)

// Run starts HTTP/1 service for scientific names verification.
func Run(vs VerifierService) {
	log.Printf("Starting the HTTP API server on port %d.", vs.Port())
	e := echo.New()
	e.Use(middleware.Gzip())
	e.Use(middleware.CORS())
	// e.Use(middleware.Logger())

	e.GET("/api/v1/ping", ping(vs))
	e.GET("/api/v1/version", ver(vs))
	e.GET("/api/v1/data_sources", dataSources(vs))
	e.GET("/api/v1/data_sources/:id", oneDataSource(vs))
	e.POST("/api/v1/verifications", verification(vs))

	addr := fmt.Sprintf(":%d", vs.Port())
	e.Logger.Fatal(e.Start(addr))
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
