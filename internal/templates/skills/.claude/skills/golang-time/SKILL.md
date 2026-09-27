---
name: golang-time
description: >-
  Boas práticas de time.Time em Go para domínio, API e database/sql (sem sqlx/sqlc).
  Use ao modelar timestamps (CreatedAt, UpdatedAt, DeletedAt, StartTime, EndTime),
  serializar JSON, persistir em SQLite/Postgres, ou quando o utilizador mencionar
  RFC3339, UTC, timestamps ou conversão string↔time.
user-invocable: true
---

# time.Time (COMU backend)

**Persona:** Engenheiro Go que trata instantes como `time.Time` no domínio e converte para string **só** na fronteira com HTTP/SQL.

**Fontes:** [pkg.go.dev/time](https://pkg.go.dev/time) · [pkg.go.dev/database/sql](https://pkg.go.dev/database/sql) · [RFC 3339](https://www.rfc-editor.org/rfc/rfc3339)

## Regra de ouro

| Camada | Tipo | Formato |
|--------|------|---------|
| **Domínio** (`internal/domain`) | `time.Time` | UTC (`domain.Now()`, `.UTC()`) |
| **API response** | `time.Time` nos DTOs de saída | JSON RFC3339 automático |
| **API request** | `string` só no DTO de entrada | Converter logo para `time.Time` |
| **Infra** (`internal/infra/sql`) | `string`/`sql.NullTime` no `Scan`/`Exec` | RFC3339 na coluna TEXT |

**Nunca** `string`, `*string` ou helpers `NowString()` para campos de instante no domínio.

## Domínio

```go
type Post struct {
    CreatedAt time.Time `json:"created_at"`
    DeletedAt time.Time `json:"-"` // zero = não apagado
}

func (p Post) IsDeleted() bool {
    return !p.DeletedAt.IsZero()
}

// Em NewPost / factories:
CreatedAt: Now(), // domain.Now() => time.Now().UTC()
```

- **Opcional / nullable:** `time.Time` zero (`time.Time{}`), não ponteiro, não string vazia.
- **Comparações:** `t.Equal(u)` ou `!t.IsZero()` — evitar `==` entre `time.Time` de origens diferentes (location/monotonic).
- **Parse de input externo:** `time.Parse(time.RFC3339, strings.TrimSpace(s))` + normalizar com `.UTC()`.
- **Sem tags `db:`** — este projeto usa `database/sql` com `Scan` explícito.

## API

```go
// Request: strings camelCase; converter no handler
type createGatheringRequest struct {
    StartTime string  `json:"startTime"`
    EndTime   *string `json:"endTime"`
}

// Response: time.Time — encoding/json emite RFC3339
type postResponse struct {
    CreatedAt time.Time `json:"created_at"`
}
```

- Preferir `time.Time` nos DTOs de **resposta**.
- Validar formato inválido no domínio (`ErrInvalidStartTime`, etc.) → 400.

## Infra (`database/sql`)

Conversão **apenas** em repositórios — não vazar strings para o domínio.

```go
func formatStoredTime(t time.Time) string {
    return t.UTC().Format(time.RFC3339)
}

func parseStoredTime(value string) (time.Time, error) {
    return time.Parse(time.RFC3339, value)
}

func scanPost(row rowScanner, p *domain.Post) error {
    var createdAt string
    var deletedAt sql.NullTime // ou sql.NullString + parse
    // Scan → parseStoredTime → p.CreatedAt / p.DeletedAt
}

func (r *SQLPostRepository) Create(ctx context.Context, post *domain.Post) error {
    _, err := r.db.ExecContext(ctx, insertSQL,
        formatStoredTime(post.CreatedAt),
    )
}

// Coluna NULL → zero no domínio
if deletedAt.Valid {
    p.DeletedAt = deletedAt.Time.UTC()
}
```

- Colunas SQLite `TEXT`: gravar RFC3339 UTC.
- `NULL` na BD → `time.Time{}` no domínio.
- Timestamps gerados no servidor: `domain.Now()` no domínio; infra só formata na escrita.

## UTC e RFC3339

1. **Persistir e calcular em UTC** — evita bugs entre ambientes/containers.
2. **RFC3339** para API e colunas TEXT — é o layout default de `json.Marshal(time.Time)` e cobre ~57% dos formatos explícitos em código Go (stdlib).
3. **Não** usar layouts ad hoc (`"2006-01-02"`) para timestamps de API/BD sem razão forte.
4. **Não** assumir timezone local em `time.Parse` sem layout com offset — preferir RFC3339 ou `ParseInLocation(..., time.UTC)`.

## Anti-padrões

```go
// ❌ string no domínio
CreatedAt string `json:"created_at"`

// ❌ ponteiro string nullable
DeletedAt *string

// ❌ helper string
func NowString() string { return time.Now().Format(time.RFC3339) }

// ❌ comparar timestamps como strings
if post.CreatedAt > cursor { ... }

// ❌ db tags (sqlx/sqlc — não usamos)
CreatedAt time.Time `db:"created_at"`
```

## Checklist (review)

- [ ] Campos `*At` / `*Time` no domínio são `time.Time`
- [ ] Opcionais usam zero value, não string/ponteiro
- [ ] `domain.Now()` para “agora” no servidor
- [ ] Infra faz parse/format RFC3339; domínio não conhece layout SQL
- [ ] DTOs de resposta expõem `time.Time`; requests convertem cedo
- [ ] Testes usam `time.Parse(time.RFC3339, "...")`, não strings soltas no struct
