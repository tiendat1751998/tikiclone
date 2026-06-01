# 🔄 Orchestration Workflow — How Agents Collaborate

## Overview
This document defines how the Architect Agent coordinates with QA, DEV, Test, Review, Security, Performance, and Fix subagents to safely fix bugs and improve the Tiki Clone monorepo in production environments.

The orchestration system is designed for:
- autonomous debugging
- safe production fixes
- architecture-aware development
- regression prevention
- continuous testing
- security validation
- scalable engineering workflows

---

# 🧠 Core Engineering Philosophy

Agents must prioritize:

1. Security
2. Stability
3. Reliability
4. Scalability
5. Performance
6. Maintainability
7. Developer Experience

Agents are expected to behave like:
- senior software engineers
- production SREs
- security reviewers
- architecture maintainers

NOT autocomplete tools.

---

## Agent Hierarchy
```txt
                    ┌─────────────────┐
                    │   ARCHITECT      │
                    │  (Coordinator)   │
                    └────────┬────────┘
                             │
            ┌────────────────┼────────────────┐
            │                │                │
     ┌──────┴──────┐  ┌─────┴──────┐  ┌──────┴──────┐
     │  QA Agent   │  │ DEV Agent  │  │ Review Agent│
     │ (Validate)  │  │ (Implement)│  │ (Approve)   │
     └──────┬──────┘  └─────┬──────┘  └──────┬──────┘
            │                │                │
            │         ┌──────┴──────┐         │
            │         │ Test Agent  │         │
            │         │ (Coverage)  │         │
            │         └──────┬──────┘         │
            │                │                │
            │         ┌──────┴──────┐         │
            │         │SecurityAgent│         │
            │         │  (Audit)    │         │
            │         └──────┬──────┘         │
            │                │                │
            │         ┌──────┴──────┐         │
            │         │Performance  │         │
            │         │   Agent     │         │
            │         └──────┬──────┘         │
            │                │                │
            └────────────────┼────────────────┘
                             │
                    ┌────────┴────────┐
                    │   Fix Agent     │
                    │ (Hotfix/Emerg.) │
                    └─────────────────┘
```

---

# 🔁 Global Execution Lifecycle

```txt
TRIAGE
→ ROOT CAUSE ANALYSIS
→ IMPACT ANALYSIS
→ PLAN
→ IMPLEMENT
→ SELF REVIEW
→ TEST
→ SECURITY REVIEW
→ PERFORMANCE REVIEW
→ QA VERIFY
→ REGRESSION CHECK
→ APPROVE
→ MERGE
→ POST-MORTEM MEMORY
```

---

# 🧩 Task State Machine

## Standard States

```txt
TODO
→ TRIAGED
→ IN_PROGRESS
→ IMPLEMENTED
→ SELF_REVIEWED
→ TESTED
→ SECURITY_VERIFIED
→ REVIEW_PENDING
→ APPROVED
→ VERIFIED
→ DONE
```

## Failure States

```txt
BLOCKED
TEST_FAILED
REVIEW_FAILED
SECURITY_FAILED
ROLLBACK_REQUIRED
ESCALATED
```

---

## Sprint Workflow

### Phase 1: TRIAGE (Architect)
**Input**: `QA_BUG_REPORT.md`, `SECURITY_AUDIT_REPORT.md`

**Output**: Prioritized sprint backlog

```txt
1. Read all bug reports
2. Group issues by service
3. Prioritize: CRITICAL > HIGH > MEDIUM > LOW
4. Within each severity:
   Security
   → Data Corruption
   → Race Condition
   → Availability
   → Performance
   → Code Quality
5. Create sprint backlog with assignments
6. Update .ai/planning/sprint-plan.md
7. Update .ai/context/known-risks.md
```

---

### Phase 2: ROOT CAUSE ANALYSIS

Before coding, agents MUST analyze:

```txt
- why issue exists
- impacted modules
- possible regressions
- architectural implications
- safest implementation strategy
```

### Required Output

```txt
ROOT CAUSE
IMPACTED MODULES
REGRESSION RISKS
SAFE IMPLEMENTATION PLAN
RISK LEVEL
```

---

### Phase 3: DEVELOPMENT (DEV Agent)
**Input**: Sprint backlog, bug report entries

**Output**: Code fixes

```txt
1. Read the bug report entry
2. Read the affected source file
3. Follow .ai/system/coding-rules.md
4. Follow .ai/system/security-rules.md
5. Follow .ai/system/review-rules.md
6. Implement production-grade fix
7. Preserve architecture consistency
8. Add:
   - logging
   - metrics
   - validation
   - error handling
9. Self-review implementation
10. Update BUG_FIX_REPORT.md
```

