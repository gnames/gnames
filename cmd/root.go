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
	_ "embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	gnames "github.com/gnames/gnames/pkg"
	gncnf "github.com/gnames/gnames/pkg/config"
	"github.com/gnames/gnsys"
	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

//go:embed gnames.yaml
var configText string

var (
	opts []gncnf.Option
)

// config purpose is to achieve automatic import of data from the
// configuration file, if it exists.
type cfgData struct {
	CacheDir      string
	JobsNum       int
	MatcherURL    string
	WebPageURL    string
	GnamesHostURL string
	MaxEditDist   int
	PgDB          string
	PgHost        string
	PgPass        string
	PgPort        int
	PgUser        string
	Port          int
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gnames",
	Short: "Verifies scientific names.",
	Long: `Provides reconciliation and resolution services for verification of
scientific names.

The app has provides REST API for GNverifier and stand-alone use.`,

	Run: func(cmd *cobra.Command, _ []string) {
		if showVersionFlag(cmd) {
			os.Exit(0)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	if err := rootCmd.Execute(); err != nil {
		slog.Error("Cannot execute", "error", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("version", "V", false, "Return version")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	configFile := "gnames"
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Error("Cannot find home directory", "error", err)
		os.Exit(1)
	}
	home = filepath.Join(home, ".config")

	// Search config in home directory with name ".gnames" (without extension).
	viper.AddConfigPath(home)
	viper.SetConfigName(configFile)

	// Set environment variables to override
	// config file settings
	_ = viper.BindEnv("CacheDir", "GN_CACHE_DIR")
	_ = viper.BindEnv("JobsNum", "GN_JOBS_NUM")
	_ = viper.BindEnv("MatcherURL", "GN_MATCHER_URL")
	_ = viper.BindEnv("WebPageURL", "GN_WEB_PAGE_URL")
	_ = viper.BindEnv("GnamesHostURL", "GN_GNAMES_HOST_URL")
	_ = viper.BindEnv("MaxEditDist", "GN_MAX_EDIT_DIST")
	_ = viper.BindEnv("PgHost", "GN_PG_HOST")
	_ = viper.BindEnv("PgPort", "GN_PG_PORT")
	_ = viper.BindEnv("PgUser", "GN_PG_USER")
	_ = viper.BindEnv("PgPass", "GN_PG_PASS")
	_ = viper.BindEnv("PgDB", "GN_PG_DB")
	_ = viper.BindEnv("Port", "GN_PORT")

	viper.AutomaticEnv() // read in environment variables that match

	configPath := filepath.Join(home, fmt.Sprintf("%s.yaml", configFile))
	touchConfigFile(configPath)

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		slog.Info("Using config file", "config", viper.ConfigFileUsed())
	}
	getOpts()
}

func getOpts() []gncnf.Option {
	cfg := &cfgData{}
	err := viper.Unmarshal(cfg)
	if err != nil {
		slog.Error("Cannot deserialize config data", "error", err)
		os.Exit(1)
	}

	if cfg.Port != 0 {
		opts = append(opts, gncnf.OptGNPort(cfg.Port))
	}

	if cfg.CacheDir != "" {
		opts = append(opts, gncnf.OptWorkDir(cfg.CacheDir))
	}
	if cfg.JobsNum != 0 {
		opts = append(opts, gncnf.OptJobsNum(cfg.JobsNum))
	}
	if cfg.MaxEditDist != 0 {
		opts = append(opts, gncnf.OptMaxEditDist(cfg.MaxEditDist))
	}
	if cfg.PgHost != "" {
		opts = append(opts, gncnf.OptPgHost(cfg.PgHost))
	}
	if cfg.PgPort != 0 {
		opts = append(opts, gncnf.OptPgPort(cfg.PgPort))
	}
	if cfg.PgUser != "" {
		opts = append(opts, gncnf.OptPgUser(cfg.PgUser))
	}
	if cfg.PgPass != "" {
		opts = append(opts, gncnf.OptPgPass(cfg.PgPass))
	}
	if cfg.PgDB != "" {
		opts = append(opts, gncnf.OptPgDB(cfg.PgDB))
	}
	if cfg.MatcherURL != "" {
		opts = append(opts, gncnf.OptMatcherURL(cfg.MatcherURL))
	}
	if cfg.WebPageURL != "" {
		opts = append(opts, gncnf.OptWebPageURL(cfg.WebPageURL))
	}
	if cfg.GnamesHostURL != "" {
		opts = append(opts, gncnf.OptGnamesHostURL(cfg.GnamesHostURL))
	}
	return opts
}

// showVersionFlag provides version and the build timestamp. If it returns
// true, it means that version flag was given.
func showVersionFlag(cmd *cobra.Command) bool {
	hasVersionFlag, err := cmd.Flags().GetBool("version")
	if err != nil {
		slog.Error("Cannot get version flag", "error", err)
	}

	if hasVersionFlag {
		fmt.Printf("\nversion: %s\nbuild: %s\n\n", gnames.Version, gnames.Build)
	}
	return hasVersionFlag
}

// touchConfigFile checks if config file exists, and if not, it gets created.
func touchConfigFile(configPath string) {
	if ok, err := gnsys.FileExists(configPath); ok && err == nil {
		return
	}

	slog.Info("Creating config file", "file", configPath)
	createConfig(configPath)
}

// createConfig creates config file.
func createConfig(path string) {
	err := gnsys.MakeDir(filepath.Dir(path))
	if err != nil {
		slog.Error("Cannot create dir", "dir", path)
		os.Exit(1)
	}

	err = os.WriteFile(path, []byte(configText), 0644)
	if err != nil {
		slog.Error("Cannot write to file", "path", path)
		os.Exit(1)
	}
}
