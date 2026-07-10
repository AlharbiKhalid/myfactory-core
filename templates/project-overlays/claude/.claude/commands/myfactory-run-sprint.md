# MyFactory: Run Sprint (Prepare Orchestration)

You prepare a sprint for execution. You do NOT implement tasks in this
command. Implementation is delegated to the Hermes controller and its
implementation agents, each working from a task package.

## Steps

1. Read `.ApplicationFactory/product.yaml` and `.ApplicationFactory/config.yaml`.
2. Read `docs/03-delivery/sprints.yaml` and select the sprint:
   - the one the user named, else the `active` sprint, else the first `ready` one.
   - If none is ready, stop and tell the user to run `/myfactory-plan-sprints`.
3. Verify entry criteria. Report any that fail; do not proceed past failures
   without explicit user confirmation.
4. For every work item in the sprint scope, ensure a task package exists at
   `.ApplicationFactory/task-packages/<TASK-ID>.md`. Create missing ones from
   `TASK-PACKAGE-TEMPLATE.md`, filled from `work-breakdown.yaml`:
   goal, source docs, `BR-*` rules, acceptance criteria, definition of done,
   allowed/forbidden files, required tests and commands.
5. Initialize or update `.ApplicationFactory/orchestrator/RUNTIME-STATE.yaml`
   with the active mission/sprint and an empty assignment list.
6. Hand off to the controller:
   - Print (or tell the user to run) `myfactory orchestrator prompt` — that is
     the prompt to give Hermes.
   - If Hermes is not enabled in config, the user can paste the controller
     prompt into a Claude/Codex session manually.

## Hard limits

- Do not write application code.
- Do not assign yourself implementation tasks.
- Do not merge, approve, or deploy anything.
- Do not mark the sprint active without its entry criteria passing.

## Finish

Report: sprint selected, entry criteria status, task packages created/existing,
runtime state written, and the exact next step for the user (give the
controller prompt to Hermes, or run tasks manually one package at a time).
