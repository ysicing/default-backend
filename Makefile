VERSION_PKG := github.com/ergoapi/util/version
ROOT_DIR := $(CURDIR)
BUILD_DIR := $(ROOT_DIR)/_output
BIN_DIR := $(BUILD_DIR)/bin

BUILD_VERSION   ?= $(shell cat VERSION || echo "0.0.0")
BUILD_DATE := $(shell date "+%Y%m%d")
GIT_COMMIT := $(shell git rev-parse --short HEAD || echo "unknown")
GIT_BRANCH := $(shell git branch -r --contains | head -1 | sed -E -e "s%(HEAD ->|origin|upstream)/?%%g" | xargs || echo "unknown")
# GIT_VERSION := $(shell git describe --always --tags --abbrev=14 $(GIT_COMMIT)^{commit})
APP_VERSION := ${BUILD_VERSION}-${BUILD_DATE}-${GIT_BRANCH}-${GIT_COMMIT}
IMAGE           ?= ysicing/defaultbackend

LDFLAGS := "-w \
	-X $(VERSION_PKG).gitVersion=$(APP_VERSION) \
	-X $(VERSION_PKG).gitCommit=$(GIT_COMMIT) \
	-X $(VERSION_PKG).gitBranch=$(GIT_BRANCH) \
	-X $(VERSION_PKG).buildDate=$(BUILD_DATE) \
	-X $(VERSION_PKG).gitTreeState=core \
	-X $(VERSION_PKG).gitMajor=0 \
	-X $(VERSION_PKG).gitMinor=5"

help: ## this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

fmt: ## fmt code
	gofmt -s -w .
	goimports -w .
	@echo gofmt -l
	@OUTPUT=`gofmt -l . 2>&1`; \
	if [ "$$OUTPUT" ]; then \
		echo "gofmt must be run on the following files:"; \
        echo "$$OUTPUT"; \
        exit 1; \
    fi

lint: ## lint code
	@echo golangci-lint run --skip-files \".*test.go\" -v ./...
	@OUTPUT=`command -v golangci-lint >/dev/null 2>&1 && golangci-lint run --skip-files ".*test.go"  -v ./... 2>&1`; \
	if [ "$$OUTPUT" ]; then \
		echo "go lint errors:"; \
		echo "$$OUTPUT"; \
	fi

default: fmt lint ## fmt code

build: ## 构建二进制
	@echo "build bin ${APP_VERSION}"
	@CGO_ENABLED=1 GOARCH=amd64 go build -o dist/defaultbackend \
    				-ldflags ${LDFLAGS} cmd/default-backend.go

docker: ## 构建镜像
	docker buildx build --push --pull -t ${IMAGE}:${BUILD_VERSION} .

.EXPORT_ALL_VARIABLES:

GO111MODULE = on
GOPROXY = https://goproxy.cn,direct
GOSUMDB = sum.golang.google.cn
