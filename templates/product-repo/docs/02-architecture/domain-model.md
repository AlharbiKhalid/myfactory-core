# Domain Model

The domain model describes the core business concepts of the system.

Agents must align implementation, tests, APIs, and documentation with this model.

## Core Entities

| Entity | Description | Key Fields | Related Business Rules |
|---|---|---|---|
| CHANGE_ME | CHANGE_ME | CHANGE_ME | BR-CHANGE_ME |

## Value Objects

| Value Object | Description | Validation Rules |
|---|---|---|
| CHANGE_ME | CHANGE_ME | CHANGE_ME |

## Domain Services

| Service | Responsibility | Related Rules |
|---|---|---|
| CHANGE_ME | CHANGE_ME | BR-CHANGE_ME |

## Business Policies

| Policy | Purpose | Related Rules |
|---|---|---|
| CHANGE_ME | CHANGE_ME | BR-CHANGE_ME |

## State Machines

### State Machine: CHANGE_ME

| Current State | Event | Next State | Rule |
|---|---|---|---|
| CHANGE_ME | CHANGE_ME | CHANGE_ME | BR-CHANGE_ME |

## Invariants

Invariants are conditions that must always be true.

| Invariant | Related Entity | Related Rule |
|---|---|---|
| CHANGE_ME | CHANGE_ME | BR-CHANGE_ME |

## Domain Events

| Event | Trigger | Consumers | Notes |
|---|---|---|---|
| CHANGE_ME | CHANGE_ME | CHANGE_ME | CHANGE_ME |

## Domain Language Rules

Code should use business language from:
```text
docs/01-business/glossary.md

Avoid vague names like:

processData
handleThing
type = 3
status = 7

Prefer domain-specific names like:

approveRefund
calculateEligibility
AppointmentStatus.CONFIRMED
CustomerType.VIP
