# gnames

[![API](https://img.shields.io/badge/OpenAPI3-1.0.0-89bf04)][OpenAPI Specification]
[![GoDoc](https://godoc.org/github.com/gnames/gnames?status.svg)](https://pkg.go.dev/github.com/gnames/gnames)

The goal of the ``gnames`` project is to provide accurate and fast verification
of scientific names in unlimited quantities. The verification should be fast
(at least 1000 names per second) and include exact and fuzzy matching to data
from the large number of data-sources.

RESTful API of the project is described by the [OpenAPI Specification]

<!-- vim-markdown-toc GFM -->

* [Features](#features)
* [Installation](#installation)
  * [Installation prerequesites](#installation-prerequesites)
  * [Installation process](#installation-process)
* [Usage as API](#usage-as-api)
* [Usage with gnverifier](#usage-with-gnverifier)
* [Known limitations for verification](#known-limitations-for-verification)
* [Development](#development)
* [Authors](#authors)
* [License](#license)

<!-- vim-markdown-toc -->

## Features

* Fast verification of unlimited number of scientific names.
* Multiple levels of verification:
  * `Exact` matching (exact string match for viruses, exact canonical form
    match for Plantae, Fungi, Bacteria, Animalia).
  * `Fuzzy` matching is detecting human and/or Optical Character Recognigioni
    (OCR) errors without producing large number of false positives. To avoid
    false positives uninomial names only checked for exact match.
  * `PartialExact` matching for cases when the complete name-string is not
    found. In such cases middle of end words are removed and each variant is
    verified.
  * `PartialFuzzy` matching is provided for partial matches of species and
    infraspecies. To avoid false positives uninomials only checked for exact
    match.
* Returning information about data-sources that contain a particular name.
  * Returning of one "best" result. Best result is calculated by scoring system.
  * Optinally, returning multiple results for data-sources that are important
    to `gnames` user.
* Providing meta-information about aggregated data-sources.

## Installation

Most of the users do not need to install `gnames` and can use fast remote
service at `http://verifier.globalnames.org/api/v1` or use a command line
client to `gnames` [gnverifier]. Nevertheless, it is possible to install a
a local copy of the service.

### Installation prerequesites

* A Linux-based operating system.
* At least 32GB of memory.
* At least 50GB of a free disk space.
* Fast internet connection during installation. After installation `gnames` can
  operate without internet.
* PostgreSQL database.

### Installation process

1. PostgreSQL

    We are not covering basics of PostgreSQL here. There are many tutorials
    and resources that for many Linux-based operating systems.

    PostgreSQL databases has to have `C` or `C.UTF-8` collation. This
    dependency on collation will be removed in the future.

    Create `gnames` database. Download the gnames database
    [dump][gnames dbdump]. Restore the database with:

    ```bash
    gunzip -c gnames_latest.sql.gz | pg_restore -d gnames
    ```

2. `gnmatcher`

    Refer to the [gnmatcher] documentation for its installation.

    You can skip this step if you decide to use gnmatcher as a library instead
    of a REST service. The drawback of such decision is in increase of the
    time required for loading gnmatcher data from disk.

3. `gnames`

    Download the latest [release] of `gnames`, unpack it, place is somewhere
    in the PATH.

    Run `gnames -V`. It will show you the version of `gnames` and also generate
    `$HOME/.config/gnames.yaml` configuration file.

    Edit `$HOME/.config/gnames.yaml` according to your preferences.

    Try it by running

    ```bash
    gnames rest -p 8888
    ```

    To load service automatically you can create systemctl configuration for
    the service, if your system supports systemctl.

    Alternatively you can use docker image to run gnames. You will need to
    create a file with corresponding environment variables that are described
    in the [.env.example] file.

    ```bash
    docker pull gnames/gnames:latest
    docker run -env_file path_to_env_file -d -i -t -p 8888:8888 \
      gnames/gnames:latest rest -p 8888
    ```

## Usage as API

Refer to [OpenAPI Specification] about interacting with `gnames` API.
We are planning to add web-GUI in the future as well.

## Usage with gnverifier

[gnverifier] is a command line client for `gnames` backend. Install and
use it according to [gnverifier] documentation.

## Known limitations for verification

As the project evolves we will try to remove verification limitations in the
future.

* Exact matches of a misspelling to a name-string in poorly curated databases
prevent to find fuzzy matches.

To increase performance we stop any further tries if a name matched
succesfully.  This prevents finding how mispelling would fuzzy-match
name-strings in better quality databases.

* Fuzzy matching of a name where genus string is broken by a space.

For example we cannot match 'Abro stola triplasia' to 'Abrostola triplasia'.
There is only 1 edit distance between the strings, however we stem specific
epithets, so in reality we fuzzy-match 'Abro stol triplas' to 'Abrostola
triplas'.  That means now we have edit distance 2 which is usually beyond our
threshold.  See issue `https://github.com/gnames/gnmatcher/issues/19`

## Development

* Install Go language according to your Linux operating system.
* Create PostgreSQL database as described in installation.
* Clone the [gnames] code.
* Clone the [gnmatcher] and set it up for development.
* Install docker and docker compose.
* Go to your local `gnames` directory
  * Run `make dc`
  * Run `docker-compose up`
  * Run `go test ./...`

## Authors

* [Dmitry Mozzherin]

## License

The `gnames` code is released under [MIT license].

[OpenAPI Specification]: https://app.swaggerhub.com/apis-docs/dimus/gnames/1.0.0
[gnverifier]: https://github.com/gnames/gnverifier
[gnmatcher]: https://github.com/gnames/gnmatcher
[gnames dbdump]: https://opendata.globalnames.org/dumps/gnames_latest.sql.gz
[.env.example]: https://github.com/gnames/gnames/blob/master/.env.example
[gnames]: https://github.com/gnames/gnames
[MIT license]: https://github.com/gnames/gnames/blob/master/LICENSE
[Dmitry Mozzherin]: https://github.com/dimus
