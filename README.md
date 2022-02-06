# GNames

[![API](https://img.shields.io/badge/OpenAPI3-1.0.0-89bf04)][OpenAPI Specification]
[![GoDoc](https://godoc.org/github.com/gnames/gnames?status.svg)][godoc]

The goal of the [GNames] project is to provide an accurate and fast
verification of scientific names in unlimited quantities. The verification
should be fast (at least 1000 names per second) and include exact and fuzzy
matching of input strings to scientific names aggregated from a large number
of data-sources.

In case if you do not need exact records of matched names from data-sources,
and just want to know if a name-string is known, you can use [GNmatcher]
instead of this project. The [GNmatcher] is significantly faster and has
simpler output.

RESTful API of the project is described using [OpenAPI Specification].

<!-- vim-markdown-toc GFM -->

* [Features](#features)
* [Installation](#installation)
  * [Installation prerequesites](#installation-prerequesites)
  * [Installation process](#installation-process)
* [Usage as API](#usage-as-api)
* [Usage with GNverifier](#usage-with-gnverifier)
* [Known limitations of the verification](#known-limitations-of-the-verification)
* [Development](#development)
* [Authors](#authors)
* [License](#license)

<!-- vim-markdown-toc -->

## Features

* Fast verification of unlimited number of scientific names.
* Multiple levels of verification:
  * `Exact` matching (exact string match for viruses, exact canonical form
    match for Plantae, Fungi, Bacteria, and Animalia).
  * `Fuzzy` matching detects human and/or Optical Character Recognition
    (OCR) errors without producing large number of false positives. To avoid
    false positives uninomial names only checked for exact match.
  * `PartialExact` matching happens when a match for the full name-string is
    not found. In such cases middle or end words are removed and each
    variant is verified. Matches of names with the last word intact does have a
    preference.
  * `PartialFuzzy` matching is provided for partial matches of species and
    infraspecies. To avoid false positives uninomials only checked for exact
    match.
  * `Virus` matching provides viruses verification.
  * `FacetedSearch` allows to use flexible query language for searching.
* Providing names information from data-sources that contain a particular name.
  * Returning the "best" result. The `BestResult` is calculated by a scoring
    algorithm.
  * Optionally, returning multiple results for data-sources that are important
    to a [GNames] user.
* Providing outlink URLs to some data-sources websites to show the original
  record of a name.
* Providing meta-information about aggregated data-sources.

## Installation

Most of the users do not need to install [GNames] and can use remote [GNames
API] service at `http://verifier.globalnames.org/api/v0` or use a command line
client [GNverifier]. Nevertheless, it is possible to install a local copy of
the service.

### Installation prerequesites

* A Linux-based operating system.
* At least 32GB of memory.
* At least 50GB of a free disk space.
* Fast Internet connection during installation. After installation [GNames] can
  operate without remote connection.
* PostgreSQL database.

### Installation process

1. **PostgreSQL**

    We are not covering basics of PostgreSQL administration here. There are
    many tutorials and resources for Linux-based operating systems that
    can help.

    PostgreSQL database has to have `C` or `C.UTF-8` collation. This
    dependency on collation will be removed in the future.

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

## Usage as API

Refer to GNames' [OpenAPI Specification] about interacting with GNames API.

## Usage with GNverifier

[GNverifier] is a command line client for [GNames] backend. Install and
use it according to the [GNverifier] documentation.

[GNverifier] also provides web-based user interface to GNames. To launch it
use something like:

```bash
gnverifier -p 8777
```

## Known limitations of the verification

* Exact matches of misspellings that might exist in poorly curated databases
prevent to find fuzzy matches from better curated sources.

    To increase performance we stop any further tries if a name matched
    successfully. This prevents fuzzy-matching if a misspelled name is found
    somewhere. It is helpful to check 'curation' field of returned result,
    and see how many data-sources do contain the name.

* Fuzzy matching of a name where genus string is broken by a space.

    For example we cannot match 'Abro stola triplasia' to 'Abrostola
    triplasia'. There is only 1 edit distance between the strings, however we
    stem specific epithets, so in reality we fuzzy-match 'Abro stol triplas'
    to 'Abrostola triplas'. That means now we have edit distance 2 which is
    usually beyond our threshold.

## Development

* Install Go language for your Linux operating system.
* Create PostgreSQL database as described in installation.
* Clone the [GNames] code.
* Clone the [GNmatcher] and set it up for development.
* Install docker and docker compose.
* Go to your local `gnames` directory
  * Run `make dc`
  * Run `docker-compose up`
  * In another terminal window run `go test ./...`

## Authors

* [Dmitry Mozzherin]

## License

The `GNames` code is released under [MIT license].

[GNames]: https://github.com/gnames/gnames
[GNames API]: https://apidoc.globalnames.org/gnames-beta
[OpenAPI Specification]: https://apidoc.globalnames.org/gnames-beta
[GNverifier]: https://github.com/gnames/gnverifier
[GNmatcher]: https://github.com/gnames/gnmatcher
[gnames dbdump]: http://opendata.globalnames.org/dumps/gnames-latest.tar.gz
[.env.example]: https://github.com/gnames/gnames/blob/master/.env.example
[MIT license]: https://github.com/gnames/gnames/blob/master/LICENSE
[Dmitry Mozzherin]: https://github.com/dimus
[godoc]: https://pkg.go.dev/github.com/gnames/gnames
[latest release]: https://github.com/gnames/gnames/releases/latest
[example of environment file]: https://github.com/gnames/gnames/blob/master/.env.example
