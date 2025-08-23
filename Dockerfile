FROM --platform=linux/$TARGETARCH golang:1.24.0-alpine AS build-stage

ARG CGO_ENABLED=0
ARG GO_SERVICE
ARG TARGETARCH
ARG BUILD_TYPE=default
ARG GO_CMD=./services/${GO_SERVICE}/cmd/main.go
RUN echo $TARGETARCH

WORKDIR /go-game-backend/

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=$CGO_ENABLED \
    GOOS=linux \
    GOARCH=$TARGETARCH \
    go build -o server $GO_CMD

FROM alpine:latest AS release-stage

ARG GO_SERVICE
ARG BUILD_TYPE

WORKDIR /

COPY --from=build-stage /go-game-backend/server ./server
COPY --from=build-stage /go-game-backend/services/${GO_SERVICE}/configs/${BUILD_TYPE}.yaml ./config.yaml

ENTRYPOINT ["./server", "-config", "./config.yaml"]