FROM golang:1.13 AS build-env

# Add namespace here to resolve /vendor dependencies
ENV NAMESPACE gitlab.com/bolinda/bolindalabs/ops/token-gen
WORKDIR /go/src/$NAMESPACE

ADD . ./

ARG opts
RUN GO111MODULE=off CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -ldflags "-w -s"  -a -installsuffix cgo -o /out/token-gen *.go


FROM gcr.io/distroless/static:latest

COPY --from=build-env /out /
