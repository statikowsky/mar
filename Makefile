BINARY := mar
PKG := github.com/statikowsky/mar
BIN_DIR := bin
DIST_DIR := dist
INSTALL_DIR := $(shell go env GOPATH)/bin

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X $(PKG)/internal/version.Version=$(VERSION)

# OS/arch targets for `make release`, as GOOS/GOARCH pairs.
RELEASE_TARGETS := darwin/arm64 darwin/amd64 linux/amd64 linux/arm64 windows/amd64

.DEFAULT_GOAL := build

.PHONY: build
build:
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) .

.PHONY: install
install:
	go install -ldflags "$(LDFLAGS)" .

.PHONY: run
run: build
	./$(BIN_DIR)/$(BINARY) $(ARGS)

.PHONY: test
test:
	go test ./...

.PHONY: test-race
test-race:
	go test -race ./...

.PHONY: cover
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: check
check: fmt vet test-race

# release cross-compiles every RELEASE_TARGETS into $(DIST_DIR)/, archives each
# (tar.gz for unix, zip for windows) with README/LICENSE, and writes checksums.
# Pure Go (no cgo) means CGO_ENABLED=0 cross-compiles need no C toolchain.
# Override VERSION for a real release: make release VERSION=v0.1.0
.PHONY: release
release: clean-dist
	@mkdir -p $(DIST_DIR)
	@for target in $(RELEASE_TARGETS); do \
		goos=$${target%/*}; goarch=$${target#*/}; \
		bin=$(BINARY); ext=""; \
		if [ "$$goos" = "windows" ]; then bin=$(BINARY).exe; ext="zip"; else ext="tar.gz"; fi; \
		stage=$(DIST_DIR)/$(BINARY)_$${goos}_$${goarch}; \
		archive=$$stage.$$ext; \
		echo "building $$goos/$$goarch -> $$archive"; \
		mkdir -p $$stage; \
		CGO_ENABLED=0 GOOS=$$goos GOARCH=$$goarch \
			go build -ldflags "$(LDFLAGS)" -o $$stage/$$bin . || exit 1; \
		[ -f README.md ] && cp README.md $$stage/ || true; \
		[ -f LICENSE ] && cp LICENSE $$stage/ || true; \
		if [ "$$ext" = "zip" ]; then \
			( cd $$stage && zip -qr ../$$(basename $$archive) . ); \
		else \
			tar -czf $$archive -C $$stage .; \
		fi; \
		rm -rf $$stage; \
	done
	@cd $(DIST_DIR) && shasum -a 256 *.tar.gz *.zip > checksums.txt 2>/dev/null || \
		( cd $(DIST_DIR) && sha256sum *.tar.gz *.zip > checksums.txt )
	@echo "release artifacts in $(DIST_DIR)/ (version $(VERSION))"

.PHONY: clean-dist
clean-dist:
	rm -rf $(DIST_DIR)

.PHONY: clean
clean: clean-dist
	rm -rf $(BIN_DIR) coverage.out

.PHONY: help
help:
	@echo "Targets:"
	@echo "  build      Build the $(BINARY) binary into $(BIN_DIR)/"
	@echo "  install    Install $(BINARY) to $(INSTALL_DIR)"
	@echo "  run        Build and run (pass args via ARGS=\"...\")"
	@echo "  test       Run all tests"
	@echo "  test-race  Run all tests with the race detector"
	@echo "  cover      Generate and open an HTML coverage report"
	@echo "  fmt        Format all Go code"
	@echo "  vet        Run go vet"
	@echo "  tidy       Tidy go.mod / go.sum"
	@echo "  check      fmt + vet + test-race"
	@echo "  release    Cross-compile all OS/arch into $(DIST_DIR)/ (VERSION=... to override)"
	@echo "  clean      Remove build artifacts"
