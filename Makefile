.PHONY: ckeck install upload

BUILD_VERSION ?= $(shell cat VERSION)
BUILD_OUTPUT ?= fun-telegram
GO ?= GO111MODULE=on CGO_ENABLED=1 go
GOOS ?= $(shell $(GO) version | cut -d' ' -f4 | cut -d'/' -f1)
GOARCH ?= $(shell $(GO) version | cut -d' ' -f4 | cut -d'/' -f2)
DOCKER_IMAGE ?= ghcr.io/teadove/fun-telegram:$(BUILD_VERSION)

upload:
	docker login ghcr.io
	docker buildx build --platform linux/arm64,linux/amd64 . --tag $(DOCKER_IMAGE) --no-cache --push

run:
	@$(GO) run main.go

check:
	pre-commit run -a
	@$(GO) test -v ./...

build:
	@$(GO) build -o $(BUILD_OUTPUT) main.go

clean:
	@echo -n ">> CLEAN"
	@$(GO) clean -i ./...
	@rm -f goteleout-*-*
	@rm -rf dist/*
	@printf '%s\n' '$(OK)'


crosscompile:
	@echo -n ">> CROSSCOMPILE linux/amd64"
	@GOOS=linux GOARCH=amd64 $(GO) build -o $(BUILD_OUTPUT)-$(BUILD_VERSION)-linux-amd64
	@printf '%s\n' '$(OK)'
	@echo -n ">> CROSSCOMPILE darwin/amd64"
	@GOOS=darwin GOARCH=amd64 $(GO) build -o $(BUILD_OUTPUT)-$(BUILD_VERSION)-darwin-amd64
	@printf '%s\n' '$(OK)'
	@echo -n ">> CROSSCOMPILE windows/amd64"
	@GOOS=windows GOARCH=amd64 $(GO) build -o $(BUILD_OUTPUT)-$(BUILD_VERSION)-windows-amd64
	@printf '%s\n' '$(OK)'

	@echo -n ">> CROSSCOMPILE linux/arm64"
	@GOOS=linux GOARCH=arm64 $(GO) build -o $(BUILD_OUTPUT)-$(BUILD_VERSION)-linux-arm64
	@printf '%s\n' '$(OK)'
	@echo -n ">> CROSSCOMPILE darwin/arm64"
	@GOOS=darwin GOARCH=arm64 $(GO) build -o $(BUILD_OUTPUT)-$(BUILD_VERSION)-darwin-arm64
	@printf '%s\n' '$(OK)'
	@echo -n ">> CROSSCOMPILE windows/arm64"
	@GOOS=windows GOARCH=arm64 $(GO) build -o $(BUILD_OUTPUT)-$(BUILD_VERSION)-windows-arm64
	@printf '%s\n' '$(OK)'
