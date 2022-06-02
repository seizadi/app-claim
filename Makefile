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

# Docker buildx options
# Platforms to build the multi-arch image for.
IMAGE_PLATFORMS ?= linux/amd64,linux/arm64

# Base build image to use.
BUILD_BASE_IMAGE ?= golang:1.18.2

# Enable build with CGO.
BUILD_CGO_ENABLED ?= 0

# Go module mirror to use.
BUILD_GOPROXY ?= https://proxy.golang.org

IMAGE_RESULT_FLAG = --output=type=oci,dest=$(shell pwd)/image/claims-$(VERSION).tar

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

#TODO
.PHONY: buildx
buildx: ## Build and optionally push a multi-arch db-controller container image to the Docker registry
	@docker buildx build --push \
		--build-arg "BUILD_GOPROXY=$(BUILD_GOPROXY)" \
		--build-arg api_version=$(API_VERSION) \
		--build-arg srv_version=$(SRV_VERSION) \
		-f $(SERVER_DOCKERFILE) \
		-t $(SERVER_IMAGE):$(IMAGE_VERSION) \
		-t $(SERVER_IMAGE):latest .

#	@mkdir -p $(shell pwd)/image
#	docker buildx build $(IMAGE_RESULT_FLAG) \
#		--platform $(IMAGE_PLATFORMS) \
#		--build-arg "BUILD_GOPROXY=$(BUILD_GOPROXY)" \
#		--build-arg "BUILD_BASE_IMAGE=$(BUILD_BASE_IMAGE)" \
#		--build-arg "BUILD_VERSION=$(BUILD_VERSION)" \
#		--build-arg "BUILD_BRANCH=$(BUILD_BRANCH)" \
#		--build-arg "BUILD_SHA=$(BUILD_SHA)" \
#		--build-arg "BUILD_CGO_ENABLED=$(BUILD_CGO_ENABLED)" \
#		--build-arg "BUILD_EXTRA_GO_LDFLAGS=$(BUILD_EXTRA_GO_LDFLAGS)" \
#		$(DOCKER_BUILD_LABELS) \
#		$(IMAGE_TAGS) \
#		$(shell pwd)

.PHONY: clean
clean:
	@/bin/rm -f ./bin/claims
