export GO111MODULE=on
export CGO_ENABLED=0

BINARY			=spark-nanny
VERSION			?=$(shell git rev-parse --abbrev-ref HEAD)
BUILD				?=$(shell date +%FT%T)
GIT_COMMIT 	?=$(shell git rev-parse HEAD)

lint:
	@echo "==> linting..."
	@golangci-lint run

build: lint
	@echo "==> building..."
	@go build -ldflags="-s -w \
							-X main.version=${VERSION} \
							-X main.buildDate=${BUILD} \
							-X main.commit=${GIT_COMMIT}" \
						-o ./bin/${BINARY} .

download:
	@echo "==> downloading dependencies..."
	@go mod download -x

install-tools: download
	@echo "==> installing tools from tools.go..."
	@cat tools/tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

helm-docs:
	@helm-docs -c ./charts
