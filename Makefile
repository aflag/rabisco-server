.PHONY: run test ensure fmt install

run: ensure fmt test install
	rabisco-server

install: ensure fmt test
	go install ./...

test: ensure fmt
	go test ./...

ensure:
	dep ensure

fmt:
	go fmt ./...
