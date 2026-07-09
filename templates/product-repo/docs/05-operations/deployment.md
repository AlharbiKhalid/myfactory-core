# Deployment

## Purpose

This document defines how the product is deployed safely.

## Environments

| Environment | Purpose | Deployment Method | Notes |
|---|---|---|---|
| local | Local development | Manual | Developer or agent environment |
| development | Shared development testing | CHANGE_ME | CHANGE_ME |
| staging | Pre-production validation | CHANGE_ME | CHANGE_ME |
| production | Live customer environment | CHANGE_ME | CHANGE_ME |

## Deployment Principles

- Main branch must be protected.
- Deployment should happen only from approved branches or release tags.
- CI must pass before deployment.
- QA and Business QA must pass for business-critical changes.
- Release notes must exist before production release.
- Rollback plan must exist before production release.
- Agents must not deploy to production without explicit approval.

## Release Checklist

Before release:

- [ ] Code merged to main.
- [ ] CI passed.
- [ ] Required QA passed.
- [ ] Required Business QA passed.
- [ ] Database migrations reviewed.
- [ ] Environment variables verified.
- [ ] Secrets verified.
- [ ] Monitoring checks prepared.
- [ ] Rollback plan prepared.
- [ ] Release notes prepared.
- [ ] Stakeholders notified if needed.

## Deployment Process

### Step 1: Prepare Release

Describe how a release is prepared.

### Step 2: Deploy to Staging

Describe staging deployment.

### Step 3: Run Smoke Tests

Describe smoke tests.

### Step 4: Approve Production Release

Describe approval process.

### Step 5: Deploy to Production

Describe production deployment.

### Step 6: Monitor After Release

Describe post-release monitoring.

## Smoke Test Plan

| Test | Expected Result | Owner |
|---|---|---|
| CHANGE_ME | CHANGE_ME | CHANGE_ME |

## Rollback Plan

Describe how to safely roll back.

### Rollback Triggers

- Critical production bug.
- Data corruption risk.
- Security issue.
- Major business logic defect.
- Failed deployment health checks.

### Rollback Steps

1. CHANGE_ME
2. CHANGE_ME
3. CHANGE_ME

## Database Migration Notes

Describe database migration risks and rollback behavior.

## Release Notes Template

```text
# Release CHANGE_ME

## Summary

## Changes

## Business Rules Changed

## Migrations

## Risks

## Rollback Plan

## Monitoring Notes
