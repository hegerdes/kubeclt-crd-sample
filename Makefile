BINARY := kubectl-crd-sample
PKG    := ./...
GOBIN  ?= $(shell go env GOPATH)/bin

.PHONY: all build test tidy install clean

all: build

tidy:
	go mod tidy

build: tidy
	CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o $(BINARY) .

test:
	go test -race -count=1 $(PKG)

install: build
	install -m 0755 $(BINARY) $(GOBIN)/$(BINARY)
	@echo "Installed $(BINARY) -> $(GOBIN)/$(BINARY)"
	@echo "Make sure $(GOBIN) is on your PATH so kubectl can discover the plugin."

clean:
	rm -f $(BINARY)
