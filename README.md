# GNames

[![API](https://img.shields.io/badge/OpenAPI3-1.0.0-89bf04)][restful api documentation]
[![GoDoc](https://godoc.org/github.com/gnames/gnames?status.svg)][godoc]

- [RESTful API Documentation]
- [RESTful API Example]

The goal of the [GNames] project is to provide an accurate and fast
verification of scientific names in unlimited quantities. The verification
should be fast (at least 1000 names per second) and include exact and fuzzy
matching of input strings to scientific names aggregated from a large number
of data-sources.

In case if you do not need exact records of matched names from data-sources,
and just want to know if a name-string is known, you can use [GNmatcher]
instead of this project. The [GNmatcher] is significantly faster and has
simpler output.

<!-- vim-markdown-toc GFM -->

- [Features](#features)
- [Installation](#installation)
  - [Installation prerequesites](#installation-prerequesites)
  - [Installation process](#installation-process)
- [Configuration](#configuration)
- [Usage as API](#usage-as-api)
- [Usage with GNverifier](#usage-with-gnverifier)
- [Web-Logs](#web-logs)
- [Known limitations of the verification](#known-limitations-of-the-verification)
- [Development](#development)
- [Authors](#authors)
- [License](#license)

<!-- vim-markdown-toc -->

## Features

- Fast verification of unlimited number of scientific names.
- Multiple levels of verification:
  - `Exact` matching (exact string match for viruses, exact canonical form
    match for Plantae, Fungi, Bacteria, and Animalia).
  - `Fuzzy` matching detects human and/or Optical Character Recognition
    (OCR) errors without producing large number of false positives. To avoid
    false positives uninomial names only checked for exact match.
  - `PartialExact` matching happens when a match for the full name-string is
    not found. In such cases middle or end words are removed and each
    variant is verified. Matches of names with the last word intact does have a
    preference.
  - `PartialFuzzy` matching is provided for partial matches of species and
    infraspecies. To avoid false positives uninomials only checked for exact
    match.
  - `Virus` matching provides viruses verification.
  - `FacetedSearch` allows to use flexible query language for searching.
- Providing names information from data-sources that contain a particular name.
  - Returning the "best" result. The `BestResult` is calculated by a scoring
    algorithm.
  - Optionally, limiting results to data-sources that are important
    to a [GNames] user.
- Providing outlink URLs to some data-sources websites to show the original
  record of a name.
- Providing meta-information about aggregated data-sources.

## Installation

Most of the users do not need to install [GNames] and can use remote [GNames
API] service at `http://verifier.globalnames.org/api/v0` or use a command line
client [GNverifier]. Nevertheless, it is possible to install a local copy of
the service.

### Installation prerequesites

- A Linux-based operating system.
- At least 32GB of memory.
- At least 50GB of a free disk space.
- Fast Internet connection during installation. After installation [GNames] can
  operate without remote connection.
- PostgreSQL database.

### Installation process

1. **PostgreSQL**

   We are not covering basics of PostgreSQL administration here. There are
   many tutorials and resources for Linux-based operating systems that
   can help.

   Create a database named `gnames`. Download the [gnames database
   dump][gnames dbdump]. Restore the database with:

   ```bash
   gunzip -c gnames_latest.tar.gz |pg_restore -d gnames
   ```

2. **GNmatcher**

   Refer to the [GNmatcher] documentation for its installation.

3. **GNames**

   Download the [latest release] of GNames, unpack it and place somewhere
   in the `PATH`.

   Run `gnames -V`. It will show you the version of `GNames` and also generate
   `$HOME/.config/gnames.yaml` configuration file.

   Edit `$HOME/.config/gnames.yaml` according to your preferences.

   Try it by running

   ```bash
   gnames rest -p 8888
   ```

   To load service automatically you can create systemctl configuration for
   the service, if your system supports systemctl.

   Alternatively you can use docker image to run GNames. You will need to
   create a file with corresponding environment variables that are described
   in the [.env.example] file.

   ```bash
   docker pull gnames/gnames:latest
   docker run -env_file path_to_env_file -d -i -t -p 8888:8888 \
     gnames/gnames:latest rest -p 8888
   ```

   We provide an [example of environment file]. Environment variables
   override configuration file settings.

## Configuration

Configuration settings can either be given in the config file
located at `$HOME/.config/gnames.yaml`, or by setting the following
environment variables:

| Env. Var.            | Configuration  |
| -------------------- | -------------- |
| GN_CACHE_DIR         | CacheDir       |
| GN_JOBS_NUM          | JobsNum        |
| GN_MATCHER_URL       | MatcherURL     |
| GN_MAX_EDIT_DIST     | MaxEditDist    |
| GN_PG_DB             | PgDB           |
| GN_PG_HOST           | PgHost         |
| GN_PG_PASS           | PgPass         |
| GN_PG_PORT           | PgPort         |
| GN_PG_USER           | PgUser         |
| GN_PORT              | Port           |
| GN_WEB_LOGS_NSQD_TCP | WebLogsNsqdTCP |
| GN_WITH_WEB_LOGS     | WithWebLogs    |

The meaning of configuration settings are provided in the [default gnames.yaml].

## Usage as API

Please note, that currently developed API ([documentation][gnames api]) is
publically served at `https://verifier.globalnames.org/api/v0`.

Legacy [API v1] public service is located at
`https://verifier.globalnames.org/api/v1`. Legacy API is not going to change,
but it will be deprecated, when current API will reach v2.

If you installed GNames locally and want to run its API, run:

```bash
gnames rest
# to change from default 8888 port
gnames rest -p 8787
```

Refer to GNames' [RESTful API Documentation] about interacting with GNames API.

## Usage with GNverifier

[GNverifier] is a command line client for [GNames] backend. It uses publically
available **remote** API of GNames. Install and use it according to the
[GNverifier] documentation.

[GNverifier] also provides web-based user interface to GNames. To launch it
use something like:

```bash
gnverifier -p 8777
```

## Web-Logs

By default Logs are not shown. To enable the service logs change
`WithWebLogs` to `true` in the configuration file.

To aggregate logs with an [NSQ] messaging service, provide an address for
TCP service of `nsqd`, for example `localhost:4150` by changing
`WebLogsNsqdTCP` in configuration file, or `GN_WEB_LOGS_NSQD_TCP`.

## Known limitations of the verification

- Exact matches of misspellings that might exist in poorly curated databases
  prevent to find fuzzy matches from better curated sources.

      To increase performance we stop any further tries if a name matched
      successfully. This prevents fuzzy-matching if a misspelled name is found
      somewhere. It is helpful to check 'curation' field of returned result,
      and see how many data-sources do contain the name.

- Fuzzy matching of a name where genus string is broken by a space.

  For example we cannot match 'Abro stola triplasia' to 'Abrostola
  triplasia'. There is only 1 edit distance between the strings, however we
  stem specific epithets, so in reality we fuzzy-match 'Abro stol triplas'
  to 'Abrostola triplas'. That means now we have edit distance 2 which is
  usually beyond our threshold.

## Development

- Install Go language for your Linux operating system.
- Create PostgreSQL database as described in installation.
- Clone the [GNames] code.
- Clone the [GNmatcher] and set it up for development.
- Install docker and docker compose.
- Go to your local `gnames` directory
  - Run `make dc`
  - Run `docker-compose up`
  - In another terminal window run `go test ./...`

## Authors

- [Dmitry Mozzherin]

## License

The `GNames` code is released under [MIT license].

[.env.example]: https://github.com/gnames/gnames/blob/master/.env.example
[dmitry mozzherin]: https://github.com/dimus
[gnames api]: https://apidoc.globalnames.org/gnames-beta
[api v1]: https://apidoc.globalnames.org/gnames
[gnames]: https://github.com/gnames/gnames
[gnmatcher]: https://github.com/gnames/gnmatcher
[gnverifier]: https://github.com/gnames/gnverifier
[mit license]: https://github.com/gnames/gnames/blob/master/LICENSE
[nsq]: https://nsq.io/
[restful api documentation]: https://apidoc.globalnames.org/gnames-beta
[restful api example]: https://verifier.globalnames.org/api/v0/verifications/Monochamus%20galloprovincialis?data_sources=1|12|170&all_matches=true
[default gnames.yaml]: https://github.com/gnames/gnames/blob/master/gnames/cmd/gnames.yaml
[example of environment file]: https://github.com/gnames/gnames/blob/master/.env.example
[gnames dbdump]: http://opendata.globalnames.org/dumps/gnames-latest.tar.gz
[godoc]: https://pkg.go.dev/github.com/gnames/gnames
[latest release]: https://github.com/gnames/gnames/releases/latest
