#!/usr/bin/env bash
# comu/backend guardrail — precise, zero-dependency check for the two footguns
# in SKILL.md. Plain grep/sed, nothing to install.
# Run from backend/:  bash .agents/skills/comu-slice-and-sql-args/assets/check.sh
# Exits non-zero on a likely violation.
set -uo pipefail

fail=0
files=$(find . -name '*.go' -not -name '*_test.go')

# Rule 1: a slice created with non-zero length (make([]T, len(x)) / make([]T, N))
# that is ALSO append()-ed to later in the same file. Index-assignment cases
# (res[i] = ...) are not flagged, because they never append.
echo "== Rule 1: make([]T, n) with a matching append =="
r1=0
for f in $files; do
  while IFS= read -r hit; do
    [ -z "$hit" ] && continue
    lineno=${hit%%:*}
    content=${hit#*:}
    var=$(printf '%s' "$content" \
      | sed -nE 's/^[[:space:]]*(var[[:space:]]+)?([A-Za-z_][A-Za-z0-9_]*)[[:space:]]*:?=[[:space:]]*make\(\[\].*/\2/p')
    [ -z "$var" ] && continue
    if grep -qE "append\([[:space:]]*${var}[[:space:]]*," "$f"; then
      echo "  $f:$lineno: $content  -> use make([]T, 0, n)"
      r1=1
    fi
  # non-zero length = second make arg begins with len( or a 1-9 digit
  done < <(grep -nE ':?=[[:space:]]*make\(\[\][^,]*,[[:space:]]*(len\(|[1-9])' "$f" 2>/dev/null || true)
done
[ "$r1" -eq 0 ] && echo "  none" || fail=1

# Rule 2: a query passed through expandInClause but called with no bound args.
# Only files that use expandInClause are inspected, so static const queries
# (e.g. listDepartmentsSQL with no placeholders) are never flagged.
echo "== Rule 2: expanded query called without args... =="
r2=0
for f in $files; do
  grep -q 'expandInClause(' "$f" 2>/dev/null || continue
  hits=$(grep -nE '\.(Query|Exec)(Row)?Context\(ctx,[[:space:]]*[A-Za-z_][A-Za-z0-9_]*\)' "$f" 2>/dev/null || true)
  if [ -n "$hits" ]; then
    printf '  %s\n' "$hits" | sed 's/$/  -> pass args.../'
    r2=1
  fi
done
[ "$r2" -eq 0 ] && echo "  none" || fail=1

if [ "$fail" -ne 0 ]; then
  echo "GUARDRAIL: potential violations above — verify before committing."
  exit 1
fi
echo "GUARDRAIL: clean."
