# Plan: Embed gnmatcher as a Library in gnames

## Goal

Allow gnames to run without a separate gnmatcher service. By default `gnames rest`
will use gnmatcher as an embedded library. If `MatcherURL` is explicitly configured,
gnames falls back to the existing HTTP API mode. This gives local users a single
binary/process while global deployments retain full flexibility.

## Architecture

```
Current (always HTTP):
  gnames → HTTP POST → gnmatcher REST service → bloom/trie/virusio → DB

New (dual mode):
  MatcherURL == ""   → gnames embeds gnmatcher → bloom/trie/virusio → DB (same DB)
  MatcherURL != ""   → gnames → HTTP POST → gnmatcher REST service (unchanged)
```

Both modes implement the same `gnmatcher.GNmatcher` interface. Everything downstream
in `gnames_verif.go` is untouched.

---

## Part 1 — Changes in gnmatcher

### 1.1 Add `NewLib` to `pkg/gnmatcher.go`

Add a public constructor that creates, wires, and **initialises** all internal
components without exposing them. This is the only change needed in gnmatcher.

Current `New()` already takes only `cfg config.Config` and wires internal
components (bloom/trie/virusio) — no external types are exposed:
```go
func New(cfg config.Config) GNmatcher {
    em := bloom.New(cfg)
    fm := trie.New(cfg)
    vm := virusio.New(cfg)
    return gnmatcher{
        cfg:     cfg,
        matcher: matcher.NewMatcher(em, fm, vm, cfg),
    }
}
```

`New()` performs no I/O; `Init()` loads caches and connects to the DB.
`NewLib` wraps both and surfaces the init error to the caller (3 lines):

```go
func NewLib(cfg config.Config) (GNmatcher, error) {
    gnm := New(cfg)
    return gnm, gnm.Init()
}
```

`NewLib` lives in the same file (`pkg/gnmatcher.go`). No new imports needed.

### 1.2 Release gnmatcher

Tag and release a new version of gnmatcher containing `NewLib` before updating
gnames `go.mod`.

---

## Part 2 — Changes in gnames

### 2.1 Add library-based matcher: `internal/io/matcher/matcherlib.go`

New file alongside the existing `matcher.go` (HTTP client). Contains a config
translation function and a `NewLib` constructor:

```go
package matcher

import (
    "path/filepath"

    gnmatcher "github.com/gnames/gnmatcher/pkg"
    gnmcfg "github.com/gnames/gnmatcher/pkg/config"
    gncfg "github.com/gnames/gnames/pkg/config"
)

func NewLib(cfg gncfg.Config) (gnmatcher.GNmatcher, error) {
    return gnmatcher.NewLib(toMatcherConfig(cfg))
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
```

Fields intentionally omitted from translation:
- `DataSources`, `WithSpeciesGroup`, `WithRelaxedFuzzyMatch`, `WithUninomialFuzzyMatch`
  — these are already passed as per-call opts in `gnames_verif.go:178-187`, not static config.

### 2.2 Rename existing HTTP constructor in `internal/io/matcher/matcher.go`

Rename `New(url string)` → `NewREST(url string)` for clarity. One-line change:

```go
// Before:
func New(url string) gnmatcher.GNmatcher {

// After:
func NewREST(url string) gnmatcher.GNmatcher {
```

### 2.3 Update constructor in `pkg/gnames.go`

Replace the single `matcher.New(cfg.MatcherURL)` call with a mode switch.
`New()` now returns an error to propagate gnmatcher init failures:

```go
// Before:
func New(
    cfg config.Config,
    vf verif.Verifier,
    vern vern.Vernaculars,
    sr srch.Searcher,
) GNames {
    return gnames{
        cfg:     cfg,
        vf:      vf,
        vern:    vern,
        sr:      sr,
        matcher: matcher.New(cfg.MatcherURL),
    }
}

// After:
func New(
    cfg config.Config,
    vf verif.Verifier,
    vern vern.Vernaculars,
    sr srch.Searcher,
) (GNames, error) {
    var m gnmatcher.GNmatcher
    var err error
    if cfg.MatcherURL != "" {
        m = matcher.NewREST(cfg.MatcherURL)
    } else {
        m, err = matcher.NewLib(cfg)
        if err != nil {
            return nil, err
        }
    }
    return gnames{
        cfg:     cfg,
        vf:      vf,
        vern:    vern,
        sr:      sr,
        matcher: m,
    }, nil
}
```