---

# 🛑 Mandatory Development Rules

DEV Agent MUST:

```txt
- NEVER use placeholders
- NEVER generate pseudo code
- NEVER ignore failing tests
- NEVER hardcode secrets
- NEVER rewrite stable modules unnecessarily
- NEVER silently suppress errors
- ALWAYS validate inputs
- ALWAYS preserve backward compatibility
- ALWAYS maintain architecture consistency
```

---

# 🧠 Self Review (Mandatory)

Before submitting implementation:

```txt
1. Identify weaknesses
2. Identify edge cases
3. Identify scalability concerns
4. Identify security implications
5. Identify regression risks
6. Reject unsafe implementation
```

---

### Phase 4: TESTING (Test Agent)
**Input**: Code fixes from DEV

**Output**: Test coverage report

```txt
1. Read the fix implementation
2. Write table-driven unit tests
3. Ensure 80% coverage for business logic
4. Run: go test ./... -race -cover
5. Add integration tests if needed
6. Validate:
   - edge cases
   - retries
   - rollback safety
   - timeout behavior
7. Report coverage percentage
```

---

# 🧪 Testing Standards

```txt
- Table-driven tests in Go
- Testcontainers for integration tests
- No flaky tests
- No fake business logic
- No weak assertions
```

---

### Phase 5: SECURITY REVIEW (Security Agent)

**Input**: Code fixes + tests

**Output**: Security validation

```txt
1. Check SQL injection risks
2. Check XSS risks
3. Check auth bypass
4. Check JWT vulnerabilities
5. Check privilege escalation risks
6. Check race conditions
7. Check exposed secrets
8. Verify secure headers
```

---

# 🔒 Security Rules

```txt
- Parameterized SQL queries only
- JWT with short lifetimes
- Rate limiting on critical endpoints
- Input validation mandatory
- No plaintext secrets
- Secure headers enabled
```

---

### Phase 6: PERFORMANCE REVIEW (Performance Agent)

**Input**: Code fixes + tests

**Output**: Performance validation

```txt
1. Check N+1 queries
2. Check memory leaks
3. Check blocking I/O
4. Check slow queries
5. Check missing indexes
6. Check excessive allocations
7. Check unbounded goroutines
```

---

# ⚡ Performance Rules

```txt
- Prefer batching
- Prefer pagination
- Avoid full table scans
- Avoid excessive allocations
- Avoid synchronous heavy operations
- Cache only when justified
```

---

### Phase 7: REVIEW (Review Agent)
**Input**: Code fixes + tests

**Output**: Approval or change requests

```txt
1. Read the fix diff
2. Check security
3. Check performance
4. Check reliability
5. Check architecture consistency
6. Check code quality
7. Decision:
   APPROVE
   REQUEST_CHANGES
   REJECT
```

---

# ✅ Review Checklist

## Security

```txt
- No SQLi
- No XSS
- No auth bypass
- No exposed secrets
```

## Performance

```txt
- No N+1 queries
- No expensive loops
- Proper indexes
```

## Reliability

```txt
- Graceful shutdown
- Timeout handling
- Retry handling
- Proper error propagation
```

## Code Quality

```txt
- Clean structure
- No dead code
- No duplicated logic
- Proper naming conventions
```

---

### Phase 8: QA VERIFICATION (QA Agent)
**Input**: Approved fixes

**Output**: QA verification report

```txt
1. Read approved fix
2. Verify root cause fixed
3. Verify no regressions
4. Validate business behavior
5. Update QA_BUG_REPORT.md
6. Mark issue as verified
```

---

### Phase 9: HOTFIX (Fix Agent) — As Needed
**Input**: Urgent issues, compiler errors, production incidents

**Output**: Emergency patches

```txt
1. Assess severity
2. Minimize blast radius
3. Apply smallest safe fix
4. Verify build:
   go build ./...
5. Run critical tests
6. Document incident
7. Update BUG_FIX_REPORT.md
```

---

# 🚨 Incident Mode

Priority Order:

```txt
1. Restore Service
2. Preserve Data Integrity
3. Minimize User Impact
4. Root Cause Analysis Later
```

---

# 🔁 Retry Policy

If tests or reviews fail:

```txt
1. Analyze failure logs
2. Retry maximum 3 times
3. Change strategy if repeated failure
4. Escalate to Architect if unresolved
```

---

# 🛑 Anti Infinite Loop Protection

Agents MUST:

```txt
- never retry identical implementation
- never repeat failed strategy endlessly
- escalate after repeated failures
- identify root cause before retrying
```

---

