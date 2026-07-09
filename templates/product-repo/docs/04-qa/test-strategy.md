# Test Strategy

## Purpose

This document defines how the product will be tested.

QA must validate both:

1. Software correctness.
2. Business logic correctness.

## Testing Principles

- QA starts before implementation.
- Every important business rule should have test coverage.
- Business logic must be tested separately from general software behavior.
- Tests should be traceable to requirements, business rules, and acceptance criteria.
- Agents must provide evidence, not only summaries.

## Test Levels

### Static Checks

Examples:

- Linting
- Formatting
- Type checking
- Schema validation
- Dependency checks

### Unit Tests

Purpose:

Validate small units of logic, especially domain/business rules.

Required for:

- Business policies
- Calculations
- Permission logic
- State transitions
- Validation rules

### Integration Tests

Purpose:

Validate that multiple internal components work together.

Required for:

- Database interactions
- Service interactions
- Background jobs
- Internal workflows

### API Tests

Purpose:

Validate backend contracts directly.

Required for:

- Business-critical endpoints
- Permission-sensitive endpoints
- Data mutation endpoints

### End-to-End Tests

Purpose:

Validate full user workflows.

Use for:

- Critical user journeys
- High-risk flows
- Release smoke tests

### Manual Exploratory Testing

Purpose:

Find issues that scripted tests may miss.

Focus areas:

- Edge cases
- Usability issues
- Unexpected user behavior
- Error states
- Race conditions
- Permissions

## Business Logic QA

Business Logic QA validates whether the system makes the correct business decisions.

It must check:

- Business rule IDs
- Decision tables
- Boundary cases
- Exception cases
- Rejection reasons
- Audit requirements
- Hidden undocumented business logic

## Regression Testing

Regression testing ensures new changes do not break existing behavior.

Critical regression areas:

- Authentication
- Authorization
- Core workflows
- Business-critical rules
- Data integrity
- External integrations
- Reporting
- Notifications

## QA Gates

Before a task can move to QA Review:

- GitHub PR must exist.
- CI must pass.
- Required tests must be added or updated.
- PR must reference the task ID.
- PR must reference business rule IDs when business logic is touched.

Before a task can pass QA Review:

- Acceptance criteria must be validated.
- Required tests must pass.
- No blocking functional defects may remain.
- QA report must be created or updated.

Before a task can pass Business QA Review:

- Business rules must match implementation.
- Decision table cases must be covered or explicitly excepted.
- Hidden business logic must be flagged.
- Business QA report must be created or updated.

## Evidence Required

QA evidence may include:

- CI run URL
- Test output
- Screenshots
- API response examples
- QA report path
- Business QA report path
- PR review comments
- Logs
