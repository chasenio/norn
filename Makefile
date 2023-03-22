.PHONY: all build compile clean generate

BUILDTIME ?= $(shell date +%Y-%m-%d_%I:%M:%S)
GITCOMMIT ?= $(shell git rev-parse -q HEAD)
ifeq ($(CI_PIPELINE_ID),)
	BUILDNUMER := private
else
	BUILDNUMER := $(CI_PIPELINE_ID)
endif
VERSION ?= $(shell git describe --tags --always --dirty)

LDFLAGS = -extldflags \
		  -static \
		  -X "main.Version=$(VERSION)" \
		  -X "main.BuildTime=$(BUILDTIME)" \
		  -X "main.GitCommit=$(GITCOMMIT)" \
		  -X "main.BuildNumber=$(BUILDNUMER)"

all: build

clean:
	rm -rf bin norns-*.zip

build:
	go build -o bin/norns -ldflags "$(LDFLAGS)"


package: bin/norns-linux-amd64-$(VERSION)-$(BUILDNUMER)
	echo $(GITCOMMIT) > commit.txt
	echo $(VERSION) > version.txt
	zip -r norns-$(VERSION)-$(BUILDNUMER).zip bin commit.txt version.txt


bin/norns-linux-amd64-$(VERSION)-$(BUILDNUMER):
	go build -o bin/norns-linux-amd64-$(VERSION)-$(BUILDNUMER) -ldflags "$(LDFLAGS)"

compile: bin/norns-linux-amd64-$(VERSION)-$(BUILDNUMER)
