ARG BUILDPLATFORM=linux/amd64
ARG BUILD_BASE_IMAGE

FROM --platform=$BUILDPLATFORM $BUILD_BASE_IMAGE AS build
LABEL stage=server-intermediate
WORKDIR /go/src/github.com/seizadi/app-claim

ARG TARGETOS
ARG TARGETARCH

ARG BUILD_GOPROXY
ENV GOPROXY=${BUILD_GOPROXY}
COPY go.mod go.sum /go/src/github.com/seizadi/app-claim/
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    go build -o bin/claims ./cmd/claims

# copy the server binary from builder stage; run the server binary
FROM alpine:latest AS runner

WORKDIR /bin
COPY --from=build /go/src/github.com/seizadi/app-claim/bin/claims .

ENTRYPOINT ["claims"]
