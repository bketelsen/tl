# tl Makefile — adapted from aholbreich/adr-tool
BINARY_NAME := tl
INSTALL_DIR ?= $(HOME)/.local/bin
VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo dev)
COMMIT_HASH := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
COUNT := $(shell git rev-list $(VERSION)..HEAD --count 2>/dev/null || echo 0)
LDFLAGS := -X main.version=$(VERSION)-$(COUNT)-$(COMMIT_HASH)

PLATFORMS = linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

.PHONY: build test bdd install rpm get-version changelog bump amend dists clean cleancache \
	binary_linux_amd64 binary_linux_arm64 binary_darwin_amd64 binary_darwin_arm64 \
	binary_windows_amd64 binary_windows_arm64

# Build the local binary with version stamping.
build:
	go fmt ./...
	go mod tidy
	go build -o $(BINARY_NAME) -ldflags "$(LDFLAGS)"

# Run all tests (Go unit tests + godog BDD suite under ./bdd/...).
test:
	go test -v ./...

# Run only the BDD suite.
bdd:
	go test -v ./bdd/...

# Install the binary to $(INSTALL_DIR) (default $(HOME)/bin).
install: test build
	mkdir -p $(INSTALL_DIR)
	install -m 0755 $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed $(BINARY_NAME) to $(INSTALL_DIR)/$(BINARY_NAME)"

# Build a local RPM package in dist/rpm/.
rpm:
	bash .github/scripts/build-rpm.sh

get-version:
	@echo $(VERSION)-$(COUNT)-$(COMMIT_HASH)

# Generate release notes from commits since the previous tag.
# Set CURRENT_TAG=x.y.z in tag-triggered release jobs to compare the previous tag to that tag.
changelog:
	@set -e; \
	current_tag="$(CURRENT_TAG)"; \
	if [ -n "$$current_tag" ]; then \
		if ! git rev-parse -q --verify "refs/tags/$$current_tag^{commit}" >/dev/null; then \
			echo "CURRENT_TAG '$$current_tag' does not exist" >&2; \
			exit 2; \
		fi; \
		if prev=$$(git describe --tags --abbrev=0 "$$current_tag^" 2>/dev/null); then \
			range="$$prev..$$current_tag"; \
		else \
			range="$$current_tag"; \
		fi; \
	else \
		if prev=$$(git describe --tags --abbrev=0 2>/dev/null); then \
			range="$$prev..HEAD"; \
		else \
			range="HEAD"; \
		fi; \
	fi; \
	git log --oneline "$$range" | awk ' \
		BEGIN { print "## What'"'"'s Changed"; print "" } \
		{ \
			msg = $$0; \
			sub(/^[0-9a-f]+[[:space:]]+/, "", msg); \
			lower = tolower(msg); \
			line = "  • " msg "\n"; \
			if (lower ~ /^(feat|feature)(\([^)]+\))?!?:/) features = features line; \
			else if (lower ~ /^fix(\([^)]+\))?!?:/) fixes = fixes line; \
			else if (lower ~ /^docs(\([^)]+\))?!?:/) docs = docs line; \
			else if (lower ~ /^refactor(\([^)]+\))?!?:/) refactors = refactors line; \
			else if (lower ~ /^chores?(\([^)]+\))?!?:/) maintenance = maintenance line; \
			else other = other line; \
		} \
		END { \
			section("### 🚀 Features", features); \
			section("### 🐛 Fixes", fixes); \
			section("### 📖 Documentation", docs); \
			section("### ♻️ Refactoring", refactors); \
			section("### 🔧 Maintenance", maintenance); \
			section("### Other Changes", other); \
			if (!printed) print "No changes since the last release."; \
		} \
		function section(title, body) { \
			if (body != "") { \
				if (printed) print ""; \
				print title; \
				printf "%s", body; \
				printed = 1; \
			} \
		}'

