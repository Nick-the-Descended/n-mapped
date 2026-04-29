.PHONY: help build run dev test fmt vet tidy clean install frontend frontend-dev

BINARY      ?= n-mapped
PREFIX      ?= /usr/local
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS     := -s -w -X main.Version=$(VERSION)
GO          ?= go

# When the real frontend is wired up, `make build` will run `make frontend` first.
# Until then, the placeholder index.html in internal/server/frontend_dist/ is used.

help:
	@echo "Targets:"
	@echo "  build         Build the n-mapped binary"
	@echo "  run           Build and run (passes ARGS, e.g. make run ARGS='--privileged')"
	@echo "  dev           Run with -race for development"
	@echo "  test          go test ./..."
	@echo "  fmt vet       go fmt / go vet"
	@echo "  tidy          go mod tidy"
	@echo "  install       Install to \$$PREFIX (binary + desktop + icon + man)"
	@echo "  uninstall     Remove the files installed by 'install'"
	@echo "  frontend      Rebuild the Svelte SPA + copy into the embed dir"
	@echo "  frontend-dev  Run the Vite dev server (HMR) — proxies /api to :8765"
	@echo "  release       goreleaser release --snapshot --clean"
	@echo "  clean         Remove build artifacts"

build:
	$(GO) build -trimpath -ldflags '$(LDFLAGS)' -o $(BINARY) .

run: build
	./$(BINARY) $(ARGS)

dev:
	$(GO) run -race . $(ARGS)

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

install: build
	install -d $(PREFIX)/bin
	install -m 0755 $(BINARY) $(PREFIX)/bin/$(BINARY)
	install -d $(PREFIX)/share/applications
	install -m 0644 packaging/n-mapped.desktop $(PREFIX)/share/applications/n-mapped.desktop
	install -d $(PREFIX)/share/icons/hicolor/scalable/apps
	install -m 0644 packaging/icon.svg $(PREFIX)/share/icons/hicolor/scalable/apps/n-mapped.svg
	install -d $(PREFIX)/share/man/man1
	install -m 0644 packaging/n-mapped.1 $(PREFIX)/share/man/man1/n-mapped.1
	@echo
	@echo "  Installed to $(PREFIX)/bin/$(BINARY)"
	@echo "  Run 'n-mapped' to launch (or 'sudo n-mapped' for raw-socket scans)."

clean:
	rm -f $(BINARY)
	rm -rf frontend/dist

# --- Frontend (wired up in Phase 1) ---

frontend:
	@if [ -d frontend ]; then \
		cd frontend && npm ci && npm run build && \
		rm -rf ../internal/server/frontend_dist && \
		cp -r dist ../internal/server/frontend_dist; \
	else \
		echo "frontend/ not yet scaffolded; using placeholder in internal/server/frontend_dist/"; \
	fi

frontend-dev:
	@if [ -d frontend ]; then cd frontend && npm run dev; else echo "frontend/ not yet scaffolded"; fi

uninstall:
	rm -f $(PREFIX)/bin/$(BINARY)
	rm -f $(PREFIX)/share/applications/n-mapped.desktop
	rm -f $(PREFIX)/share/icons/hicolor/scalable/apps/n-mapped.svg
	rm -f $(PREFIX)/share/man/man1/n-mapped.1

release:
	goreleaser release --snapshot --clean
