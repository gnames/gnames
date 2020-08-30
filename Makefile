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
	$(FLAG_MODULE) $(GOGET) github.com/spf13/cobra/cobra@v1.0.0; \
	$(FLAG_MODULE) $(GOGET) github.com/onsi/ginkgo/ginkgo@v1.12.0; \
	$(FLAG_MODULE) $(GOGET) github.com/onsi/gomega@v1.10.0; \
	$(FLAG_MODULE) $(GOGET) github.com/golang/protobuf/protoc-gen-go@v1.4.1; \
	$(GOGENERATE)

build: proto
	$(GOGENERATE)
	cd gnames; \
	$(GOCLEAN); \
	$(FLAGS_SHARED) $(GOBUILD);

release: proto
	cd gnames; \
	$(GOCLEAN); \
	$(FLAGS_SHARED) GOOS=linux $(GOBUILD); \
	tar zcvf /tmp/gnames-${VER}-linux.tar.gz gnames; \
	$(GOCLEAN); \
	$(FLAGS_SHARED) GOOS=darwin $(GOBUILD); \
	tar zcvf /tmp/gnames-${VER}-mac.tar.gz gnames; \
	$(GOCLEAN); \
	$(FLAGS_SHARED) GOOS=windows $(GOBUILD); \
	zip -9 /tmp/gnames-${VER}-win-64.zip gnames.exe; \
	$(GOCLEAN);

install: proto
	$(GOGENERATE)
	cd gnames; \
	$(FLAGS_SHARED) $(GOINSTALL);

proto:
	cd protob; \
	protoc -I . ./protob.proto --go_out=plugins=grpc:.;
