# MyFactory Agent Protocol

This protocol defines the mandatory behavior for all MyFactory agents.

Every agent must follow this protocol unless a higher-priority human instruction explicitly overrides it.

---

# 1. Core Mission

MyFactory agents convert software ideas into traceable, tested, reviewed, and maintainable software.

Agents must optimize for:

- Correctness
- Traceability
- Maintainability
- Business logic accuracy
- Testability
- Clear evidence
- Safe handoffs between agents

Agents must not optimize only for producing code quickly.

---

# 2. Source of Truth

The source of truth is the product repository.

Agents must use files in the product repository as the authoritative context.

Primary source-of-truth locations:

```text
docs/00-product/
docs/01-business/
docs/02-architecture/
docs/03-delivery/
docs/04-qa/
docs/05-operations/
.ApplicationFactory/
Agents must not treat chat history as the long-term source of truth.

Chat may be used for discussion, but final decisions must be written into source-of-truth files.

3. Required Context Before Work

Before starting any task, an agent must read the task package.

Task packages live in:

.ApplicationFactory/task-packages/

The task package defines:

Task ID
Task type
Goal
Non-goals
Source documents
Business rule IDs
Acceptance criteria
Definition of Done
Allowed files
Forbidden files
Required tests
Required commands
Required evidence

If there is no task package, the agent must not perform implementation work.

4. Traceability Rules

Every important change must be traceable.

The trace should connect:

Product requirement
  → Business rule
  → Architecture decision
  → Work item
  → Task package
  → Git branch
  → Pull request
  → Tests
  → QA report
  → Business QA report

Agents must preserve IDs such as:

FR-001
AC-001
BR-001
API-001
TASK-001
BRT-001

Agents must not remove or rename stable IDs without explicit instruction.

5. Business Logic Rules

Business logic is a first-class asset.

Agents must not invent hidden business logic.

If implementation requires a business rule that does not exist, the agent must stop and report:

Missing business rule detected.

The agent should then request or create a BUSINESS_RULE task instead of silently implementing the rule.

Business logic must be documented in:
docs/01-business/business-rules.yaml
docs/01-business/decision-tables.md
docs/04-qa/business-rule-test-matrix.csv

Important business logic must include:

Stable rule ID
Plain-language description
Positive examples
Negative examples
Boundary cases when relevant
Exception cases when relevant
Test requirements
6. Architecture Rules

Agents must respect the architecture documents.

Important architecture files:

docs/02-architecture/system-overview.md
docs/02-architecture/domain-model.md
docs/02-architecture/data-model.md
docs/02-architecture/api-contracts.md
docs/02-architecture/adr/

Agents must not change architectural direction inside implementation work unless the task explicitly allows it.

If a task requires architecture change, create or request an ARCHITECTURE task.

Core business logic should live in the domain/business layer, not randomly inside:

UI components
API controllers
Database triggers
Background jobs
Utility files

7. File Boundary Rules

Agents must respect allowed and forbidden files in the task package.

If a required change touches forbidden files, the agent must stop and report:

Task boundary conflict detected.

Agents must avoid unrelated changes.

Unrelated formatting, renaming, refactoring, or dependency updates are not allowed unless the task explicitly asks for them.

8. Testing Rules

Agents must add or update tests when behavior changes.

Testing should match the task type.

Examples:

Business rules → unit tests and business-rule tests
API behavior → API tests
Database behavior → integration tests
User workflows → end-to-end or workflow tests
Bugs → regression tests

Agents must not claim tests passed unless they actually ran the required commands or CI produced evidence.

9. Evidence Rules

Agents must provide evidence before marking work complete.

Valid evidence includes:

GitHub PR URL
Git branch name
CI run URL
Test output
QA report path
Business QA report path
Review comments
Screenshots when relevant
Logs when relevant

Summaries are not evidence by themselves.

10. Pull Request Rules

Implementation agents must work through pull requests.

A PR must include:

Task ID
Plane issue ID if available
Task package path
Source documents
Business rule IDs touched
Summary of changes
Tests added or updated
Commands run
Known risks
Evidence links

Agents must not merge their own PRs.

Agents must not approve their own work.

11. QA Rules

Functional QA and Business QA are separate.

Functional QA asks:

Does the software work according to requirements and acceptance criteria?

Business QA asks:

Does the software make the correct business decision according to business rules?

Business QA is required when business rules are created, changed, or implemented.

12. Status Update Rules

Agents must update task status only when required evidence exists.

Examples:

In Progress → PR Open
requires:
- GitHub PR URL
- Branch name
- Summary of changes

PR Open → QA Review
requires:
- CI passed
- PR template completed

- Tests added or updated when required

QA Review → Business QA Review
requires:
- QA report exists
- Business rule IDs identified

Business QA Review → Ready to Merge
requires:
- Business QA report passed
- No blocking findings

Agents must not mark work Done just because the implementation appears finished.

13. Blocking Conditions

Agents must stop and report a blocker when:

Source-of-truth files are missing.
Task package is missing.
Acceptance criteria are unclear.
Required business rules are missing.
Architecture conflicts with the task.
Required files are forbidden by the task package.
Tests cannot be run.

CI is failing for unclear reasons.
Security-sensitive behavior is ambiguous.
The requested task would require production deployment without approval.

A blocker report should include:

Blocker:
Why it blocks progress:
Files inspected:
Recommended next task:
14. Output Format

When completing a task, agents should report:

Task ID:
Result:
Files changed:
Business rules touched:
Tests added or updated:
Commands run:
Evidence:
Risks:
Next recommended step:

When blocked, agents should report:

Task ID:
Status: Blocked
Blocker:
Reason:
Evidence:
Recommended next action:
15. Forbidden Actions

No agent may:

Approve its own work.
Merge to main.
Deploy to production without explicit approval.
Invent undocumented business rules.
Ignore task package boundaries.
Change unrelated files.
Mark work done without evidence.
Hide test failures.
Delete source-of-truth documents.
Remove traceability IDs without explicit instruction.
Change security-sensitive behavior without review.

16. MyFactory Principle

A fast but untraceable result is not acceptable.

A correct MyFactory result must be:

Documented
Traceable
Tested
Reviewed
Evidence-backed
Reusable

