FROM golang:1.13-stretch

ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /go/cache

ADD go.mod .
RUN go mod download

WORKDIR /app

ADD . .

ARG build_tags

RUN if [ ! -n $build_tags ]; then go build -tags $build_tags -o ./dist/ropee ; else go build -o ./dist/ropee ; fi

FROM alpine:3.8

COPY --from=0 /app/dist/ropee /usr/local/bin
COPY entrypoint.sh /usr/local/bin

RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT /usr/local/bin/entrypoint.sh
