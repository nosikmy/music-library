FROM golang:1.23-alpine3.19

ENV GOPATH=/
RUN go env -w GOCACHE=/.cache

COPY ./ ./

RUN --mount=type=cache,target=/.cache go build -v -o music-library ./cmd/music-library

ENTRYPOINT ./music-library