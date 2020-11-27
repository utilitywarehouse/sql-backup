FROM golang:1-alpine AS build

COPY . /go/src/github.com/utilitywarehouse/sql-backup
WORKDIR /go/src/github.com/utilitywarehouse/sql-backup

ENV CGO_ENABLED=0
ENV CRDB_VERSION="v20.1.0"
ENV GOLANGCI_LINT_VERSION="v1.33.0"

RUN apk --no-cache add make git ca-certificates && \
  wget -qO- https://binaries.cockroachdb.com/cockroach-${CRDB_VERSION}.linux-musl-amd64.tgz | tar xvz && \
  cp -i cockroach-${CRDB_VERSION}.linux-musl-amd64/cockroach / && \
  chmod +x /cockroach && \
  go get -v -d ./... && \
  go test -v -cover -p=1 ./... && \
  wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s ${GOLANGCI_LINT_VERSION} && \
  ./bin/golangci-lint run --deadline=2m && \
  go build -o /sql-backup \
    -ldflags "\
      -s \
      -X main.gitSummary=$(git describe --tags --dirty --always) \
      -X main.gitBranch=$(git rev-parse --abbrev-ref HEAD) \
      -X main.buildStamp=$(date -u '+%Y-%m-%dT%H:%M:%S%z')" \
    ./cmd/sql-backup

FROM alpine:latest

RUN apk add --no-cache ca-certificates postgresql
COPY --from=build /cockroach /usr/local/bin/cockroach
COPY --from=build /sql-backup /sql-backup

ENTRYPOINT ["/sql-backup"]