# Cut a release: verify, then tag HEAD with the next semver (computed by svu from
# conventional commits) and push it. GitHub Actions builds archives and publishes
# the GitHub Release on the tag push.
#
# svu always emits a v-prefixed tag (vX.Y.Z), which is REQUIRED for Go modules —
# consumers (e.g. omnius `require github.com/bketelsen/tl vX.Y.Z`) cannot resolve
# a bare `X.Y.Z` tag. This replaced an older `make release VERSION=x.y.z` target
# that produced bare, Go-invisible tags. (Fork note: standardized on the same
# svu-based `make bump` flow used across the other repos.)
bump: ## verify, then tag the next svu version and push it
	@$(MAKE) build
	@$(MAKE) test
	@unformatted=$$(gofmt -l .); \
		if [ -n "$$unformatted" ]; then \
			echo "These files need gofmt:" >&2; echo "$$unformatted" >&2; exit 1; \
		fi
	@go vet ./...
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Working tree is not clean; commit or stash changes before bumping." >&2; \
		exit 1; \
	fi
	@branch=$$(git rev-parse --abbrev-ref HEAD); \
		if [ "$$branch" != "main" ]; then \
			echo "Releases must be cut from main; current branch is $$branch." >&2; \
			exit 1; \
		fi
	@git fetch origin main --tags
	@local_head=$$(git rev-parse HEAD); \
		remote_head=$$(git rev-parse origin/main); \
		if [ "$$local_head" != "$$remote_head" ]; then \
			echo "HEAD ($$local_head) is not pushed to origin/main ($$remote_head). Push main first." >&2; \
			exit 1; \
		fi
	@version=$$(svu next); \
		current=$$(git describe --tags --abbrev=0 2>/dev/null || echo ""); \
		if [ "$$version" = "$$current" ]; then \
			echo "Already at $$current; no conventional commits to release. Nothing to bump." ; \
			exit 0; \
		fi; \
		if git rev-parse -q --verify "refs/tags/$$version" >/dev/null || \
			git ls-remote --exit-code --tags origin "refs/tags/$$version" >/dev/null 2>&1; then \
			echo "Tag $$version already exists (local or origin)." >&2; \
			exit 1; \
		fi; \
		echo "Tagging $$version..."; \
		git tag -a "$$version" -m "tl $$version"; \
		git push origin "$$version"; \
		echo "Pushed tag $$version. GitHub Actions will build archives and publish the GitHub Release."

# Amend the last commit with all current changes and force-push. Use with care.
amend:
	git add .
	git commit --amend --no-edit
	git push --force

# Cross-platform release archives.
binary_linux_amd64:
	mkdir -p build/linux-amd64
	GOOS=linux GOARCH=amd64 go build -o build/linux-amd64/$(BINARY_NAME) -ldflags "$(LDFLAGS)"
	tar -C build/linux-amd64 -czvf $(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)
	rm -rf build/linux-amd64

binary_linux_arm64:
	mkdir -p build/linux-arm64
	GOOS=linux GOARCH=arm64 go build -o build/linux-arm64/$(BINARY_NAME) -ldflags "$(LDFLAGS)"
	tar -C build/linux-arm64 -czvf $(BINARY_NAME)-linux-arm64.tar.gz $(BINARY_NAME)
	rm -rf build/linux-arm64

binary_darwin_amd64:
	mkdir -p build/darwin-amd64
	GOOS=darwin GOARCH=amd64 go build -o build/darwin-amd64/$(BINARY_NAME) -ldflags "$(LDFLAGS)"
	tar -C build/darwin-amd64 -czvf $(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)
	rm -rf build/darwin-amd64

binary_darwin_arm64:
	mkdir -p build/darwin-arm64
	GOOS=darwin GOARCH=arm64 go build -o build/darwin-arm64/$(BINARY_NAME) -ldflags "$(LDFLAGS)"
	tar -C build/darwin-arm64 -czvf $(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)
	rm -rf build/darwin-arm64

binary_windows_amd64:
	mkdir -p build/windows-amd64
	GOOS=windows GOARCH=amd64 go build -o build/windows-amd64/$(BINARY_NAME).exe -ldflags "$(LDFLAGS)"
	zip -j $(BINARY_NAME)-windows-amd64.zip build/windows-amd64/$(BINARY_NAME).exe
	rm -rf build/windows-amd64

binary_windows_arm64:
	mkdir -p build/windows-arm64
	GOOS=windows GOARCH=arm64 go build -o build/windows-arm64/$(BINARY_NAME).exe -ldflags "$(LDFLAGS)"
	zip -j $(BINARY_NAME)-windows-arm64.zip build/windows-arm64/$(BINARY_NAME).exe
	rm -rf build/windows-arm64

dists: binary_linux_amd64 binary_linux_arm64 binary_darwin_amd64 binary_darwin_arm64 binary_windows_amd64 binary_windows_arm64

# Remove local build artifacts.
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME).exe $(BINARY_NAME)-*
	rm -rf build/

# Also clear Go build and module caches (slow; rarely needed).
cleancache: clean
	go clean -cache -testcache -modcache
