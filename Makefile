mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
base_dir := $(notdir $(patsubst %/,%,$(dir $(mkfile_path))))

SERVICE ?= $(base_dir)
DOCKER_REGISTRY=registry.uw.systems
DOCKER_ID=telco
DOCKER_REPOSITORY_IMAGE=$(SERVICE)
DOCKER_REPOSITORY=$(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY_IMAGE)

BUILDENV :=
BUILDENV += CGO_ENABLED=0
GIT_SUMMARY := $(shell git describe --tags --dirty --always)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GIT_HASH := $(CIRCLE_SHA1)
ifeq ($(GIT_HASH),)
  GIT_HASH := $(shell git rev-parse HEAD)
endif
LINKFLAGS :=-s -X main.gitSummary=$(GIT_SUMMARY) -X main.gitBranch=$(GIT_BRANCH) -X main.buildStamp=$(shell date -u '+%Y-%m-%dT%H:%M:%S%z') -extldflags "-static"
TESTFLAGS := -v -cover -p=1
LINT_FLAGS :=--disable-all --enable=vet --enable=vetshadow --enable=golint --enable=ineffassign --enable=goconst --enable=gofmt
LINTER_EXE := gometalinter.v2
LINTER := $(GOPATH)/bin/$(LINTER_EXE)

EMPTY :=
SPACE := $(EMPTY) $(EMPTY)
join-with = $(subst $(SPACE),$1,$(strip $2))

LINT_EXCLUDE=pb
LEXC :=
ifdef LINT_EXCLUDE
	LEXC := $(call join-with,|,$(LINT_EXCLUDE))
endif

.PHONY: install
install:
	go get -v -d ./... 2>&1 | sed -e "s/[[:alnum:]]*:x-oauth-basic/redacted/"

$(LINTER):
	GO111MODULE=off  go get -u gopkg.in/alecthomas/$(LINTER_EXE)
	$(LINTER) --install

.PHONY: lint
lint: $(LINTER)
ifdef LEXC
	$(LINTER) --exclude '$(LEXC)' $(LINT_FLAGS) ./...
else
	$(LINTER) $(LINT_FLAGS) ./...
endif

.PHONY: clean
clean:
	rm -f $(SERVICE)

# builds our binary + code generation
$(SERVICE):
	$(BUILDENV) go build -o $(SERVICE) -a -ldflags '$(LINKFLAGS)' ./cmd/$(SERVICE)

build: $(SERVICE)
	
.PHONY: test
test:
	$(BUILDENV) go test $(TESTFLAGS) ./...

.PHONY: all
all: clean $(LINTER) lint test build build-proxy

docker-image:
	docker build -t $(DOCKER_REPOSITORY):local . --build-arg SERVICE=$(SERVICE) --build-arg GITHUB_TOKEN=$(GITHUB_TOKEN)

ci-docker-auth:
	@echo "Logging in to $(DOCKER_REGISTRY) as $(DOCKER_ID)"
	@docker login -u $(DOCKER_ID) -p $(DOCKER_PASSWORD) $(DOCKER_REGISTRY)

ci-docker-build: ci-docker-auth
	docker build -t $(DOCKER_REPOSITORY):$(CIRCLE_SHA1) . --build-arg APP=$(SERVICE) --build-arg SERVICE=$(SERVICE) --build-arg GITHUB_TOKEN=$(GITHUB_TOKEN)
	docker tag $(DOCKER_REPOSITORY):$(CIRCLE_SHA1) $(DOCKER_REPOSITORY):latest
	docker push $(DOCKER_REPOSITORY)


ci-docker-build-proxy: ci-docker-auth
	docker build -t $(DOCKER_REPOSITORY_PROXY):$(CIRCLE_SHA1) . --build-arg APP=$(SERVICE_PROXY) --build-arg SERVICE=$(SERVICE) --build-arg GITHUB_TOKEN=$(GITHUB_TOKEN)
	docker tag $(DOCKER_REPOSITORY_PROXY):$(CIRCLE_SHA1) $(DOCKER_REPOSITORY_PROXY):latest
	docker push $(DOCKER_REPOSITORY_PROXY)
