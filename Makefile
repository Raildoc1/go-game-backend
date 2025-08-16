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

gen_proto:
	mkdir -p gen/
	find api \
	-name "*.proto" \
	-exec protoc \
	--proto_path=api/ \
	--go_out=gen/ \
	--go-grpc_out=.. \
	--go_opt=paths=source_relative {} +

golangci-lint-fix:
	docker run -t --rm -v .:/app -w /app golangci/golangci-lint:v2.4.0-alpine golangci-lint run --fix

golangci-lint:
	docker run -t --rm -v .:/app -w /app golangci/golangci-lint:v2.4.0-alpine golangci-lint run

test:
	go test ./...
