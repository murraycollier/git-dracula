PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin

.PHONY: all build test install uninstall clean

all: build

build:
	@mkdir -p bin
	go build -ldflags="-s -w" -o bin/git-dracula ./cmd/git-dracula

test:
	go test -v ./...

install: build
	@mkdir -p $(BINDIR)
	install -m 755 bin/git-dracula $(BINDIR)/git-dracula
	@echo "✓ Installed git-dracula to $(BINDIR)/git-dracula"
	@echo "  You can now run 'git-dracula' or 'git dracula' directly!"

uninstall:
	rm -f $(BINDIR)/git-dracula
	@echo "✓ Removed git-dracula from $(BINDIR)"

clean:
	rm -rf bin/
