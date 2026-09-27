---
name: comu-slice-and-sql-args
description: "Project guardrail against two footguns that caused a real 500: make([]T, n) + append (leading zero-values / doubled length), and SQL placeholder/argument count mismatch. Use when building slices with append, or when writing IN-clause / parameterized queries."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code or similar AI coding agents, and for this Go project (comu/backend).
metadata:
  author: comu
  version: "1.0.0"
  openclaw:
    emoji: "🛡️"
    install: []
---

> **Company guardrail.** This project skill supersedes the generic slice/SQL advice in `samber/cc-skills-golang@golang-code-style` and `@golang-database` skills for `comu/backend`.

# Slice & SQL-Args Safety

Two mistakes, from one real incident, that MUST NOT recur. Both are silent — the code compiles and looks right.

## The incident

`GET /v1/members?id=<uuid>` returned HTTP 500:

```
error="missing argument with index 1"
query="SELECT ... FROM members WHERE 1 = 1 AND id IN (?,?)"
```

One `id` in the URL, yet the query had **two** placeholders and **zero** bound args. Root cause was two separate instances of the same footgun (`MemberIDsFromQuery` and `SQLMemberRepository.List`) plus a forgotten `args...`.

## Rule 1 — `make([]T, n)` + `append` is a bug

`make([]T, n)` creates a slice of length `n` **already filled with zero values**. `append` adds *after* them, so you get leading zero-values **and** ~double the length.

```go
// ❌ BUG — returns ["", realID]: one phantom zero-value + doubled length
ids := make([]MemberID, len(rawIDs))
for _, raw := range rawIDs {
    ids = append(ids, parse(raw))
}

// ✅ length 0, capacity preallocated — append fills it correctly
ids := make([]MemberID, 0, len(rawIDs))
for _, raw := range rawIDs {
    ids = append(ids, parse(raw))
}
```

**Rule:** if you `append` to a slice, create it with `make([]T, 0, n)` (or `[]T{}`), never `make([]T, n)`.
Only use `make([]T, n)` when you assign by index (`s[i] = ...`) and never `append`.

## Rule 2 — placeholders and args must match, and args MUST be passed

A parameterized query with N `?` placeholders needs exactly N bound arguments, passed to the driver.

```go
// ❌ BUG — args built but never passed → "missing argument with index 1"
rows, err := db.QueryContext(ctx, sql)

// ✅ pass the variadic args
rows, err := db.QueryContext(ctx, sql, args...)
```

Checklist for any `IN (?)` / dynamic query:
- `expandInClause(sql, n)` where `n == len(args)` (see `internal/infra/sql/infra.go`).
- The `args` slice is built with `make([]any, 0, n)` (Rule 1 applies here too).
- The final `Query/Exec/QueryRowContext` call actually receives `args...`.
- `n > 0` before you build an `IN` clause — `IN ()` is invalid SQL.

## Checking your work

Run the bundled zero-dependency script from `backend/` before finishing a change:

```bash
bash .agents/skills/comu-slice-and-sql-args/assets/check.sh
```

It flags `make([]T, n)`-then-`append` shapes and `QueryContext(ctx, sql)` calls with no
args on an `expandInClause` query, and exits non-zero if it finds anything. No tools to
install — it is plain `grep`/`sed`.

For any query builder, a table-driven test asserting `len(args) == strings.Count(sql, "?")`
for counts `{1, 2, 3}` locks the boundary down — count 1 is where off-by-one bugs hide.

## Quick review prompt

When reviewing or writing Go in this repo, ask on every slice and every query:
> "Does this `append` to a `make([]T, n)`? Does every `?` have a matching passed arg?"
