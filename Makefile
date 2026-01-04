.PHONY: all fmt vet build run test clean

all: fmt vet test build

fmt:
	go fmt ./...

vet: fmt
	go vet ./...

build: vet
	go build -o grimorio .

run: build
	./grimorio

test: vet
	go test -v ./...

clean:
	rm -f grimorio
	go clean ./...
