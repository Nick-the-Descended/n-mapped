# n-mapped

**A friendly local GUI for [Nmap](https://nmap.org).** Click and search your way through hundreds of flags and NSE scripts, save your favorite scans, watch live progress, browse parsed results, and learn nmap from inline plain-English descriptions — without memorizing CLI syntax.

`n-mapped` is a single static binary that embeds a Svelte SPA. It launches an HTTP server on `127.0.0.1`, opens your browser, and drives the `nmap` already installed on your machine.

```text
┌──────────────────────────────────────────────────────┐
│  Browser tab @ http://127.0.0.1:8765                 │
│   builder · scripts · live results · history · diff  │
└────────────┬───────────────────────────┬─────────────┘
       REST  │                       SSE │
┌────────────▼───────────────────────────▼─────────────┐
│  n-mapped (single Go binary, ~10 MB, no runtime)     │
└────────────────────────────┬─────────────────────────┘
                             │ os/exec (argv only, never a shell)
                       ┌─────▼─────┐
                       │   nmap    │  ← the system binary you already have
                       └───────────┘
```

---

## Why it exists

Nmap is one of the most powerful tools in networking, but its CLI assumes you've already read [the reference guide](https://nmap.org/book/). Beginners hit a wall — the difference between `-sS`, `-sT`, and `-sU` is not obvious from `--help`, and intrusive scripts can have real consequences if you don't know what they do. `n-mapped` keeps the speed and full power of nmap while giving every flag a short description, a longer explanation, examples, and a clear warning when it's noisy or destructive.

It's aimed at:

- **Beginners** who want to learn nmap by clicking around safely.
- **Sysadmins / homelab folks** who run scans rarely and forget the syntax in between.
- **Security learners** doing CTFs or HTB-style boxes who want a reproducible record of what they ran.

It is **not** an offensive-toolkit replacement; we're a builder + viewer, not a workflow engine.

---

## Features at a glance

- **Catalog-driven builder.** Flags grouped by category (Host Discovery, Scan Techniques, Port Spec, Service/Version, OS Detection, Timing, Output) with fuzzy search and a beginner / intermediate / advanced filter.
- **NSE script library.** Curated set of commonly-used scripts with category filters, an inline arg builder, and an explicit confirmation prompt for intrusive / exploit categories.
- **Built-in profiles.** Quick scan, Ping sweep, Service & version, Intense, Vuln (vulners), TLS audit, SMB recon, All TCP. One click to apply.
- **Favorites.** Save the current builder state with a name, re-apply later. Persisted to disk.
- **Live results.** Streaming XML parser turns nmap output into per-host cards with a ports table, OS guess, and per-port script output. Filter "open ports only" or search across all results.
- **History.** Every scan is persisted (one JSON per scan in your data dir) with parsed structure and the original raw XML. Re-run, export XML, or delete.
- **Reverse parser.** Paste any `nmap …` command and we'll pre-fill the builder.
- **Privilege-aware.** Detects whether you launched as root or with `cap_net_raw`; greys out raw-socket scans (with a friendly explanation) when you didn't.
- **Cross-platform.** Linux (any modern distro), macOS, WSL2. Detects your installed Nmap version and gates flags accordingly.
- **Glossary + ethical-use splash** for first-time users; **keyboard shortcuts** + **light/dark/system theme** + **draft auto-save** for everyone else.
- **Desktop notifications** when a scan finishes (opt-in).
- **Auto-update notify.** Best-effort once-an-hour check against GitHub Releases — purely informational, never auto-installs. `--no-update-check` disables it.

---

## Install

You only need **`nmap`** itself on your machine; everything else (UI, server, parser, store) is in the binary. Nmap 7.40 or newer is recommended.

### Debian / Ubuntu / WSL-Ubuntu

```sh
# Replace VERSION with the latest tag from the Releases page.
curl -fLO https://github.com/nick-the-descended/n-mapped/releases/latest/download/n-mapped_VERSION_amd64.deb
sudo apt install ./n-mapped_VERSION_amd64.deb
```

The `.deb` declares `Depends: nmap`, so apt pulls nmap if it's missing. After install, you'll see "n-mapped" in the Activities/Network menu, plus a `n-mapped(1)` man page.

### macOS / Linux (any distro) — one-line installer

```sh
curl -fsSL https://raw.githubusercontent.com/nick-the-descended/n-mapped/main/scripts/install.sh | sh
```

Resolves the latest release, downloads the matching `linux-amd64` / `linux-arm64` / `darwin-amd64` / `darwin-arm64` tarball, **verifies the SHA256** against the published `SHA256SUMS`, and drops the binary in `~/.local/bin/n-mapped`.

### Direct download

Grab a release tarball from <https://github.com/nick-the-descended/n-mapped/releases/latest>, verify it against `SHA256SUMS`, extract, and put the `n-mapped` binary somewhere on your `$PATH`.

### From source

```sh
git clone https://github.com/nick-the-descended/n-mapped
cd n-mapped
make build         # uses the committed frontend bundle (no Node required)
sudo make install  # installs binary, .desktop, icon, and man page
```

If you change the frontend, `make frontend` rebuilds the Svelte SPA and refreshes the embedded copy. That step needs Node 22+. Plain users don't need Node — the committed bundle is always usable.

---

## Quickstart

```sh
n-mapped                        # opens http://127.0.0.1:8765 in your browser
n-mapped --no-browser           # don't auto-open
sudo n-mapped                   # unlock raw-socket scans (-sS, -sU, -O, -A, ...)
```

A safe first scan to try, no sudo needed:

1. Targets: `scanme.nmap.org` (the Nmap project's public test target — explicitly authorized for testing).
2. Click the **Service & version detection** profile card.
3. Hit **Run scan** (or `⌘⏎` / `Ctrl+Enter`).

The Results tab will populate live as nmap reports each host, finishing with a ports table per host.

---

## CLI options

| Flag                | Default          | Purpose                                                                          |
| ------------------- | ---------------- | -------------------------------------------------------------------------------- |
| `--bind`            | `127.0.0.1:8765` | host:port the local HTTP server listens on. Refuses non-loopback by design.     |
| `--nmap-path`       | (PATH lookup)    | Override the location of the `nmap` binary.                                      |
| `--privileged`      | `false`          | Advisory tag; the app does **not** elevate itself — run via `sudo` if you need raw sockets. |
| `--no-browser`      | `false`          | Skip auto-opening a browser tab on startup.                                      |
| `--no-update-check` | `false`          | Disable the once-an-hour "newer release available" GitHub poll.                  |
| `--version`         | —                | Print version and exit.                                                          |

---

## Privilege model

Most TCP scan types (`-sS`, `-sU`, `-O`, `-A`, plus a few NSE scripts) need raw sockets. `n-mapped` decides what's available **once at launch** — there are no per-scan password prompts.

| Launch                                              | UI behavior                                          |
| --------------------------------------------------- | ---------------------------------------------------- |
| `n-mapped`                                          | Privileged flags greyed out with explainer tooltip; TCP Connect (`-sT`) suggested as a no-root alternative. |
| `n-mapped` (binary has Linux `cap_net_raw` cap)     | All flags available; banner reads "capability mode". |
| `sudo n-mapped` (or `pkexec n-mapped --privileged`) | All flags available; banner reads "running as root". |

The backend **always re-validates** privilege server-side, so a malicious frontend payload can never sneak elevated-only flags past us.

---

## Where things live

| Path                                                   | What                                                                                         |
| ------------------------------------------------------ | -------------------------------------------------------------------------------------------- |
| `$XDG_DATA_HOME/n-mapped/history/<id>.json` (Linux)    | One JSON file per past scan. Parsed run + original XML.                                      |
| `~/Library/Application Support/n-mapped/history/...`   | Same on macOS.                                                                               |
| `<datadir>/favorites/<id>.json`                        | Saved templates from "Save current as favorite".                                             |
| `localStorage`                                         | Theme, ethical-use ack, in-progress draft, notification consent — browser-side only.         |

Override the data dir with `N_MAPPED_DATA_DIR=/path/to/dir n-mapped`.

---

## Keyboard shortcuts

| Keys                              | Action                              |
| --------------------------------- | ----------------------------------- |
| `?`                               | Open the keyboard shortcut help     |
| `⌘ K` / `Ctrl K`                  | Focus the flag search box           |
| `⌘ ⏎` / `Ctrl ⏎`                  | Run the current scan                |
| `⌘ .` / `Ctrl .`                  | Stop the running scan               |
| `⌘ S` / `Ctrl S`                  | Save current command as a favorite  |
| `1` / `2` / `3`                   | Switch Builder / Results / History  |
| `Esc`                             | Close any open dialog               |

---

## Architecture

- **Backend (Go, stdlib-only).** `net/http` server with the Svelte bundle embedded via `//go:embed`. `os/exec` runs nmap with `-oX -` so XML streams to a goroutine that decodes per-host elements with `encoding/xml.Decoder`. Streaming events are pushed to the browser as **Server-Sent Events** (`/api/scans/{id}/events`). History and favorites are JSON files; SQLite migration is queued for the diff/asset-tracking phase.
- **Frontend (Svelte 5 + Vite).** Catalog-driven UI. Reactive `$state` / `$derived` only — no Redux, no global Svelte stores beyond a tiny window event bus for keyboard-driven actions.
- **Catalog as data.** Flags, scripts, and built-in profiles live in `internal/catalog/data/*.json` and are embedded at compile time. Adding a new flag is a data-only change; the UI re-renders from the JSON without code edits.

### Project layout

```
n-mapped/
├── main.go, browser.go               # CLI entrypoint, cross-platform browser opener
├── Makefile                          # build / install / uninstall / release / dev
├── .goreleaser.yaml                  # cross-platform release pipeline (Linux+macOS, .deb)
├── packaging/                        # .desktop, icon.svg, n-mapped.1 man page
├── scripts/install.sh                # one-line installer with checksum verification
├── .github/workflows/                # CI (vet/test/build matrix), release (goreleaser)
├── internal/
│   ├── auth/                         # EUID / cap_net_raw detection
│   ├── catalog/                      # embedded flag + script + profile catalog
│   ├── config/                       # XDG / macOS data-dir resolution
│   ├── nmap/                         # version detection, command builder (validates everything),
│   │                                 # subprocess runner, streaming XML parser
│   ├── server/                       # net/http + REST + SSE + embedded SPA
│   ├── store/                        # JSON-file history + favorites
│   └── update/                       # once-an-hour GitHub Releases check
└── frontend/
    ├── src/components/               # Builder, ResultsView, HistoryPanel, ...
    ├── src/lib/                      # api client, types, draft, theme, notify, events
    └── dist/                         # generated; copied into internal/server/frontend_dist
```

---

## Security notes

- The HTTP server **only** binds `127.0.0.1` by default. Binding a non-loopback address requires an explicit override.
- The frontend never sends raw `nmap` arguments — it sends **flag IDs** that the backend re-validates against the catalog. This keeps the CLI surface tightly fenced.
- Targets are validated against a strict character allowlist before reaching nmap, and we always exec via `argv` (never `sh -c`).
- **Authorization gate:** non-RFC1918 / non-loopback targets prompt an explicit "I have permission to scan this" check; the first-run splash makes the same point legally.
- The auto-update check is metadata-only (read of GitHub Releases JSON). We never download or execute anything on your behalf.

---

## Development

See [`CONTRIBUTING.md`](./CONTRIBUTING.md) for the dev setup, code layout, and how to add a new flag or script (it's just JSON).

Quick path:

```sh
make frontend-dev      # Terminal 1 — Vite HMR on :5173, proxies /api -> :8765
make dev               # Terminal 2 — Go server on :8765 (use --no-browser if Vite is what you visit)
```

---

## License

[**PolyForm Noncommercial License 1.0.0**](./LICENSE) (Copyright © 2026 Nick-the-Descended).

In short:

- ✅ **Free for any noncommercial use** — personal projects, learning, hobby networks, hackathons, classroom teaching, public research, charities, government, etc.
- ✅ **Modify and redistribute** under the same license, as long as you pass the license text along.
- ❌ **No commercial use without a separate agreement** — you can't sell this, host it as a paid service, bundle it into a commercial product, or use it to generate revenue. If you want to do any of that, contact the maintainer for a commercial license.
- The copyright holder reserves the right to use this code commercially, license it on different terms to specific parties, and re-release it under any future license.

This is a **source-available** license, not OSI-approved "open source"; the OSI definition rules out commercial restrictions. GitHub will badge it as PolyForm-NC-1.0.0.

## Acknowledgements

This is a UI on top of the Nmap project's work. All the heavy lifting belongs to <https://nmap.org>; we just make it friendlier to click.
