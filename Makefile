SHELL := /bin/bash -o pipefail
VERSION := $(shell git describe --tags --abbrev=0)

fetch:
	go install github.com/modocache/gover@latest
	go install github.com/github-release/github-release@latest

clean:
	rm -f ./jabba
	rm -rf ./build

fmt:
	gofmt -l -s -w `find . -type f -name '*.go' -not -path "./vendor/*"`

test:
	go vet `go list ./... | grep -v /vendor/`
	SRC=`find . -type f -name '*.go' -not -path "./vendor/*"` && gofmt -l -s $$SRC | read && gofmt -l -s -d $$SRC && exit 1 || true
	go test `go list ./... | grep -v /vendor/`

test-coverage:
	go list ./... | grep -v /vendor/ | xargs -L1 -I{} sh -c 'go test -coverprofile `basename {}`.coverprofile {}' && \
	gover && \
	go tool cover -html=gover.coverprofile -o coverage.html && \
	rm *.coverprofile

build:
	CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}"

build-release:
	mkdir -p release
	@echo "Building for multiple platforms..."
	GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" -o release/jabba-${VERSION}-windows-386.exe
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" -o release/jabba-${VERSION}-windows-amd64.exe
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" -o release/jabba-${VERSION}-linux-386
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" -o release/jabba-${VERSION}-linux-amd64
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" -o release/jabba-${VERSION}-darwin-amd64
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" -o release/jabba-${VERSION}-darwin-arm64
	GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" -o release/jabba-${VERSION}-linux-arm
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=${VERSION}" -o release/jabba-${VERSION}-linux-arm64
	@echo "Build complete!"
	@ls -lh release/

install: build
	JABBA_MAKE_INSTALL=true JABBA_VERSION=${VERSION} bash install.sh

publish: clean build-release
	test -n "$(GITHUB_TOKEN)" # $$GITHUB_TOKEN must be set
	github-release release --user Jabba-Team --repo jabba --tag ${VERSION} \
	--name "${VERSION}" --description "${VERSION}" && \
	for file in release/*; do \
		filename=$$(basename $$file); \
		github-release upload --user Jabba-Team --repo jabba --tag ${VERSION} \
		--name "$$filename" --file "$$file"; \
	done
