VERSION = $(shell git describe --tags)
VER = $(shell git describe --tags --abbrev=0)
DATE = $(shell date -u '+%Y-%m-%d_%H:%M:%S%Z')
FLAG_MODULE = GO111MODULE=on
FLAGS_SHARED = $(FLAG_MODULE) CGO_ENABLED=0 GOARCH=amd64
FLAGS_LD=-ldflags "-X github.com/gnames/gnames.Build=${DATE} \
                  -X github.com/gnames/gnames.Version=${VERSION}"
GOCMD=go
GOINSTALL=$(GOCMD) install $(FLAGS_LD)
GOBUILD=$(GOCMD) build $(FLAGS_LD)
GOCLEAN=$(GOCMD) clean
GOGENERATE=$(GOCMD) generate
GOGET = $(GOCMD) get

all: install

test: deps install
	$(FLAG_MODULE) go test ./...

deps:
	$(GOCMD) mod download; \
	$(GOGENERATE)

build:
	$(GOGENERATE)
	cd gnames; \
	$(GOCLEAN); \
	$(FLAGS_SHARED) $(GOBUILD);

dc: build
	docker-compose build;

release: dockerhub
	cd gnames; \
	$(GOCLEAN); \
	$(FLAGS_SHARED) GOOS=linux $(GOBUILD); \
	tar zcvf /tmp/gnames-${VER}-linux.tar.gz gnames; \
	$(GOCLEAN);

install:
	$(GOGENERATE)
	cd gnames; \
	$(FLAGS_SHARED) $(GOINSTALL);

docker: build
	docker build -t gnames/gnames:latest -t gnames/gnames:${VERSION} .; \
	cd gnames; \

dockerhub: docker
	docker push gnames/gnames; \
	docker push gnames/gnames:${VERSION}

