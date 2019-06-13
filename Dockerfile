FROM golang:1.12-stretch

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

WORKDIR /app

COPY --from=0 /app/dist/ropee /app
# todo add entrypoint
CMD /app/ropee
