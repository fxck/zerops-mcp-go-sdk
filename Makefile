.PHONY: help

define helpMessage
possible targets:
- test
- lint
- clean
- all
- windows-amd
- linux-amd
- linux-i386
- darwin-amd
- darwin-arm
endef
export helpMessage

help:
	@echo "$$helpMessage"

test:
	go test -v ./cmd/... ./internal/...

lint:
	GOOS=darwin GOARCH=arm64 golangci-lint run ./cmd/... ./internal/... --verbose
	GOOS=linux GOARCH=amd64 golangci-lint run ./cmd/... ./internal/... --verbose
	GOOS=windows GOARCH=amd64 golangci-lint run ./cmd/... ./internal/... --verbose

clean:
	rm -rf bin/ releases/

#########
# BUILD #
#########
all: windows-amd linux-amd linux-i386 darwin-amd darwin-arm

windows-amd:
	GOOS=windows GOARCH=amd64 tools/build.sh zerops-mcp-win-x64.exe

linux-amd:
	GOOS=linux GOARCH=amd64 tools/build.sh zerops-mcp-linux-amd64

linux-i386:
	GOOS=linux GOARCH=386 tools/build.sh zerops-mcp-linux-i386

darwin-amd:
	GOOS=darwin GOARCH=amd64 tools/build.sh zerops-mcp-darwin-amd64

darwin-arm:
	GOOS=darwin GOARCH=arm64 tools/build.sh zerops-mcp-darwin-arm64