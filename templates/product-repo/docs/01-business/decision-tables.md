# Decision Tables

Decision tables convert business rules into clear, testable examples.

They help Product, Engineering, QA, Business QA, and AI agents agree on expected behavior.

---

## Decision Table Template

### Rule ID

BR-001

### Rule Name

CHANGE_ME

| Case ID | Scenario | Input / Conditions | Expected Decision | Expected Reason | Notes |
|---|---|---|---|---|---|
| BR-001-DT-001 | Positive case | CHANGE_ME | Allowed |  |  |
| BR-001-DT-002 | Negative case | CHANGE_ME | Rejected | CHANGE_ME |  |
| BR-001-DT-003 | Boundary case | CHANGE_ME | CHANGE_ME | CHANGE_ME |  |
| BR-001-DT-004 | Exception case | CHANGE_ME | CHANGE_ME | CHANGE_ME |  |

---

## Decision Table Rules

Every important business rule should have:

- At least one positive case.
- At least one negative case.
- Boundary cases when numbers, dates, limits, permissions, money, or time are involved.
- Exception cases when special users, statuses, countries, plans, roles, or conditions behave differently.
- Clear expected reasons for rejected decisions.
