.PHONY: auth-build auth-up auth-down create build install test

auth-build:
	make -C auth build

auth-up:
	make -C auth up

auth-down:
	make -C auth down

create:
	go run cmd/mergify.go create

build:
	go build -o bin/mergify cmd/mergify.go

install:
	go install ./cmd/mergify.go

test:
	go test -cover ./pkg/...