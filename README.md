# MyFactory

MyFactory is a reusable AI software factory. It turns any repository into a
structured environment where AI agents (Claude, Codex, and others) can
discover product requirements, extract business rules, define architecture,
break work down, and implement it — under strict quality gates, with Git as
the source of truth and Plane as the execution tracker.

MyFactory is not one app. It is the factory that builds apps.

## Install

Goal (future, once hosted):

```bash
curl -sSL https://domain.com/cli/install | bash
```

Today, from a local checkout:

```bash
bash scripts/install-local.sh   # creates a `myfactory` shim in ~/.local/bin
myfactory --help
```

Or without installing anything:

```bash
python -m myfactory --help
```

The remote installer (`scripts/install.sh`) reads `MYFACTORY_REPO_URL` and
installs into `~/.myfactory/core`. It never requires root and never stores
secrets.

## Core commands

```text
myfactory init                       Set up factory structure in a repo (non-interactive)
myfactory doctor                     Readiness report for the current repo
myfactory discover --print-prompt    Print the AI discovery prompt (paste into Claude/Codex)
myfactory plan --dry-run             Report whether source docs are ready for planning
myfactory plan --print-prompt        Print the AI planning prompt
myfactory plane sync --dry-run       Show what would sync to Plane (never calls APIs by default)
myfactory orchestrator prompt        Print the Hermes controller prompt
```

All commands accept `--target PATH` to operate on another directory.

## Setup vs discovery — the key distinction

- **`myfactory init` is setup only.** It copies structure: docs skeletons,
  `.ApplicationFactory/` metadata, agent instructions, GitHub/GitLab helpers,
  Claude commands, Codex `AGENTS.md`. It asks no questions, calls no AI, and
  never overwrites existing files (unless `--force`).
- **Discovery is agent-driven.** After init, you paste the discovery prompt
  (or run `/myfactory-discover` in Claude) and an AI agent interviews you and
  fills the source-of-truth docs. The same applies to business rules,
  architecture, work breakdown, and sprint planning.

## Lifecycle

1. Install the CLI globally.
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

## Repository layout

```text
myfactory/            CLI package (python -m myfactory)
agents/               Reusable agent role instructions
config/               Lifecycle, agent registry, global config example
scripts/              create-product.py, install.sh, install-local.sh
templates/
  product-repo/       Everything `init` copies into a project
  project-overlays/   Codex AGENTS.md + Claude .claude/commands
```

`scripts/create-product.py` still works for creating a brand-new product repo
from the template; `myfactory init` is the path for existing repos.

## Current limitations

- Plane sync is dry-run only; the live API client is not implemented yet.
- Hermes is a prompt/protocol, not a daemon — you run it inside an AI session.
- GitHub/GitLab communication is convention-based (templates, branch/PR
  rules); no direct API automation yet.
- Development server delegation is a registry format + protocol, not yet
  automated transport.
- No packaged release yet; install is via shim or `pip install -e .`.

## Next steps

- Live Plane API client behind `plane sync --apply`.
- `myfactory new` wrapping `scripts/create-product.py`.
- GitHub/GitLab API adapters for PR/MR status reading.
- Hermes runtime harness for development servers.
- Hosted install endpoint for `curl | bash`.
