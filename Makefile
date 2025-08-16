build-players:
	docker build .                     \
		-t players:latest              \
		--build-arg CGO_ENABLED=0      \
		--build-arg GO_SERVICE=players \
		--build-arg BUILD_TYPE=default

build-auth:
	docker build .                     \
		-t auth:latest                 \
		--build-arg CGO_ENABLED=0      \
		--build-arg GO_SERVICE=auth    \
		--build-arg BUILD_TYPE=default

build-all: build-players build-auth

PROTO_FILES := $(shell find api -name "*.proto")
GOLANG_VERSION ?= 1.24

gen_proto:
	docker run -t --rm                                            \
    	        -v $(CURDIR):/app                                 \
    	        -w /app                                           \
    	        bufbuild/buf generate --template buf.gen.yaml api


golangci-lint-fix:
	docker run -t --rm -v .:/app -w /app golangci/golangci-lint:v2.4.0-alpine golangci-lint run --fix

golangci-lint:
	docker run -t --rm -v .:/app -w /app golangci/golangci-lint:v2.4.0-alpine golangci-lint run

test:
	docker run --rm -v $(CURDIR):/app -w /app golang:$(GOLANG_VERSION) go test ./...
