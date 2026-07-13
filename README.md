# MyFactory

MyFactory is a reusable AI software factory. It turns any repository into a
structured environment where AI agents (Claude, Codex, and others) can
discover product requirements, extract business rules, define architecture,
break work down, and implement it — under strict quality gates, with Git as
the source of truth and Plane as the execution tracker.

MyFactory is not one app. It is the factory that builds apps.

The CLI is a **standalone native executable** written in Go. All project
templates are embedded in the binary — users need **no Python, no Go, no
Node.js, and no source checkout**.

## Install

### Linux / macOS / Git Bash (Bash)

```bash
curl -fsSL "https://raw.githubusercontent.com/AlharbiKhalid/myfactory-core/main/scripts/install.sh" | bash
```

### Windows (PowerShell)

```powershell
irm "https://raw.githubusercontent.com/AlharbiKhalid/myfactory-core/main/scripts/install.ps1" | iex
```

Both installers download a prebuilt binary from GitHub Releases, **verify its
SHA-256 checksum** against the release's `checksums.txt`, install into a
user-writable directory (`~/.local/bin` or `%LOCALAPPDATA%\Programs\myfactory`),
require no root/admin, and finish by running `myfactory version`. HTTP
failures fail closed (`curl -f`); TLS verification is never disabled; only
versioned release assets are downloaded, never branch source. Signed release
artifacts (e.g. Sigstore/minisign) can be added later.

Options via environment variables: `MYFACTORY_VERSION` (default: latest),
`MYFACTORY_REPOSITORY` (default: `AlharbiKhalid/myfactory-core`),
`MYFACTORY_INSTALL_DIR`.

### Manual download

