# Contributing to n-mapped

Thanks for considering a contribution. This guide is short on purpose — read once, then look for specific patterns in the code.

## Prerequisites

- **Go 1.24+** (we use `embed.FS` and modern `encoding/xml`).
- **Node 22+** + npm. Only required if you change frontend source. Pure backend work needs neither.
- **nmap 7.40+** if you want to actually run scans during development.
- macOS or Linux for development (Windows works via WSL2).

The dev setup deliberately avoids CGO so the binary stays a single static file. Don't introduce CGO without a good reason.

## One-time setup

```sh
git clone https://github.com/nick-the-descended/n-mapped
cd n-mapped
go mod download
( cd frontend && npm ci )
```

## The dev loop

Two terminals:

```sh
# Terminal 1 — Vite dev server with HMR on :5173, proxies /api/* to :8765.
make frontend-dev

# Terminal 2 — Go backend on :8765. Pick your nmap binary (system or fake).
make dev ARGS="--no-browser"
```

Visit <http://localhost:5173> (Vite) for the live-reload SPA backed by your local Go server.

### Quick fake nmap for offline testing

The backend runner doesn't care whether the binary it spawns is the real nmap. A trivial shell script that prints canned XML lets you exercise the entire pipeline without root or even network access. There's an example in the project's commit history; the gist is:

```sh
cat > /tmp/fake-nmap <<'EOF'
#!/usr/bin/env bash
case "$1" in
  --version) echo "Nmap version 7.94"; exit 0 ;;
esac
cat <<XML
<?xml version="1.0"?>
<nmaprun scanner="nmap" version="7.94" start="1700000000">
  <host><status state="up"/><address addr="192.0.2.5" addrtype="ipv4"/>
    <ports><port protocol="tcp" portid="22"><state state="open"/><service name="ssh" product="OpenSSH" version="9.6"/></port></ports>
  </host>
  <runstats><finished elapsed="2.0" exit="success"/><hosts up="1" down="0" total="1"/></runstats>
</nmaprun>
XML
EOF
chmod +x /tmp/fake-nmap
make dev ARGS="--nmap-path /tmp/fake-nmap --no-browser"
```

## Build, test, lint

```sh
make test       # go test ./... -race
make vet        # go vet ./...
make build      # produces ./n-mapped
( cd frontend && npx svelte-check )   # frontend typecheck
( cd frontend && npm run build )      # produce dist/
```

When you change Svelte source, refresh the embedded copy:

```sh
make frontend          # builds frontend, copies dist/ -> internal/server/frontend_dist/
```

CI fails if `internal/server/frontend_dist/` drifts from `frontend/src/`, so commit both.

## How to add a new flag

The flag catalog lives in [`internal/catalog/data/flags.json`](./internal/catalog/data/flags.json). Each entry is a JSON object that the UI renders directly — there's no code to write.

```json
{
  "id": "no-dns",
  "short": "-n",
  "category": "host-discovery",
  "value_type": "boolean",
  "skill_level": "intermediate",
  "short_description": "Never do DNS resolution",
  "long_description": "Skips reverse-DNS lookups on every host. Speeds up scans when DNS is slow or unreachable.",
  "tags": ["dns", "fast"]
}
```

Required: `id`, `category`, `skill_level`, `short_description`, and one of `short` or `long`. Everything else is optional but encouraged — examples and warnings are what makes the UI useful for beginners.

After editing, `make build && ./n-mapped` and the new flag is live.

## How to add a new NSE script

Same idea, in [`internal/catalog/data/scripts.json`](./internal/catalog/data/scripts.json). Mark intrusive ones honestly:

```json
{
  "id": "smb-vuln-ms17-010",
  "categories": ["intrusive", "vuln", "exploit"],
  "skill_level": "advanced",
  "short_description": "Test for MS17-010 (EternalBlue)",
  "warnings": [
    { "level": "danger", "text": "Active exploit check — only run with explicit authorization." }
  ]
}
```

The UI will prompt the user for an explicit confirmation before adding any script in `intrusive` / `exploit` / `dos` / `brute` categories.

## How to add a built-in profile

Edit [`internal/catalog/data/profiles.json`](./internal/catalog/data/profiles.json):

```json
{
  "id": "fast-vuln",
  "name": "Fast vuln scan",
  "description": "Top 100 ports + service detect + vulners CVE lookup.",
  "icon": "🩹",
  "skill_level": "intermediate",
  "flag_ids": ["fast-scan", "version-detect", "timing-template"],
  "flag_values": { "timing-template": "4" },
  "script_ids": ["vulners"]
}
```

## Code style

- **Go:** plain stdlib unless absolutely necessary. `gofmt` and `go vet` clean. Prefer narrow interfaces and explicit error wrapping (`%w`). Avoid panics in request paths.
- **Svelte:** Svelte 5 runes (`$state`, `$derived`, `$effect`). No global stores; use the small `lib/events.ts` bus for window-level signals. Keep components self-contained and props-driven.
- **Comments:** prefer naming over commentary. Comment the *why* — invariants, surprising constraints — not the *what*.
- **Don't add backward-compat shims** when you can just change the code; we have semver releases for breaking changes.

## Releasing

Tag and push:

```sh
git tag vX.Y.Z
git push origin vX.Y.Z
```

The `release.yml` workflow runs goreleaser, cross-compiles Linux/macOS (amd64+arm64), packages a `.deb`, and publishes a GitHub Release with `SHA256SUMS`. Local dry-run:

```sh
goreleaser release --snapshot --clean
```

## Reporting issues

- Bugs: include the output of `n-mapped --version` and the relevant log lines (logs go to stderr).
- Feature requests: please describe the user flow you'd want, not just "add support for X".
- Security issues: open a private security advisory on GitHub rather than a public issue.

## License & contributor terms

This project is licensed under [PolyForm Noncommercial 1.0.0](./LICENSE) — free for noncommercial use, but commercial use requires a separate agreement with the copyright holder.

By submitting a pull request, you agree that:

1. You wrote the code yourself (or have the right to contribute it).
2. Your contribution is licensed under PolyForm Noncommercial 1.0.0 like the rest of the project.
3. **You grant the project's copyright holder a perpetual, irrevocable license to relicense your contribution under different terms** (including commercial terms) without further consent. This is what lets the maintainer offer commercial licenses to companies who want them, dual-license to a more permissive license later, or accept a corporate sponsorship that requires it. If this is a problem for you, please open an issue first to discuss.

If your employer has any claim on the code you write, please make sure they're OK with you contributing under these terms before opening a PR.
