package matcher

import (
	"path/filepath"

	gnmatcher "github.com/gnames/gnmatcher/pkg"
	gnmcfg "github.com/gnames/gnmatcher/pkg/config"
	gncfg "github.com/gnames/gnames/pkg/config"
)

// NewLib creates an embedded gnmatcher instance, initialises it, and returns
// it. An error is returned if initialisation fails (e.g. DB unreachable).
func NewLib(cfg gncfg.Config) (gnmatcher.GNmatcher, error) {
	gnm := gnmatcher.New(toMatcherConfig(cfg))
	return gnm, gnm.Init()
}

func toMatcherConfig(cfg gncfg.Config) gnmcfg.Config {
	return gnmcfg.New(
		gnmcfg.OptCacheDir(filepath.Join(cfg.CacheDir, "gnmatcher")),
		gnmcfg.OptJobsNum(cfg.JobsNum),
		gnmcfg.OptMaxEditDist(cfg.MaxEditDist),
		gnmcfg.OptPgHost(cfg.PgHost),
		gnmcfg.OptPgUser(cfg.PgUser),
		gnmcfg.OptPgPass(cfg.PgPass),
		gnmcfg.OptPgPort(cfg.PgPort),
		gnmcfg.OptPgDB(cfg.PgDB),
	)
}
