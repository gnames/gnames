PROJ_NAME = gnames

VERSION = $(shell git describe --tags)
VER = $(shell git describe --tags --abbrev=0)
DATE = $(shell date -u '+%Y-%m-%d_%H:%M:%S%Z')

NO_C = CGO_ENABLED=0
FLAGS_SHARED = $(NO_C) GOARCH=amd64
FLAGS_LD = -ldflags "-X github.com/gnames/$(PROJ_NAME)/pkg.Build=$(DATE) \
                  -X github.com/gnames/$(PROJ_NAME)/pkg.Version=$(VERSION)"
FLAGS_REL = -trimpath -ldflags "-s -w \
						-X github.com/gnames/$(PROJ_NAME)/pkg.Build=$(DATE)"

GOCMD = go
GOBUILD = $(GOCMD) build $(FLAGS_LD)
GORELEASE = $(GOCMD) build $(FLAGS_REL)
GOINSTALL = $(GOCMD) install $(FLAGS_LD)
GOCLEAN = $(GOCMD) clean
GOGENERATE = $(GOCMD) generate
GOGET = $(GOCMD) get

.PHONY: all test tools deps build buildrel install release docker dc dockerhub

all: install

test: deps install
	@echo Run tests
	$(GOCMD) test -shuffle=on -count=1 -race -coverprofile=coverage.txt -covermode=atomic ./...

tools: deps
	@echo Installing tools from tools.go
	@cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

generate:
	@echo Generate code
	$(GOGENERATE)

deps:
	@echo Download go.mod dependencies
	$(GOCMD) mod download; \
	$(GOGENERATE)

build: generate
	@echo Building
	$(GOCLEAN); \
	$(NO_C) $(GOBUILD);

buildrel: generate
	@echo Building release binary
	$(GOCLEAN); \
	$(NO_C) $(GORELEASE);

install: generate
	@echo Build and install locally
	$(NO_C) $(GOINSTALL);

release: dockerhub
	@echo Make release
	$(GOCLEAN); \
	$(FLAGS_SHARED) GOOS=linux $(GORELEASE); \
	tar zcvf /tmp/$(PROJ_NAME)-$(VER)-linux.tar.gz $(PROJ_NAME); \
	$(GOCLEAN);

docker: buildrel
	@echo Build Docker images
	docker buildx build -t gnames/$(PROJ_NAME):latest -t gnames/$(PROJ_NAME):$(VERSION) .; \
	$(GOCLEAN);

dc: build
	@echo Build Docker Compose
	docker-compose build;

dockerhub: docker
	@echo Push Docker images to DockerHub
	docker push gnames/$(PROJ_NAME); \
	docker push gnames/$(PROJ_NAME):$(VERSION)
help:
	@echo "make test - run tests"
	@echo "make tools - install tools from tools.go"
	@echo "make generate - generate code"
	@echo "make deps - download go.mod dependencies"
	@echo "make build - build binary"
	@echo "make buildrel - build release binary"
	@echo "make install - build and install locally"
	@echo "make release - make release"
	@echo "make docker - build Docker images"
	@echo "make dc - build Docker Compose"
	@echo "make dockerhub - push Docker images to DockerHub"
	@echo "make help - print this help"
