.PHONY: run test ensure fmt

run: ensure fmt test
	go install ./... && rabisco-server

test: ensure fmt
	go test ./...

ensure:
	dep ensure

fmt:
	go fmt ./...
