FROM golang:1.26-alpine AS build

COPY . /go/src/github.com/utilitywarehouse/sql-backup
WORKDIR /go/src/github.com/utilitywarehouse/sql-backup

ENV GOLANGCI_LINT_VERSION="v1.59.1"

# GOTOOLCHAIN pins the exact toolchain declared by go.mod's `go` line, so the
# build isn't at the mercy of whatever patch version the base image ships.
RUN apk --no-cache add make build-base git ca-certificates && \
  GOTOOLCHAIN=go$(awk '/^go /{print $2; exit}' go.mod) && \
  go mod download && \
  go test -v -cover -p=1 ./... && \
  wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s ${GOLANGCI_LINT_VERSION} && \
  ./bin/golangci-lint run && \
  go build -o /sql-backup \
    -ldflags "\
      -s \
      -X main.gitSummary=$(git describe --tags --dirty --always) \
      -X main.gitBranch=$(git rev-parse --abbrev-ref HEAD) \
      -X main.buildStamp=$(date -u '+%Y-%m-%dT%H:%M:%S%z')" \
    ./cmd/sql-backup

FROM alpine:3.24

RUN apk add --no-cache ca-certificates postgresql
COPY --from=build /sql-backup /sql-backup

ENTRYPOINT ["/sql-backup"]
