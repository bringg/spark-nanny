export GO111MODULE=on
export CGO_ENABLED=0

BINARY			=spark-nanny
VERSION			?=$(shell git rev-parse --abbrev-ref HEAD)
BUILD				?=$(shell date +%FT%T)
GIT_COMMIT 	?=$(shell git rev-parse HEAD)

.PHONY: lint
lint:
	@echo "==> Linting..."
	@golangci-lint run

.PHONY: build
build: lint
	@echo "==> Building..."
	@go build -ldflags="-s -w \
							-X main.version=${VERSION} \
							-X main.buildDate=${BUILD} \
							-X main.commit=${GIT_COMMIT}" \
						-o ./bin/${BINARY} .

.PHONY: dependencies
dependencies:
	@echo "==> Downloading dependencies..."
	@go mod download -x

.PHONY: upgrade-deps
upgrade-deps:
	@echo "==> Upgrading dependencies..."
	@go get -t -u ./...
	@go mod tidy

.PHONY: tools
tools:
	@echo "==> Installing tools from tools.go..."
	@awk -F'"' '/_/ {print $$2}' tools/tools.go | xargs -tI % go install %

.PHONY: helm-docs
helm-docs:
	@helm-docs -c ./charts
