update:
	go mod tidy
	go mod vendor

build:
	go build -v ./...

test:
	go test -v ./...

cover:
	go test -c	over -v ./...