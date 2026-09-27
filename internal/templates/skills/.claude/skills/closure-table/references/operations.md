# Closure Table — operações avançadas

## Mover subárvore

Para mover nó `D` de baixo de `B` para baixo de `N`:

1. Apagar caminhos que ligam ancestrais **fora** da subárvore a descendentes **dentro** da subárvore.
2. Inserir novos caminhos: produto cartesiano de ancestrais de `N` (incl. `N`) × descendentes de `D` (incl. `D`).

```sql
-- Inserir novos caminhos após detach (simplificado; ver Percona para DELETE completo)
INSERT INTO post_paths (ancestor, descendant, length)
SELECT supertree.ancestor, subtree.descendant,
       supertree.length + subtree.length + 1
FROM post_paths AS supertree
JOIN post_paths AS subtree
WHERE subtree.ancestor = ?
  AND supertree.descendant = ?;
```

Referência completa: [Moving Subtrees in Closure Table](https://www.percona.com/blog/moving-subtrees-in-closure-table/).

## Apagar subárvore

Depende da política de domínio:

- **Hard delete:** `DELETE FROM post_paths WHERE ancestor IN (SELECT descendant FROM post_paths WHERE ancestor = ?)` + apagar nós.
- **Soft delete:** marcar nós com `deleted_at`; filtrar em queries de leitura (padrão actual em `posts`).

## Breadcrumbs (caminho do nó à raiz)

Bill Karwin usa `GROUP_CONCAT` (MySQL). Em SQLite/PostgreSQL, juntar labels ordenados por `length`:

```sql
-- SQLite: filhos directos de um nó (padrão do projecto)
SELECT p.*
FROM posts p
INNER JOIN post_paths pp ON pp.descendant = p.id
WHERE pp.ancestor = ? AND pp.length = 1
ORDER BY p.created_at DESC;
```

Para breadcrumb completo, juntar `post_paths` a si mesma filtrando `descendant = ?`, ordenar por `length ASC` (raiz primeiro) ou `DESC` (folha primeiro).

## Exemplo de dados (Bill Karwin)

Árvore `root → mid → core → leaf`:

```
(1,1,0) (1,2,1) (1,3,1) (1,4,2) (1,5,3)
(2,2,0)
(3,3,0) (3,4,1) (3,5,2)
(4,4,0) (4,5,1)
(5,5,0)
```

Cada nó tem caminho reflexivo `(n,n,0)`. Cada aresta pai→filho gera linhas para **todos** os ancestrais do pai até ao filho.
