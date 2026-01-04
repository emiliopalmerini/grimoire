.PHONY: fmt build run test clean

fmt:
	go fmt ./...

build: fmt
	go build -o grimorio .

run: build
	./grimorio

test: fmt
	go test ./...

clean:
	rm -f grimorio
