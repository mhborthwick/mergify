.PHONY: create build

create:
	go run cmd/mergify.go create

build:
	go build -o bin/mergify cmd/mergify.go
