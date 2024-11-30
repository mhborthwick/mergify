FROM golang:1.23-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/main.go ./cmd/main.go
RUN CGO_ENABLED=0 go build -o /bin/entry ./cmd/main.go

FROM alpine:latest

WORKDIR /home/auth_user

RUN addgroup -S spotify && \
  adduser -S auth_user -G spotify && \
  mkdir -p /home/auth_user/.mergify && \
  chown -R auth_user:spotify /home/auth_user && \
  chmod -R 700 /home/auth_user/.mergify

USER auth_user

COPY --from=build /bin/entry /home/auth_user/entry

ENTRYPOINT ["/home/auth_user/entry"]
