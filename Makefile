ZAI_PKG := github.com/zaibyte/zaipkg

LDFLAGS += -X "$(ZAI_PKG)/version.ReleaseVersion=$(shell git describe --tags --dirty)"
LDFLAGS += -X "$(ZAI_PKG)/version.GitHash=$(shell git rev-parse HEAD)"
LDFLAGS += -X "$(ZAI_PKG)/version.GitBranch=$(shell git rev-parse --abbrev-ref HEAD)"

TEST_PKGS := $(shell find . -iname "*_test.go" -exec dirname {} \; | \
                     uniq | sed -e "s/^\./github.com\/zaibyte\/zaipkg/")

all: test

test:
	go test -race -cover $(TEST_PKGS)

tidy:
	@echo "go mod tidy"
	go mod tidy
	git diff --quiet

clean:
	rm -rf ./bin/*

.PHONY: all tidy clean