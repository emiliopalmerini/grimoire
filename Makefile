.PHONY: fmt build run test clean

fmt:
	go fmt ./...

build: fmt
	go build -o grimoire .

run: build
	./grimoire

test: fmt
	go test ./...

clean:
	rm -f grimoire
