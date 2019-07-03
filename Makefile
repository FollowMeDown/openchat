#!/usr/bin/make -f

PACKAGES_NOSIMULATION=$(shell go list ./... | grep -v '/simulation')
PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
CAT := $(if $(filter $(OS),Windows_NT),type,cat)
LEDGER_ENABLED ?= true
GOBIN ?= $(GOPATH)/bin
GOSUM := $(shell which gosum)

export GO111MODULE = on

# process build tags

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support)
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

ifeq ($(WITH_CLEVELDB),yes)
  build_tags += gcc
endif
build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))


# process linker flags

ldflags = -X github.com/openchatproject/openchat/version.Version=$(VERSION) \
		  -X github.com/openchatproject/openchat/version.Commit=$(COMMIT) \
		  -X "github.com/openchatproject/openchat/version.BuildTags=$(build_tags)"

ifneq ($(GOSUM),)
ldflags += -X github.com/cosmos/cosmos-sdk/version.GoSumHash=$(shell $(GOSUM) go.sum)
endif

ifeq ($(WITH_CLEVELDB),yes)
  ldflags += -X github.com/cosmos/cosmos-sdk/types.DBBackend=cleveldb
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'

# The below include contains the tools target.
include devtools/Makefile

all: tools install lint test

########################################
### Build/Install

build: go.sum
ifeq ($(OS),Windows_NT)
	go build -mod=readonly $(BUILD_FLAGS) -o build/chatd.exe ./cmd/chatd
	go build -mod=readonly $(BUILD_FLAGS) -o build/chatcli.exe ./cmd/chatcli
	go build -mod=readonly $(BUILD_FLAGS) -o build/chatdebug.exe ./cmd/chatdebug
else
	go build -mod=readonly $(BUILD_FLAGS) -o build/chatd ./cmd/chatd
	go build -mod=readonly $(BUILD_FLAGS) -o build/chatcli ./cmd/chatcli
	go build -mod=readonly $(BUILD_FLAGS) -o build/chatdebug ./cmd/chatdebug
endif

build-linux: go.sum update-chat-lite-docs
	LEDGER_ENABLED=false GOOS=linux GOARCH=amd64 $(MAKE) build

install: go.sum check-ledger update-chat-lite-docs
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/chatd
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/chatcli

install-debug: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/chatdebug

########################################
### Tools & dependencies

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

go.sum: tools go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

draw-deps: tools
	@# requires brew install graphviz or apt-get install graphviz
	go get github.com/RobotsAndPencils/goviz
	@goviz -i ./cmd/chatd -d 2 | dot -Tpng -o dependency-graph.png

update-chat-lite-docs:
	@statik -src=client/lcd/swagger-ui -dest=client/lcd/ -f

clean:
	rm -rf snapcraft-local.yaml build/

distclean: clean
	rm -rf vendor/

########################################
### Testing

check: check-unit check-build
check-all: check check-race check-cover

check-unit:
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./...

check-race:
	@VERSION=$(VERSION) go test -mod=readonly -race -tags='ledger test_ledger_mock' ./...

check-cover:
	@go test -mod=readonly -timeout 30m -race -coverprofile=coverage.txt -covermode=atomic -tags='ledger test_ledger_mock' ./...

check-build: build
	@go test -mod=readonly -p 4 `go list ./cli_test/...` -tags=cli_test

lint: tools ci-lint
ci-lint:
	golangci-lint run
	go vet -composites=false -tests=false ./...
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -d -s
	go mod verify

format: tools
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs misspell -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs goimports -w -local github.com/cosmos/cosmos-sdk

benchmark:
	@go test -mod=readonly -bench=. $(PACKAGES_NOSIMULATION)


########################################
### Local validator nodes using docker and docker-compose

build-docker-chatdnode:
	$(MAKE) -C networks/local

# Run a 4-node testnet locally
localnet-start: localnet-stop
	@if ! [ -f build/node0/chatd/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/chatd:Z openchat/chatdnode testnet --v 4 -o . --starting-ip-address 192.168.20.2 ; fi
	docker-compose up -d

# Stop testnet
localnet-stop:
	docker-compose down


.PHONY: all build-linux install install-debug \
	go-mod-cache draw-deps clean \
	check check-all check-build check-cover check-ledger check-unit check-race

