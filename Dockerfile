FROM golang:1-alpine AS build

RUN apk update && apk add make git ca-certificates gcc

ARG GITHUB_TOKEN
ARG SERVICE
ARG APP

RUN git config --global url."https://${GITHUB_TOKEN}:x-oauth-basic@github.com/".insteadOf "https://github.com/"


ADD . /build/${SERVICE}

WORKDIR /build/${SERVICE}


RUN wget -qO- https://binaries.cockroachdb.com/cockroach-v2.1.6.linux-musl-amd64.tgz | tar  xvz
RUN cp -i cockroach-v19.1.0.linux-musl-amd64/cockroach /
RUN chmod +x /cockroach

RUN make install
RUN make build
RUN mv ./db-backup /db-backup

FROM alpine:latest

RUN apk add --no-cache ca-certificates postgresql
COPY --from=build /cockroach /usr/local/bin/cockroach
COPY --from=build /db-backup /db-backup

ENTRYPOINT ["/db-backup"]
