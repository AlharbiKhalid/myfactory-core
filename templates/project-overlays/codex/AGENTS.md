# AGENTS.md — MyFactory-Enabled Repository

You are operating inside a MyFactory-enabled repository. MyFactory is an AI
software factory: Git is the source of truth, Plane is the execution tracker,
and all agents work from structured task packages under strict quality gates.

## Read before doing anything

1. `.ApplicationFactory/product.yaml` — product identity and the source-of-truth file map.
2. `.ApplicationFactory/config.yaml` — project configuration (git provider, autonomy mode, Plane, orchestration).
3. `.ApplicationFactory/task-packages/` — your task package, if you were assigned one.
4. The docs referenced by your task package (product, business, architecture, QA).

## Source of truth

- Git files are the source of truth. Chat, issue comments, and your own memory are not.
- If chat instructions conflict with the source-of-truth files, follow the files and flag the conflict.
- Business rules live in `docs/01-business/business-rules.yaml`. Every rule has a stable `BR-*` ID.

## Rules you must follow

- Do not implement anything without a task package in `.ApplicationFactory/task-packages/`. If none exists for your work, stop and say so.
- Do not change business rules unless your task package is a BUSINESS_RULE task.
- Do not change architecture docs unless your task package is an ARCHITECTURE task.
- Do not approve your own PR/MR. Do not merge your own PR/MR. Do not merge to main at all.
- Do not deploy anything.
- Do not edit files outside the Allowed Files list in your task package.
- Do not mark work done without evidence: tests run, CI links, QA reports.

## How to work

### Discovery tasks (product / business / architecture docs)

- Update documentation files only. Do not touch source code.
- Record assumptions and open questions inside the docs you edit.
- Business rules you add must get stable `BR-*` IDs and entries in `docs/04-qa/business-rule-test-matrix.csv`.

### Implementation tasks

1. Read your task package fully. Read every file it lists under Source of Truth.
2. Create a branch named after the task ID (for example `task/APP-BE-001`).
3. Implement only what the task package's acceptance criteria require.
4. Add or update the tests the task package requires. Run the required commands.
5. Open a PR/MR using the repository's PR template. Fill every section.
6. Reference the task package path and any `BR-*` IDs touched in the PR description.
7. Attach evidence: test output, commands run, CI links.
8. Respond to review and QA findings on the PR. Do not self-approve.

## Quality gates you are subject to

- CI must pass before QA review.
- Functional QA (QA Agent) validates behavior against acceptance criteria.
- Business QA (Business QA Agent) validates code against `BR-*` business rules whenever business logic is touched. It is a separate gate from functional QA.
- Failing either gate produces a fix task; expect one rather than arguing with the report.
