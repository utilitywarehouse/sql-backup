# get name of directory containing this Makefile
# (stolen from https://stackoverflow.com/a/18137056)
mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
base_dir := $(notdir $(patsubst %/,%,$(dir $(mkfile_path))))

SERVICE ?= $(base_dir)

DOCKER_REGISTRY=registry.uw.systems
DOCKER_REPOSITORY_NAMESPACE=btg-build
DOCKER_REPOSITORY_IMAGE=db-backup
DOCKER_REPOSITORY=$(DOCKER_REGISTRY)/$(DOCKER_REPOSITORY_NAMESPACE)/$(DOCKER_REPOSITORY_IMAGE)

BUILDENV :=
BUILDENV += CGO_ENABLED=0
GIT_HASH := $(shell git rev-parse --short HEAD)
LINKFLAGS := -s -w -X main.gitHash=$(GIT_HASH) -extldflags "-static"
LINTFLAGS := --timeout=10m
TESTFLAGS := -v -cover
LINTER := golangci-lint

EMPTY :=
SPACE := $(EMPTY) $(EMPTY)
join-with = $(subst $(SPACE),$1,$(strip $2))

LEXC :=

.PHONY: install
install:
	@GO111MODULE=on GOPRIVATE="github.com/utilitywarehouse/*" go mod download

$(LINTER):
	@[ -e ./bin/$(LINTER) ] || wget -O - -q https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s latest

.PHONY: lint
lint: $(LINTER)
	@./bin/$(LINTER) run $(LINTFLAGS)


.PHONY: clean
clean:
	@rm -f $(SERVICE)

# builds our binary
$(SERVICE): clean
	@GO111MODULE=on $(BUILDENV) go build -o $(SERVICE) -a -ldflags '$(LINKFLAGS)' ./cmd/$(SERVICE)

build: $(SERVICE)

.PHONY: verify
verify:
	@GO111MODULE=on $(BUILDENV) go mod verify

.PHONY: test
test: verify
	@GO111MODULE=on $(BUILDENV) go test $(TESTFLAGS) ./...

.PHONY: all
all: clean $(LINTER) lint test build

docker-image:
	@docker buildx build --platform linux/amd64 -t $(DOCKER_REPOSITORY):$(GIT_HASH) . --build-arg SERVICE=$(SERVICE) --build-arg GITHUB_TOKEN=$(GITHUB_TOKEN)

ci-docker-auth:
	@echo "Logging in to $(DOCKER_REGISTRY) as $(DOCKER_USERNAME)"
	@docker login -u $(DOCKER_USERNAME) -p $(DOCKER_PASSWORD) $(DOCKER_REGISTRY)

ci-docker-build: ci-docker-auth
	@docker buildx build --platform linux/amd64 --load -t $(DOCKER_REPOSITORY):$(GIT_HASH) . --build-arg SERVICE=$(SERVICE) --build-arg GITHUB_TOKEN=$(GITHUB_TOKEN)
	@docker tag $(DOCKER_REPOSITORY):$(GIT_HASH) $(DOCKER_REPOSITORY):latest
	@docker push $(DOCKER_REPOSITORY)