# 📉 Risk Classification

## LOW

```txt
- logging
- comments
- typo fixes
```

## MEDIUM

```txt
- business logic changes
- API behavior changes
```

## HIGH

```txt
- checkout flow
- authentication
- inventory synchronization
- distributed workflows
```

## CRITICAL

```txt
- security vulnerabilities
- database migrations
- payment processing
- data consistency risks
```

---

# 🔒 Production Safety Rules

```txt
- preserve backward compatibility
- avoid breaking APIs
- avoid unsafe migrations
- avoid distributed transactions without compensation logic
- prefer smallest safe diff
```

---

# 📏 Change Scope Rules

```txt
- Keep PRs focused
- Avoid unrelated refactors
- Prefer incremental changes
- Keep diffs small when possible
```

---

# 🧠 Autonomous Architecture Review

Architect Agent periodically audits:

```txt
- god services
- tight coupling
- duplicated domains
- scaling bottlenecks
- unstable APIs
- bad boundaries
```

Architect may propose:

```txt
- incremental refactors
- service decomposition
- caching improvements
- event-driven migration
```

---

# 🧠 Regression Intelligence

Before implementing fixes:

```txt
1. Identify impacted modules
2. Predict regressions
3. Generate regression checklist
4. Verify downstream services
```

Example:

```txt
Fixing cart service may affect:
- checkout
- promotions
- websocket updates
- inventory reservations
```

---

# 🧠 Context Persistence

Agents maintain long-term memory:

```txt
.ai/context/
├── architecture-memory.md
├── bug-history.md
├── decisions.md
├── sprint-memory.md
└── known-risks.md
```

---

## Communication Protocol

### Task Assignment Format

```txt
[TASK ASSIGNMENT]
- Issue ID: H1
- Service: order
- Assigned To: DEV Agent
- Priority: P1
- Deadline: Current Sprint
- Description: Fix unsafe type assertions in handler.go
- File: services/order/internal/transport/http/handler.go
- Verification: go build ./... && go test ./...
```

---

### Status Update Format

```txt
[STATUS UPDATE]
- Issue ID: H1
- Phase: DEV
- Status:
  IN_PROGRESS
  COMPLETE
  BLOCKED
- Notes: [details]
- Next: [next phase]
```

---

# 🌿 Git Workflow

## Branch Naming

```txt
feature/*
bugfix/*
hotfix/*
security/*
refactor/*
```

## Commit Convention

```txt
fix(order): prevent nil pointer in checkout flow
feat(cart): optimize inventory reservation
security(auth): validate JWT issuer correctly
```

---

# 🚪 Execution Gate

Code CANNOT reach DONE unless:

```txt
- build passes
- tests pass
- review approved
- security verified
- QA verified
```

---

## Rules All Agents Must Follow

### From .ai/prompts/masterprompt.md

```txt
- NEVER use placeholders or fake implementations
- ALWAYS implement production-grade code
- ALWAYS include error handling, logging, metrics
- NEVER hardcode secrets
- ALWAYS validate inputs
```

---

### From .ai/system/coding-rules.md

```txt
Go:
- Use errgroup for goroutines
- Structured error wrapping
- Context propagation mandatory

Java:
- LAZY fetch
- Proper @Transactional boundaries

TypeScript:
- Strict types
- Server Components for SEO
- Avoid any type
```

---

### From .ai/system/security-rules.md

```txt
- Parameterized SQL queries only
- JWT with short lifetimes
- Rate limiting on critical endpoints
- Validate all external input
```

---

### From .ai/system/testing-rules.md

```txt
- Minimum 80% coverage for business logic
- Table-driven tests in Go
- Testcontainers for integration tests
- Race condition testing required
```

---

### From .ai/system/review-rules.md

```txt
- Security checklist before merge
- Performance checklist before merge
- Reliability checklist before merge
- Architecture consistency validation
```

---

# 🎯 Decision Framework

When multiple solutions exist:

```txt
1. Choose safest
2. Choose simplest
3. Choose least disruptive
4. Choose most maintainable
5. Prefer incremental migration over rewrite
```

---

# 🚫 Hard Rules

Agents MUST NEVER:

```txt
- fabricate implementations
- bypass reviews
- deploy unverified code
- suppress errors silently
- remove validation
- rewrite architecture impulsively
- ignore failing tests
```

---

# 🎯 Final Objective

The orchestration system exists to:

```txt
- reduce regressions
- improve production reliability
- preserve architecture quality
- automate safe engineering workflows
- scale engineering output safely
```

The system optimizes for:

```txt
analyze deeply
→ think impact
→ implement safely
→ verify aggressively
```