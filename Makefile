.PHONY: all build compile clean generate

NAME=norn
BINDIR=bin
BUILDTIME ?= $(shell date +%Y-%m-%d_%I:%M:%S)
GITCOMMIT ?= $(shell git rev-parse -q HEAD)
VERSION ?= $(shell git describe --tags --always --dirty)

LDFLAGS = -extldflags \
		  -static \
		  -X "main.Version=$(VERSION)" \
		  -X "main.BuildTime=$(BUILDTIME)" \
		  -X "main.GitCommit=$(GITCOMMIT)" \
		  -X "main.BuildNumber=$(BUILDNUMER)"

GOBUILD=CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)"
CLI=cmd/cli.go

PLATFORM_LIST = \
	linux-amd64 \
	linux-arm64 \
	darwin-amd64 \
	darwin-arm64 \
	windows-amd64

all: linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64

build:
	go build -ldflags "$(LDFLAGS)" -o bin/norn $(CLI)
	go build -ldflags "$(LDFLAGS)" -o bin/norns $(SERVE)

linux-arm64:
	GOARCH=arm64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(CLI)

linux-amd64:
	GOARCH=amd64 GOOS=linux $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(CLI)

darwin-amd64:
	GOARCH=amd64 GOOS=darwin $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(CLI)

darwin-arm64:
	GOARCH=arm64 GOOS=darwin $(GOBUILD) -o $(BINDIR)/$(NAME)-$@ $(CLI)

windows-amd64:
	GOARCH=amd64 GOOS=windows $(GOBUILD) -o $(BINDIR)/$(NAME)-$@.exe $(CLI)

lint:
	GOOS=darwin golangci-lint run ./...
	GOOS=windows golangci-lint run ./...
	GOOS=linux golangci-lint run ./...

gz_releases=$(addsuffix .gz, $(PLATFORM_LIST))

release: $(PLATFORM_LIST)

clean:
	rm $(BINDIR)/*