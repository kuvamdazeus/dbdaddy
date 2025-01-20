# needs to be run in a go environment

run:
	go run cmd/main/main.go $(ARGS)

vet:
	go vet cmd/main/main.go $(ARGS)

build:
	go run cmd/build/build.go $(ARGS)

build-dev:
	go build -o bin/dbdaddy cmd/main/main.go

release:
	go run cmd/release/release.go

manrelease:
	go run cmd/manrelease/manrelease.go $(ARGS)

get_version:
	go run cmd/version/version.go
