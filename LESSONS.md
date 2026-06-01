# Lessons Learned

> Grows with every bug fixed, every task completed, every mistake made.
> This is the agent's evolving intelligence — not just what we built, but what we learned.

---

## 2026-05-28: Rate Limiter 429 False Blocking

**Symptom**: All users blocked after 3 registration attempts in Docker.
**Root cause**: `c.ClientIP()` sees gateway IP, not real client IP, in multi-hop Docker setup.
**Fix**: Forward `X-Forwarded-For` + `X-Real-IP` at every hop: Next.js proxy → gateway ratelimiter → auth handler.
**Files**: `apps/web/src/app/api/v1/[...path]/route.ts`, `services/gateway/internal/ratelimit/ratelimit.go`, `services/auth/internal/transport/http/handler.go`
**Pattern**: In Docker/K8s, NEVER trust `c.ClientIP()` alone. Always check forwarded headers.

---

## 2026-05-28: identity-auth user_id Column Mismatch

**Symptom**: `SchemaManagementException: missing column [user_id] in table [users]`
**Root cause**: Java entity used `user_id` but Go schema + DB used `id`. Flyway skipped existing table.
**Fix**: Align entity field names with actual DB schema: `userId` → `id`.
**Pattern**: Cross-language monorepo — always verify entity↔schema↔actual DB alignment.

---

## 2026-07-17: Context Budget — The #1 Agent Killer

**Lesson**: Accumulated 150-198K tokens per session causing HTTP 500. Root cause was behavioral, not config.
**Rules learned**:
1. Simple question → answer directly, zero tool calls
2. Complex task → delegate_task immediately, main agent = thin orchestrator
3. Browser in main agent = instant 10-30K token burn, always delegate
4. Read files with offset/limit, never full file
5. Sub-agent context auto-discards after return — only summary enters parent
6. Memory persists independently of conversation history

---

## Patterns: Frontend Store-to-UI Wiring Gaps

Common issues found during audits:
1. Server component needs client store data → hydration mismatch
2. Header uses hardcoded values instead of Zustand store selectors
3. Search input not wrapped in `<form>` → no Enter key submit
4. `alert()` instead of toast notifications
5. Redundant recalculation when mapper already provides the field
6. `mutationFn` params include unused fields → bloated request bodies
7. `useUser` staleTime too low combined with initialData → excessive refetches

---

## Go Microservice Patterns

- `patch` tool can corrupt Go files → always `go build` after each patch
- sonic v1.12.4 incompatible with Go 1.26.3 → use v1.15.1
- K8s YAML: never use multi-document (`---`) in a single file
- SQL escaping: `''` not `\'`
- DB enum: UPPERCASE
- Prices: int64 (BIGINT), weight: int, currency: VND
- Domain errors: `*DomainError` pointers
- MySQL: host=mysql-primary, port=3306, user=tiki, pass=tiki_dev
