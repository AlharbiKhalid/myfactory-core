# Sprint Run Loop

The deterministic control loop Hermes follows when running a sprint.
Each numbered step must complete (or be explicitly skipped with a recorded reason)
before the next begins.

## Preconditions

- `.ApplicationFactory/product.yaml` and `.ApplicationFactory/config.yaml` exist.
- A sprint exists with `status: ready` or `status: active`.
- Sprint entry criteria are met.
- Work breakdown items referenced by the sprint scope exist.

## Loop

```text
1. LOAD STATE
   Read config, missions, sprints, work breakdown, runtime state,
   server registry, and Plane state (if enabled + key present).

2. SELECT SPRINT
   active sprint = sprint with status: active
   else first sprint with status: ready -> set status: active

3. BUILD READY QUEUE
   ready task = in sprint scope
     AND dependencies done
     AND has required fields
     AND not already assigned or in progress

4. PACKAGE
   For each ready task without a task package:
     create .ApplicationFactory/task-packages/<TASK-ID>.md
     from TASK-PACKAGE-TEMPLATE.md, filled from work-breakdown.yaml.

5. ASSIGN
   While assigned_count < max_parallel_tasks and queue not empty:
     pick highest-priority ready task
     pick an available server/agent from the registry
     record assignment in RUNTIME-STATE.yaml
     instruct agent: work only from the task package; branch + PR required.

6. MONITOR
   For each in-progress task:
     check branch exists, PR/MR opened, CI status, review comments.
     record evidence (branch, PR URL, CI run URL).

7. QA GATE
   When CI passes on a PR:
     trigger QA Agent (functional QA) -> QA report required.
   If task touches business rules (BR-* IDs):
     trigger Business QA Agent -> Business QA report required.
   QA and Business QA are separate; both must pass when both apply.

8. RESULT
   PASS: mark task Ready to Merge. Merge requires human approval. STOP for that task.
   FAIL: create fix task with QA report as input, add to sprint scope, go to 3.

9. SYNC
   Update RUNTIME-STATE.yaml.
   Sync statuses to Plane if enabled (dry-run unless explicitly applied).

10. EXIT CHECK
    If sprint exit criteria met -> report and set sprint status: done (pending human confirmation).
    If all tasks blocked or approval needed -> STOP and report.
    Else -> go to 3.
```

## Invariants

- Nothing is deleted. Failed work produces fix tasks, not history rewrites.
- Every state change writes evidence into RUNTIME-STATE.yaml.
- The loop never merges, never deploys, never approves its own gates.
- A task that fails the same gate twice escalates to a human instead of retrying.
