PROJECT_ROOT    := github.com/seizadi/app-claim
BUILD_PATH      := bin
DOCKERFILE_PATH := $(CURDIR)/docker

# configuration for image names
USERNAME       := $(USER)
GIT_COMMIT     := $(shell git describe --dirty=-unsupported --always || echo pre-commit)
IMAGE_VERSION  ?= $(USERNAME)-dev-$(GIT_COMMIT)
IMAGE_REGISTRY ?= soheileizadi

# configuration for server binary and image
SERVER_BINARY     := $(BUILD_PATH)/claims
SERVER_PATH       := $(PROJECT_ROOT)/cmd/claims
SERVER_IMAGE      := $(IMAGE_REGISTRY)/claims
SERVER_DOCKERFILE := $(DOCKERFILE_PATH)/Dockerfile
# Placeholder. modify as defined conventions.
DB_VERSION        := 3
SRV_VERSION       := $(shell git describe --tags)
API_VERSION       := v1

# configuration for building on host machine
GO_CACHE       := -pkgdir $(BUILD_PATH)/go-cache
GO_BUILD_FLAGS ?= $(GO_CACHE) -i -v
GO_TEST_FLAGS  ?= -v -cover

.PHONY: build
build:
	@go build -o ./bin/claims ./cmd/claims

.PHONY: docker
docker:
	@docker build --build-arg api_version=$(API_VERSION) --build-arg srv_version=$(SRV_VERSION) -f $(SERVER_DOCKERFILE) -t $(SERVER_IMAGE):$(IMAGE_VERSION) .
	@docker tag $(SERVER_IMAGE):$(IMAGE_VERSION) $(SERVER_IMAGE):latest
	@docker image prune -f --filter label=stage=server-intermediate

.PHONY: push
push:
	@docker push $(SERVER_IMAGE)

.PHONY: clean
clean:
	@/bin/rm -f ./bin/claims
