# System Overview

## Purpose

Describe the technical purpose of the system.

## Product Context

Summarize the product from a technical perspective.

## Architecture Goals

- Maintainable
- Testable
- Secure
- Observable
- Scalable enough for the expected product stage
- Business logic is explicit and traceable

## Non-Goals

What this architecture intentionally does not support yet.

## System Context Diagram

Describe the system and its external actors.

```text
User / External System
        ↓
Application
        ↓
Database / External Services

Main Modules
Module	Responsibility	Owns Business Logic?	Notes
CHANGE_ME	CHANGE_ME	Yes / No	CHANGE_ME

Layering Rules

The system should follow this logical separation:
Interface Layer
  - UI
  - API routes
  - Controllers

Application Layer
  - Use cases
  - Workflow orchestration
  - Transaction boundaries

Domain Layer
  - Business entities
  - Business rules
  - Policies
  - Domain services

Infrastructure Layer
  - Database access
  - External APIs
  - File storage
  - Email/SMS/payment providers

  Business Logic Location

Core business logic must live in the domain or business layer.

Business logic must not be scattered across:

UI components
API controllers
Database triggers
Background jobs
Random utility files
Hardcoded frontend checks

Frontend checks may improve user experience, but backend/domain logic is the final authority.

Data Ownership

Describe which module owns which data.

Data	Owning Module	Notes
CHANGE_ME	CHANGE_ME	CHANGE_ME
External Integrations
Integration	Purpose	Criticality	Failure Behavior
CHANGE_ME	CHANGE_ME	Low / Medium / High	CHANGE_ME

Authentication

Describe how users or services authenticate.

Authorization

Describe how permissions are enforced.

Observability

The system should define:

Logs
Metrics
Traces
Audit events
Business decision events
Failure Modes
Failure	Expected Behavior	User Impact	Recovery
CHANGE_ME	CHANGE_ME	CHANGE_ME	CHANGE_ME

Security Considerations
Authentication
Authorization
Secrets
Input validation
Data exposure
Audit logging
Dependency risk
Architecture Risks
Risk	Impact	Mitigation
CHANGE_ME	CHANGE_ME	CHANGE_ME

Open Technical Questions
OPEN QUESTION:
