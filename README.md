# n-mapped

A beginner-friendly local GUI for [Nmap](https://nmap.org). Click and search your way through every flag, save your favorite scans, watch live progress, and learn nmap from inline plain-English explanations.

> **Status:** Phase 0 — backend skeleton is in. The frontend (Svelte) lands next. The Go binary already serves a placeholder page and a working API (`/api/catalog`, `/api/version`, `/api/privilege`, `/api/scans`).

## Goals

- **One-command install.** No `pip install` *and* `npm install` chain. Eventually: `apt install n-mapped`, or download one binary, or `git clone && make install`.
- **No runtime dependencies on the user's machine** other than `nmap` itself. The frontend is embedded into the Go binary at build time.
- **Beginner-first.** Plain-English flag descriptions, skill-level filter, goal-driven wizard, ethical-use splash on first launch.
- **Cross-version & cross-platform.** Detects the installed nmap at startup and gates flags accordingly. Works on Linux, macOS, and WSL.

## Quickstart (during development)

```sh
git clone https://github.com/nick-the-descended/n-mapped
cd n-mapped
make build
./n-mapped              # opens http://127.0.0.1:8765
```

To enable raw-socket scans (`-sS`, `-sU`, `-O`, ...):

```sh
sudo ./n-mapped         # or: ./n-mapped --privileged (advisory; sudo is the actual elevator)
```

## CLI flags

| Flag           | Default          | Purpose                                                          |
| -------------- | ---------------- | ---------------------------------------------------------------- |
| `--bind`       | `127.0.0.1:8765` | Where the local HTTP server listens                              |
| `--nmap-path`  | (PATH lookup)    | Override the nmap binary location                                |
| `--privileged` | `false`          | Advisory tag; actual elevation comes from running via `sudo`     |
| `--no-browser` | `false`          | Skip auto-opening a browser tab on startup                       |
| `--version`    | —                | Print version and exit                                           |

## Project layout

```
n-mapped/
├── main.go                            # CLI entrypoint
├── browser.go                         # cross-platform "open my browser" helper
├── Makefile
├── go.mod
├── internal/
│   ├── auth/        # privilege detection (EUID / Linux capabilities)
│   ├── catalog/     # flag catalog (embedded JSON, queryable by id/search)
│   │   └── data/flags.json
│   ├── config/      # XDG / macOS data dir resolution
│   ├── nmap/        # version detection, command builder, subprocess runner
│   └── server/      # net/http server + embedded SPA + REST + SSE streaming
└── frontend/        # Vite + Svelte SPA (added in Phase 1)
```

See the full plan in `/root/.claude/plans/final-goal-is-for-zippy-crescent.md`.

## License

TBD.
