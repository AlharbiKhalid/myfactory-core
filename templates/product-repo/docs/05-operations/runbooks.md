# Runbooks

## Purpose

Runbooks describe how to respond to operational issues.

They should be clear enough for humans or approved operations agents to follow safely.

## Incident Severity Levels

| Severity | Meaning | Example |
|---|---|---|
| SEV-1 | Critical production outage or severe data/business impact | Production unavailable |
| SEV-2 | Major functionality broken | Payments failing |
| SEV-3 | Degraded functionality | Some reports delayed |
| SEV-4 | Minor issue | Non-critical UI bug |

## General Incident Process

1. Detect issue.
2. Assign incident owner.
3. Assess severity.
4. Mitigate user/business impact.
5. Communicate status.
6. Fix root cause.
7. Verify recovery.
8. Write post-incident review.
9. Create prevention tasks.

## Runbook Template

### Runbook: CHANGE_ME

#### Symptoms

How does this issue appear?

#### Impact

Who or what is affected?

#### Detection

How is the issue detected?

#### Immediate Mitigation

1. CHANGE_ME
2. CHANGE_ME

#### Diagnosis Steps

1. CHANGE_ME
2. CHANGE_ME

#### Recovery Steps

1. CHANGE_ME
2. CHANGE_ME

#### Validation

How do we confirm the issue is resolved?

#### Escalation

Who should be contacted?

#### Related Dashboards

- CHANGE_ME

#### Related Logs

- CHANGE_ME

#### Related Business Rules

- BR-CHANGE_ME

## Common Runbooks

### Application Is Down

#### Symptoms

Users cannot access the application.

#### Immediate Checks

- [ ] Check hosting provider status.
- [ ] Check application logs.
- [ ] Check recent deployments.
- [ ] Check database availability.
- [ ] Check external dependencies.

#### Recovery Steps

1. Roll back recent deployment if correlated.
2. Restart affected service if safe.
3. Escalate if root cause is infrastructure-related.

---

### CI Is Failing

#### Symptoms

Pull requests cannot pass required checks.

#### Immediate Checks

- [ ] Check failing job.
- [ ] Check test output.
- [ ] Check dependency install.
- [ ] Check recent changes to CI configuration.

#### Recovery Steps

1. Identify failing check.
2. Reproduce locally or in agent environment.
3. Fix root cause.
4. Re-run CI.
5. Update Plane issue with evidence.

---

### Business Rule Behavior Looks Wrong

#### Symptoms

The system is technically working but business decisions appear incorrect.

#### Immediate Checks

- [ ] Identify affected business rule IDs.
- [ ] Review business-rules.yaml.
- [ ] Review decision-tables.md.
- [ ] Review business-rule-test-matrix.csv.
- [ ] Check recent PRs touching the rule.
- [ ] Check logs or audit events.

#### Recovery Steps

1. Disable feature flag if available.
2. Roll back recent business logic change if necessary.
3. Create blocking Business QA issue.
4. Add regression test for the failed case.
5. Update business rule documentation if policy was unclear.
