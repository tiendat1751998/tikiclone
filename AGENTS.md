## Session Startup
1. Read `TASK_LOG.md` (active-only), scan `LESSONS.md` for patterns.
2. Read `PROJECT_BRIEF.md` only if architecture context is needed.

## Autonomous Workflow (MANDATORY)
Execute this loop per logical unit of work. Do not ask for confirmation between steps.
1. **Analyze** — Read source, check patterns, identify root cause. Use Explore sub-agents (≤8K tokens) for broad codebase sweeps.
2. **Plan** — Brief plan. State assumptions to user for course correction. If ambiguity risks catastrophic failure, ask once before executing.
3. **Execute** — **Coder** (write minimal focused code) or **Fixer** (diagnose root cause first, then patch smallest scope). Run complex reasoning via `sequentialthinking`. All code edits in main session; browser and heavy research in sub-agents.
4. **Verify** — Run build/lint/test. Fix breakages via Fixer workflow.
5. **Review** — Orphaned imports, dead code, cross-repo contract mismatches.
6. **Report** — Minimal summary: what was done, verification evidence, risks. Ask for deployment approval.

## Coding Rules
- **Simplicity:** Minimum code, zero speculative abstractions. 200 lines that could be 50 → rewrite.
- **Surgical:** Touch only what you must. Match existing style. Every changed line traces to the request.
- **Local Abstraction:** Extract helpers only if duplication risks bugs after an update. Otherwise keep it local.
- **Contract Alignment:** Run global type-checkers/code-gen before final build when altering shared types, schemas, or protos.

## Verification & Self-Correction
- **Atomic Guardrail:** Run build/lint/test after each logical unit. Defer only for synchronized multi-file breaking changes. Never leave a broken build.
- **Pivot Limits:** Max 3 attempts per strategy, 2 pivots per task. If still failing: stop, dump logs, exit.
- **Fail-Fast:** If a tool/patch fails, change strategy immediately. Never retry same args.

<!-- NEW SECTION START -->
## Session Lifecycle & Sub-Agent Architecture
- **Context Reset Protocol:** After the Report step, signal termination and spawn a fresh session for the next task. Carry forward only `TASK_LOG.md` and `LESSONS.md`. Never re-use a bloated session.
- **Sub-Agent Ephemeral Guarantee:** Sub-agent contexts are destroyed on completion. Only the structured summary survives; raw tool output, logs, and history never pollute the main context.
- **Sub-Agent Types & Token Budget:**
  - **Explore** (≤8K): Read-only research. Returns structured findings (key files, patterns, decisions).
  - **Plan** (≤4K): Task decomposition. Returns prioritized steps with file-level scope.
  - **Implement** (≤80K): Code edits in isolated context. Returns diff + verification evidence.
  - **Review** (≤6K): Post-task validation. Checks regressions, orphans, contracts.
- **Sub-Agent Output Contract (JSON):**
  ```json
  {
    "status": "success|failed|partial",
    "findings": ["key finding 1", "key finding 2"],
    "files_changed": ["path/to/file"],
    "verification": "build/lint/test results summary",
    "risks": ["remaining concerns"],
    "handoff_note": "context for next step"
  }
