export GO111MODULE=on
export CGO_ENABLED=0

TAG=0.0.5
BINARY=spark-nanny
BUILD=`date +%FT%T%z`
REPO=bringg

lint:
	@echo "==> linting..."
	@golangci-lint run

build: lint
	@echo "==> building..."
	@go build -ldflags="-s -w -X main.version=${TAG} -X main.buildDate=${BUILD}" -o bin/${BINARY} .

download:
	@echo "==> downloading dependencies..."
	@go mod download -x

install-tools: download
	@echo "==> installing tools from tools.go..."
	@cat tools/tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

build-image:
	@echo "==> building docker image..."
	@docker build -t ${REPO}/${BINARY}:${TAG} .

push-image: build-image
	@echo "==> pushing image to docker hub"
	@docker push ${REPO}/${BINARY}:${TAG}
