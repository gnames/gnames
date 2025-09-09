/*
Copyright © 2020-2023 Dmitry Mozzherin <dmozzherin@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"log/slog"
	"os"

	"github.com/gnames/gnames/internal/io/pgio"
	"github.com/gnames/gnames/internal/io/rest"
	"github.com/gnames/gnames/internal/io/srchio"
	"github.com/gnames/gnames/internal/io/verifio"
	"github.com/gnames/gnames/internal/io/vernio"
	"github.com/gnames/gnames/internal/logr"
	gnames "github.com/gnames/gnames/pkg"
	"github.com/gnames/gnames/pkg/config"
	"github.com/spf13/cobra"
)

// restCmd represents the rest command
var restCmd = &cobra.Command{
	Use:   "rest",
	Short: "HTTP interface to scientific names verification.",
	Long: `Runs an HTTP/1 service that takes a list of scientific names,
  normalizes input names and finds them in a variety of biodiversity data
  sources, returning back the results.`,
	Run: func(cmd *cobra.Command, _ []string) {
		debug, _ := cmd.Flags().GetBool("debug")
		if debug {
			logr.LogDev()
			slog.Info("Log level is set to DEBUG.")
		}
		// add default values to the logger
		slog.SetDefault(slog.Default().With(
			slog.String("gnApp", "gnames"),
		))

		port, _ := cmd.Flags().GetInt("port")
		opts = append(opts, config.OptGNPort(port))

		cfg := config.New(opts...)
		db, err := pgio.New(cfg)
		if err != nil {
			slog.Error("Cannot create DB connection", "error", err)
			os.Exit(1)
		}

		vf, err := verifio.New(cfg, db)
		if err != nil {
			slog.Error("Cannot create verifier service", "error", err)
			os.Exit(1)
		}

		vern := vernio.New(cfg, db)

		srch, err := srchio.New(cfg, db)
		if err != nil {
			slog.Error("Cannot create facet search service", "error", err)
			os.Exit(1)
		}

		gn := gnames.New(cfg, vf, vern, srch)

		rest.Run(gn, port)
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(restCmd)

	restCmd.Flags().IntP("port", "p", 8888, "REST port")
	restCmd.Flags().BoolP("debug", "d", false, "set logs level to DEBUG")
}
