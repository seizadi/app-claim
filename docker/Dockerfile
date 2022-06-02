# build the binary
FROM golang:1.18.2 AS builder
LABEL stage=server-intermediate
WORKDIR /go/src/github.com/seizadi/app-claim

ARG BUILD_GOPROXY
ENV GOPROXY=${BUILD_GOPROXY}
COPY go.mod go.sum /go/src/github.com/seizadi/app-claim/
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/claims ./cmd/claims

# copy the server binary from builder stage; run the server binary
FROM alpine:latest AS runner

WORKDIR /bin
COPY --from=builder /go/src/github.com/seizadi/app-claim/bin/claims .

ENTRYPOINT ["claims"]