**Note:** The signature change (`error` return) requires updating all callers of
`gnames.New()`. Check `cmd/rest.go` — it is the only caller.

### 2.4 Update `cmd/rest.go`

Handle the new error return from `gnames.New()`. Likely a small change:

```go
// Before (approximate):
gn := gnames.New(cfg, vf, vern, sr)

// After:
gn, err := gnames.New(cfg, vf, vern, sr)
if err != nil {
    slog.Error("Cannot initialize gnames", "error", err)
    os.Exit(1)
}
```

### 2.5 Update `pkg/config/config.go` — change `MatcherURL` default

In `New()`, change:
```go
// Before:
MatcherURL: "https://matcher.globalnames.org/api/v1/",

// After:
MatcherURL: "",
```

Everything else in config stays the same. `MatcherURL` field, env var `GN_MATCHER_URL`,
and `OptMatcherURL` are all kept — they are the mechanism to opt into HTTP mode.

### 2.6 Update `cmd/gnames.yaml` — update comment for MatcherURL

```yaml
# Before:
# MatcherURL is a URL to a GNmatcher service.
# Example for localhost: http://0.0.0.0:8080/api/v1/
#
# MatcherURL: "https://matcher.globalnames.org/api/v1/"

# After:
# MatcherURL is a URL to a remote GNmatcher service.
# When set, gnames uses the remote service instead of the embedded matcher.
# Useful for global deployments with multiple gnmatcher instances.
# Example: MatcherURL: "https://matcher.globalnames.org/api/v1/"
#
# MatcherURL: ""
```

### 2.7 Update `go.mod`

Bump gnmatcher to the new version containing `NewLib`:
```
github.com/gnames/gnmatcher v1.1.23  (or whatever the new tag is)
```

During development, use a `replace` directive to point to the local gnmatcher:
```
replace github.com/gnames/gnmatcher => ../gnmatcher
```

Remove the `replace` directive before the final release.

---

## Summary of all files changed

### gnmatcher
| File | Change |
|------|--------|
| `pkg/gnmatcher.go` | Add `NewLib(cfg config.Config) (GNmatcher, error)` — calls `New(cfg)` then `Init()` |

### gnames
| File | Change |
|------|--------|
| `internal/io/matcher/matcher.go` | Rename `New` → `NewREST` |
| `internal/io/matcher/matcherlib.go` | New file: `NewLib` + `toMatcherConfig` |
| `pkg/gnames.go` | Mode switch in `New()`, add error return |
| `cmd/rest.go` | Handle error return from `gnames.New()` |
| `pkg/config/config.go` | Change `MatcherURL` default to `""` |
| `cmd/gnames.yaml` | Update `MatcherURL` comment |
| `go.mod` | Bump gnmatcher version |

**Files not touched:** `pkg/gnames_verif.go`, `pkg/interface.go`, all of `pkg/ent/`,
`internal/io/pgio/`, `internal/io/rest/`, `internal/io/verifio/`, tests.

---

## Consequences and risks

| Area | Impact |
|------|--------|
| Memory | gnames process now holds bloom filters + trie in RAM. Expect significant increase in memory usage vs current (HTTP mode uses negligible memory for the client). |
| Startup time | Slower first start — gnmatcher must build/load caches from disk. Subsequent starts are faster (caches exist). DB must be reachable at startup. |
| CacheDir | gnmatcher caches land in `~/.cache/gnames/gnmatcher/` (subdirectory of gnames CacheDir). Must be writable and persistent. In Docker, requires a volume. |
| DB | Same PostgreSQL connection as gnames. No second DB config needed. gnmatcher reads from the same `gnames` database to build its caches. |
| Deployment | Single process/image for local use. Docker image needs the gnmatcher cache volume. |
| Global deployment | No change — set `GN_MATCHER_URL` to restore HTTP mode. |
| Scalability | Embedded mode cannot distribute matching across multiple gnmatcher instances. Acceptable for local use; global operators use HTTP mode. |
| Error handling | `gnames.New()` now returns an error. Startup fails fast if gnmatcher cannot initialise (e.g. DB unreachable, bad CacheDir). |
