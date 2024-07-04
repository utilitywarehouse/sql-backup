FROM golang:1.22-alpine AS build

COPY . /go/src/github.com/utilitywarehouse/sql-backup
WORKDIR /go/src/github.com/utilitywarehouse/sql-backup

ENV GOLANGCI_LINT_VERSION="v1.59.1"

RUN apk --no-cache add make build-base git ca-certificates && \
  go get -v -d ./... && \
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

FROM alpine:latest

RUN apk add --no-cache ca-certificates postgresql
COPY --from=build /sql-backup /sql-backup

ENTRYPOINT ["/sql-backup"]
