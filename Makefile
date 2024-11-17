.PHONY: create build install

create:
	go run cmd/mergify.go create

build:
	go build -o bin/mergify cmd/mergify.go

install:
	go install ./cmd/mergify.go
