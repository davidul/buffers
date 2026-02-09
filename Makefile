update:
	go mod tidy
	go mod vendor

build:
	go build -v ./...

test:
	go test -v $(shell go list ./... | grep -v /examples/)

cover:
	go test -cover -v $(shell go list ./... | grep -v /examples/)