Grab the archive for your platform from the
[Releases page](https://github.com/AlharbiKhalid/myfactory-core/releases),
verify it against `checksums.txt` (`sha256sum -c`), extract, and put
`myfactory`/`myfactory.exe` on your PATH.

| OS | amd64 | arm64 |
|---|---|---|
| Linux | `myfactory_<v>_linux_amd64.tar.gz` | `myfactory_<v>_linux_arm64.tar.gz` |
| macOS | `myfactory_<v>_darwin_amd64.tar.gz` | `myfactory_<v>_darwin_arm64.tar.gz` |
| Windows | `myfactory_<v>_windows_amd64.zip` | `myfactory_<v>_windows_arm64.zip` |

## Core commands

```text
myfactory init                       Set up factory structure in a repo (non-interactive)
myfactory doctor                     Readiness report for the current repo
myfactory discover --print-prompt    Print the AI discovery prompt (paste into Claude/Codex)
myfactory plan --dry-run             Report whether source docs are ready for planning
myfactory plan --print-prompt        Print the AI planning prompt
myfactory plane sync --dry-run       Show what would sync to Plane (never calls APIs by default)
myfactory orchestrator prompt        Print the Hermes controller prompt
myfactory version                    Version, git commit, and build date
```

All commands accept `--target PATH` to operate on another directory. The
binary works from any working directory; templates come from inside the
executable (`MYFACTORY_ASSETS_DIR` can override them for template development).

## Setup vs discovery — the key distinction

- **`myfactory init` is setup only.** It copies structure: docs skeletons,
  `.ApplicationFactory/` metadata, agent instructions, GitHub/GitLab helpers,
  Claude commands, Codex `AGENTS.md`. It asks no questions, calls no AI, and
  never overwrites existing files (unless `--force`). Nothing is ever deleted.
- **Discovery is agent-driven.** After init, you paste the discovery prompt
  (or run `/myfactory-discover` in Claude) and an AI agent interviews you and
  fills the source-of-truth docs. The same applies to business rules,
  architecture, work breakdown, and sprint planning.

## Lifecycle

1. Install the CLI (one binary).
2. `myfactory init` inside any repo → factory structure appears.
3. Claude/Codex discovery agents populate product docs and business rules
   (`/myfactory-discover`, `/myfactory-business-rules`, `/myfactory-architecture`).
4. Planning agents generate work breakdown, missions, and sprints
   (`/myfactory-work-breakdown`, `/myfactory-plan-sprints`).
5. `myfactory plane sync` mirrors the plan into Plane (execution tracker).
6. Hermes — the orchestrator/controller — runs the sprint loop
   (`myfactory orchestrator prompt`), delegating tasks to Claude/Codex agents
   on development servers.
7. Agents communicate exclusively through Git: branches, PRs/MRs, comments,
   CI results, QA reports.
8. Functional QA validates behavior; Business QA validates business rules.
   Failures create fix tasks and the loop continues.

## How source of truth works

Git files are authoritative. Chat never is.

| Concern | Source of truth |
|---|---|
| Product | `docs/00-product/` (idea brief, PRD, journeys, acceptance criteria) |
| Business rules | `docs/01-business/business-rules.yaml` — every rule has a stable `BR-*` ID |
| Architecture | `docs/02-architecture/` + ADRs |
| Delivery plan | `docs/03-delivery/` (work-breakdown, missions, sprints) |
| QA | `docs/04-qa/` (strategy, test matrices, report templates) |
| Project config | `.ApplicationFactory/config.yaml` |
| Task packages | `.ApplicationFactory/task-packages/` — agents work only from these |

## How Plane fits

Plane tracks execution state; it never holds truth that Git lacks.
Mapping: **MyFactory Mission → Plane Module/Label · Sprint → Plane Cycle ·
Task → Plane Issue.** `myfactory plane sync` is dry-run by default and only
attempts live sync with `--apply` + `plane.enabled: true` + the API key env
var set (live calls are not yet implemented; see Limitations).

## How Hermes fits

Hermes is the per-project orchestrator/controller. It reads the active sprint,
selects ready tasks, generates task packages, assigns work to Claude/Codex
agents on registered development servers, watches CI, triggers QA and Business
QA, creates fix tasks on failure, and stops when gates fail or human approval
is needed. Its contract lives in
`.ApplicationFactory/orchestrator/HERMES-CONTROLLER-PROMPT.md` and
`SPRINT-RUN-LOOP.md`. Hard rules: no agent approves its own work, no agent
merges to main, no production deploys, no work without a task package, no
progress without evidence.

## Missions, sprints, tasks

- **Mission** — a larger goal (`MISSION-001: Build booking MVP`), in
  `docs/03-delivery/missions.yaml`.
- **Sprint** — an executable scope inside a mission, in
  `docs/03-delivery/sprints.yaml`. One sprint = one Plane Cycle.
- **Task** — a work item in `docs/03-delivery/work-breakdown.yaml` with type,
  source docs, acceptance criteria, definition of done, and dependencies.
  One task = one Plane Issue = one task package = one agent session.

## QA and Business QA

Two separate gates, run by separate agents:

- **Functional QA** (`agents/qa-agent.md`) validates behavior against
  acceptance criteria and test matrices.
- **Business QA** (`agents/business-qa-agent.md`) validates implementation
  against `BR-*` business rules and decision tables, and hunts for hidden
  business logic.

A task touching business rules must pass both. A failure in either produces a
fix task — the factory keeps moving.

## Development

Local development requires **Go 1.23+** (users never need it):

```bash
go build -o dist/myfactory ./cmd/myfactory   # build
go test ./...                                # tests
go vet ./...                                 # static checks
bash scripts/install-local.sh                # build + install from checkout
powershell -File scripts/install-local.ps1   # same, native Windows
```

The Go module's only third-party dependency is `gopkg.in/yaml.v3`, which is
compiled into the binary — end users still install nothing beyond the
executable. MyFactory accepts any standards-compliant YAML (as written by
Claude, Codex, PyYAML, or ordinary editors), including anchors/aliases,
indentless block sequences, and block scalars. A delivery file that fails to
parse makes `plane sync` exit non-zero with the file path and reason instead
of reporting a zero-item plan. Templates under
`templates/` are embedded at build time via `go:embed all:...` (the `all:`
prefix is what preserves hidden paths like `.ApplicationFactory` and
`.claude`; `internal/assets/assets_test.go` guards this).

### Creating a release

```bash
git tag v0.3.0
git push origin v0.3.0
```

`.github/workflows/release.yml` then runs tests and vet, cross-compiles
`CGO_ENABLED=0` binaries for linux/darwin/windows × amd64/arm64, injects
version metadata via ldflags, smoke-tests a real `init`, generates
`checksums.txt` (SHA-256), and uploads everything to the GitHub Release.

## Repository layout

```text
cmd/myfactory/        Go CLI entrypoint
internal/             Go implementation (cli, commands, assets, fsops, ...)
assets.go             go:embed of templates/ (must sit at repo root)
agents/               Reusable agent role instructions
config/               Lifecycle, agent registry, global config example
scripts/              Installers + create-product.py
templates/
  product-repo/       Everything `init` copies into a project
  project-overlays/   Codex AGENTS.md + Claude .claude/commands
myfactory/            LEGACY: Python CLI, kept only as migration reference
```

## Python CLI migration status

The original Python implementation (`myfactory/`, `pyproject.toml`) is
**retained temporarily as the behavioral reference only** — see
`myfactory/LEGACY.md`. The Go CLI is the primary, supported CLI, and the
installers only ever install the Go binary. `scripts/create-product.py` still
works for creating a brand-new product repo from the template.

No product/business workflow behavior was intentionally changed in the port.
Intentional CLI differences from the Python version:

- Repeat `init` is now truly idempotent: placeholders are only filled in
  files created by that run (Python re-ran replacement on every invocation
  and could consume placeholders deeper in existing files, e.g.
  `plane.workspace.name`).
- `--git-provider gitlab` no longer leaves empty `.github/` directories.
- Files are written byte-for-byte from templates (LF); Python rewrote
  placeholder files with platform line endings (CRLF on Windows).
- New `myfactory version` command with build metadata.

## Current limitations

- Plane sync is dry-run only; the live API client is not implemented yet.
- Hermes is a prompt/protocol, not a daemon — you run it inside an AI session.
- GitHub/GitLab communication is convention-based (templates, branch/PR
  rules); no direct API automation yet.
- Development server delegation is a registry format + protocol, not yet
  automated transport.
- Release binaries are checksummed but not yet signed.

## Next steps

- First tagged release (`v0.3.0`) to light up the binary installers.
- Live Plane API client behind `plane sync --apply`.
- `myfactory new` wrapping the create-product flow.
- GitHub/GitLab API adapters for PR/MR status reading.
- Hermes runtime harness for development servers.
- Remove the legacy Python package after a deprecation window
  (list in `myfactory/LEGACY.md`).
