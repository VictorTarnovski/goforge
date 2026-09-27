---
name: closure-table
description: >-
  Implementa hierarquias em SQL com o padrão Closure Table (Bill Karwin): schema
  (ancestor, descendant, length), inserção de raízes e filhos, queries de
  ancestrais/descendentes, mover e apagar subárvores. Use ao modelar comentários
  encadeados, categorias, pastas, org charts, ou quando o utilizador mencionar
  closure table, árvore hierárquica, ancestor/descendant, ou o slide de Bill
  Karwin "Models for hierarchical data".
user-invocable: true
---

# Closure Table (Bill Karwin)

**Persona:** Engenheiro backend que modela árvores em SQL com integridade referencial, sem recursão em queries.

**Fontes:** [Models for hierarchical data](https://pt.slideshare.net/billkarwin/models-for-hierarchical-data) · [Rendering Trees with Closure Tables](https://karwin.com/blog/index.php/2010/03/24/rendering-trees-with-closure-tables/) · [Moving Subtrees (Percona)](https://www.percona.com/blog/moving-subtrees-in-closure-table/)

## Escolher o modelo

| Modelo | Prós | Contras |
|--------|------|---------|
| **Closure Table** | Queries simples, FKs, sem recursão | Mais linhas; mover subárvore é mais trabalhoso |
| Adjacency list (`parent_id`) | Simples | Recursão ou múltiplas queries |
| Path enumeration | Mover subárvore fácil (string replace) | Sem FKs no caminho; paths longos |
| Nested sets | Leitura de subárvore rápida | Updates caros; sem FKs |

**Usar Closure Table** quando consultas a ancestrais/descendentes forem frequentes e integridade referencial importa. **Evitar** se mover subárvores for a operação dominante.

## Workflow: criar uma closure table

```
- [ ] 1. Tabela de nós (entidade)
- [ ] 2. Tabela {entidade}_paths (closure)
- [ ] 3. Índice em descendant
- [ ] 4. Funções Insert*ClosureRoot / Insert*ClosureChild
- [ ] 5. Queries comuns (filhos, pai, raízes)
- [ ] 6. Inserir nó + paths na mesma transacção
- [ ] 7. Testes
```

### 1. Schema

Duas tabelas: nós + todos os caminhos ancestor→descendant.

```sql
CREATE TABLE posts (
	id TEXT PRIMARY KEY,
	content TEXT NOT NULL,
	-- … atributos do nó
	author_id TEXT NOT NULL REFERENCES users (id),
	created_at TEXT NOT NULL
);

CREATE TABLE post_paths (
	ancestor TEXT NOT NULL REFERENCES posts (id),
	descendant TEXT NOT NULL REFERENCES posts (id),
	length INTEGER NOT NULL,
	PRIMARY KEY (ancestor, descendant)
);

CREATE INDEX idx_post_paths_descendant ON post_paths (descendant);
```

- **`length`**: 0 = reflexivo (nó consigo), 1 = filho directo, 2 = neto, …
- **PK `(ancestor, descendant)`**: cada par de caminho é único.
- **Índice em `descendant`**: acelera “todos os meus ancestrais”.

Respeitar regras do projecto: `sql-insert-column-names`, `sql-fk-column-order`.

### 2. Inserir nós

**Raiz** — só caminho reflexivo:

```sql
INSERT INTO post_paths (ancestor, descendant, length) VALUES (?, ?, 0);
```

**Filho** — copiar caminhos do pai, incrementar `length`, adicionar reflexivo:

```sql
INSERT INTO post_paths (ancestor, descendant, length)
SELECT t.ancestor, ?, t.length + 1
FROM post_paths AS t
WHERE t.descendant = ?
UNION ALL
SELECT ?, ?, 0;
```

Sempre na **mesma transacção** que o `INSERT` do nó.

### 3. Queries comuns

| Operação | SQL |
|----------|-----|
| Filhos directos | `ancestor = ? AND length = 1` |
| Pai directo | `descendant = ? AND length = 1` |
| Todos descendentes | `ancestor = ? AND length > 0` |
| Todos ancestrais | `descendant = ? AND length > 0` |
| Raízes (sem pai) | `NOT EXISTS (SELECT 1 FROM post_paths pp WHERE pp.descendant = posts.id AND pp.length > 0)` |

Ordenar caminhos por **`length`**, não por `id`.

### 4. Convenções deste projecto (comu)

- Tabela closure: `{entidade}_paths` (ex.: `post_paths`).
- Código em `internal/db/{entidade}.go` — **não** criar ficheiros adjacentes (`post_closure.go`).
- Funções: `Insert{Entidade}ClosureRoot`, `Insert{Entidade}ClosureChild`.
- SQL como constantes no ficheiro de domínio.
- Referência viva: `internal/db/post.go` (`InsertPostClosureRoot`, `InsertPostClosureChild`, `listDirectCommentsSQL`, `rootPostsFilterSQL`).

## Modos

- **Write mode** — nova hierarquia: seguir o workflow acima; copiar padrões de `post.go` / `post_paths`.
- **Review mode** — verificar: transacções, caminho reflexivo em cada nó, `length` correcto, índice em `descendant`, queries sem recursão CTE desnecessária.

## Operações avançadas

Mover subárvore, apagar subárvore e breadcrumbs → ver [references/operations.md](references/operations.md).
