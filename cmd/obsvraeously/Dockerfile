FROM golang:1.16 as builder

RUN mkdir /opt/obsvraeously
WORKDIR /opt/obsvraeously

COPY . /opt/obsvraeously

ENV GO111MODULE on
RUN go build ./cmd/obsvraeously/obsvraeously.go

FROM debian:stretch-slim

RUN apt-get update
RUN apt-get install -y ca-certificates

COPY --from=builder /opt/obsvraeously/obsvraeously /usr/local/bin/obsvraeously

EXPOSE 8080

ENTRYPOINT /usr/local/bin/obsvraeously