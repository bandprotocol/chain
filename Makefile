PACKAGES=$(shell go list ./... | grep -v '/simulation')
VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
LEDGER_ENABLED ?= true
BINDIR ?= $(GOPATH)/bin
MOCKS_DIR = $(CURDIR)/tests/mocks
APP = ./app

DOCKER := $(shell which docker)
DOCKER_BUF := $(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace bufbuild/buf

export GO111MODULE = on

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
	build_tags += ledger
endif

build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
empty = $(whitespace) $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(empty),$(comma),$(build_tags))

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=bandchain \
	-X github.com/cosmos/cosmos-sdk/version.AppName=bandd \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X "github.com/cosmos/cosmos-sdk/version.BuildTags=$(build_tags_comma_sep)"

ifeq ($(LINK_STATICALLY),true)
	ldflags += -linkmode=external -extldflags "-Wl,-z,muldefs -static"
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags_comma_sep)" -ldflags '$(ldflags)'

include contrib/devtools/Makefile

all: install

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/bandd
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/yoda
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/cylinder
	go install -mod=readonly $(BUILD_FLAGS) ./cmd/grogu

release: go.sum
	env GOOS=linux GOARCH=amd64 \
		go build -mod=readonly -o ./build/bandd_linux_amd64 $(BUILD_FLAGS) ./cmd/bandd
	env GOOS=darwin GOARCH=amd64 \
		go build -mod=readonly -o ./build/bandd_darwin_amd64 $(BUILD_FLAGS) ./cmd/bandd
	env GOOS=windows GOARCH=amd64 \
		go build -mod=readonly -o ./build/bandd_windows_amd64 $(BUILD_FLAGS) ./cmd/bandd
	env GOOS=linux GOARCH=amd64 \
		go build -mod=readonly -o ./build/yoda_linux_amd64 $(BUILD_FLAGS) ./cmd/yoda
	env GOOS=darwin GOARCH=amd64 \
		go build -mod=readonly -o ./build/yoda_darwin_amd64 $(BUILD_FLAGS) ./cmd/yoda
	env GOOS=windows GOARCH=amd64 \
		go build -mod=readonly -o ./build/yoda_windows_amd64 $(BUILD_FLAGS) ./cmd/yoda

mocks:
	@go install go.uber.org/mock/mockgen@latest
	sh ./scripts/mockgen.sh

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	GO111MODULE=on go mod verify
	touch go.sum

test:
	@go test -mod=readonly $(PACKAGES)

###############################################################################
###                                Protobuf                                 ###
###############################################################################

protoVer=0.13.0
protoImageName=ghcr.io/cosmos/proto-builder:$(protoVer)
protoImage=$(DOCKER) run --rm -v $(CURDIR):/workspace --workdir /workspace $(protoImageName)

proto-all: proto-format proto-lint proto-gen

proto-gen:
	@echo "Generating Protobuf files"
	@$(protoImage) sh ./scripts/protocgen.sh

proto-swagger-gen:
	@echo "Generating Protobuf Swagger"
	@$(protoImage) sh ./scripts/protoc-swagger-gen.sh
	$(MAKE) update-swagger-docs

proto-format:
	@$(protoImage) find ./ -name "*.proto" -exec clang-format -i {} \;

proto-lint:
	@$(protoImage) buf lint --error-format=json

proto-check-breaking:
	@$(protoImage) buf breaking --against $(HTTPS_GIT)#branch=main

###############################################################################
###                               Simulation                                ###
###############################################################################

test-sim-import-export: runsim
	@echo "Running application import/export simulation. This may take several minutes..."
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(APP) -ExitOnFail 25 5 TestAppImportExport

test-sim-multi-seed-short: runsim
	@echo "Running short multi-seed application simulation. This may take awhile!"
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(APP) -ExitOnFail 25 5 TestFullAppSimulation

test-sim-after-import: runsim
	@echo "Running application simulation-after-import. This may take several minutes..."
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(APP) -ExitOnFail 25 5 TestAppSimulationAfterImport

test-sim-deterministic: runsim
	@echo "Running application deterministic simulation. This may take awhile!"
	@$(BINDIR)/runsim -Jobs=4 -SimAppPkg=$(APP) -ExitOnFail 1 1 TestAppStateDeterminism

.PHONY: proto-all proto-gen proto-swagger-gen proto-format proto-lint proto-check-breaking mocks \
	test-sim-import-export test-sim-multi-seed-short test-sim-after-import test-sim-deterministic
