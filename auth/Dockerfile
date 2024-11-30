FROM golang:1.23-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/server.go ./cmd/server.go
RUN CGO_ENABLED=0 go build -o /bin/server ./cmd/server.go

FROM alpine:latest

WORKDIR /home/usr

RUN addgroup -S spotify && \
  adduser -S usr -G spotify && \
  mkdir -p /home/usr/.mergify && \
  chown -R usr:spotify /home/usr && \
  chmod -R 700 /home/usr/.mergify

USER usr

COPY --from=build /bin/server /home/usr/server

ENTRYPOINT ["/home/usr/server"]
