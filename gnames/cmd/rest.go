/*
Copyright © 2020 Dmitry Mozzherin <dmozzherin@gmail.com>

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
	"os"

	"github.com/gnames/gnames"
	gncfg "github.com/gnames/gnames/config"
	"github.com/gnames/gnames/io/facetpg"
	"github.com/gnames/gnames/io/rest"
	"github.com/gnames/gnames/io/verifierpg"
	log "github.com/sirupsen/logrus"
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
			log.SetLevel(log.DebugLevel)
			log.Printf("Log level is set to '%s'.", log.Level.String(log.GetLevel()))
		}

		port, _ := cmd.Flags().GetInt("port")
		opts = append(opts, gncfg.OptGNPort(port))

		cfg := gncfg.New(opts...)
		vf := verifierpg.New(cfg)
		srch := facetpg.New(cfg)
		gn := gnames.NewGNames(cfg, vf, srch)

		rest.Run(gn, port)
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(restCmd)

	restCmd.Flags().IntP("port", "p", 8888, "REST port")
	restCmd.Flags().BoolP("debug", "d", false, "set logs level to DEBUG")
}
