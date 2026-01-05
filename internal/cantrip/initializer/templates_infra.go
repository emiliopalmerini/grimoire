package initializer

import "fmt"

func ciWorkflowTemplate(name, goVersion string, opts ProjectOptions) string {
	templStep := ""
	if opts.Type == "web" {
		templStep = `
      - name: Install Templ
        run: go install github.com/a-h/templ/cmd/templ@latest

      - name: Generate Templ templates
        run: templ generate ./...
`
	}

	bufStep := ""
	if hasTransport(opts.Transports, "grpc") {
		bufStep = `
      - name: Install Buf
        uses: bufbuild/buf-setup-action@v1

      - name: Generate Proto
        run: buf generate
`
	}

	return fmt.Sprintf(`name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  check-rebase:
    if: github.event_name == 'pull_request'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Check if branch is rebased on main
        run: |
          git fetch origin main
          merge_base=$(git merge-base HEAD origin/main)
          main_head=$(git rev-parse origin/main)
          if [ "$merge_base" = "$main_head" ]; then
            echo "Branch is properly rebased on main"
          else
            echo "Branch is NOT rebased on main"
            echo "To rebase: git rebase origin/main"
            exit 1
          fi

  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: "%s"
%s%s
      - name: Check formatting
        run: |
          if [ -n "$(gofmt -l .)" ]; then
            echo "The following files are not formatted:"
            gofmt -l .
            exit 1
          fi

      - name: Run go vet
        run: go vet ./...

      - name: Build
        run: go build -o %s .

      - name: Run tests
        run: go test ./...
`, goVersion, templStep, bufStep, name)
}

func rabbitmqTemplate() string {
	return `package infra

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func NewRabbitMQ(url string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	return conn, nil
}
`
}
