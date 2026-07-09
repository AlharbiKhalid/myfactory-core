# Monitoring

## Purpose

This document defines how the product is monitored after release.

Monitoring must cover both:

1. Technical system health.
2. Business behavior correctness.

## Technical Metrics

| Metric | Purpose | Alert Threshold |
|---|---|---|
| error_rate | Detect application errors | CHANGE_ME |
| request_latency | Detect slow requests | CHANGE_ME |
| uptime | Detect availability issues | CHANGE_ME |
| job_failures | Detect background job issues | CHANGE_ME |

## Business Metrics

| Metric | Purpose | Related Business Rules | Alert Threshold |
|---|---|---|---|
| CHANGE_ME | CHANGE_ME | BR-CHANGE_ME | CHANGE_ME |

## Business Decision Metrics

Business-critical decisions should be measured.

Examples:

| Decision | Metric | Related Rule | Why It Matters |
|---|---|---|---|
| CHANGE_ME | CHANGE_ME | BR-CHANGE_ME | CHANGE_ME |

## Logs

The system should log:

- Errors
- Security-relevant events
- Business-critical decisions
- External integration failures
- Background job failures
- Permission denials
- Data mutation events where needed

## Audit Events

| Event | Actor | Data Logged | Related Rule |
|---|---|---|---|
| CHANGE_ME | CHANGE_ME | CHANGE_ME | BR-CHANGE_ME |

## Alerts

| Alert | Condition | Severity | Response |
|---|---|---|---|
| CHANGE_ME | CHANGE_ME | Low / Medium / High / Critical | CHANGE_ME |

## Dashboards

| Dashboard | Purpose | Audience |
|---|---|---|
| CHANGE_ME | CHANGE_ME | CHANGE_ME |

## Post-Release Monitoring Checklist

- [ ] Application health checked.
- [ ] Error rate checked.
- [ ] Latency checked.
- [ ] Business metrics checked.
- [ ] Business decision metrics checked.
- [ ] Background jobs checked.
- [ ] External integrations checked.
- [ ] Logs checked for new errors.
