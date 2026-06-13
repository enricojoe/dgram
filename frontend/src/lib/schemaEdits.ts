import type { Column, Ref, Schema, Table } from '@/types/schema'

// Pure, immutable schema transforms used by the diagram editing UI. Each returns
// a NEW Schema; none mutate their input. They are applied through the store's
// applySchemaEdit(), which flips the sync origin to 'diagram' so the DDL regen
// runs.

function mapTables(schema: Schema, fn: (t: Table) => Table): Schema {
  return { ...schema, tables: schema.tables.map(fn) }
}

/** Generate a name not already used by an existing table. */
function uniqueTableName(schema: Schema): string {
  const used = new Set(schema.tables.map((t) => t.name))
  let i = 1
  let name = 'new_table'
  while (used.has(name)) name = `new_table_${++i}`
  return name
}

/** Generate a column name not already used within a table. */
function uniqueColumnName(table: Table): string {
  const used = new Set(table.columns.map((c) => c.name))
  let i = 1
  let name = 'column'
  while (used.has(name)) name = `column_${++i}`
  return name
}

export function addTable(schema: Schema): Schema {
  const name = uniqueTableName(schema)
  const table: Table = {
    name,
    columns: [{ name: 'id', type: 'serial', pk: true, notNull: true, autoInc: true }],
  }
  return { ...schema, tables: [...schema.tables, table] }
}

export function deleteTable(schema: Schema, name: string): Schema {
  return {
    ...schema,
    tables: schema.tables.filter((t) => t.name !== name),
    // Drop any refs touching the removed table.
    refs: schema.refs.filter((r) => r.fromTable !== name && r.toTable !== name),
  }
}

export function renameTable(schema: Schema, oldName: string, newName: string): Schema {
  const name = newName.trim()
  if (!name || name === oldName) return schema
  if (schema.tables.some((t) => t.name === name)) return schema // keep names unique
  return {
    ...schema,
    tables: schema.tables.map((t) => (t.name === oldName ? { ...t, name } : t)),
    refs: schema.refs.map((r) => ({
      ...r,
      fromTable: r.fromTable === oldName ? name : r.fromTable,
      toTable: r.toTable === oldName ? name : r.toTable,
    })),
  }
}

export function addColumn(schema: Schema, table: string): Schema {
  return mapTables(schema, (t) =>
    t.name === table
      ? { ...t, columns: [...t.columns, { name: uniqueColumnName(t), type: 'text' }] }
      : t,
  )
}

export function deleteColumn(schema: Schema, table: string, column: string): Schema {
  return {
    ...mapTables(schema, (t) =>
      t.name === table
        ? { ...t, columns: t.columns.filter((c) => c.name !== column) }
        : t,
    ),
    // Drop refs that referenced the removed column.
    refs: schema.refs.filter(
      (r) =>
        !(r.fromTable === table && r.fromColumn === column) &&
        !(r.toTable === table && r.toColumn === column),
    ),
  }
}

export function updateColumn(
  schema: Schema,
  table: string,
  column: string,
  patch: Partial<Column>,
): Schema {
  const newName = patch.name?.trim()
  return {
    ...mapTables(schema, (t) => {
      if (t.name !== table) return t
      // Reject a rename that collides with another column in the same table.
      if (newName && newName !== column && t.columns.some((c) => c.name === newName)) {
        patch = { ...patch, name: column }
      }
      return {
        ...t,
        columns: t.columns.map((c) => (c.name === column ? { ...c, ...patch } : c)),
      }
    }),
    // Keep refs in sync if the column was renamed.
    refs:
      newName && newName !== column
        ? schema.refs.map((r) => ({
            ...r,
            fromColumn: r.fromTable === table && r.fromColumn === column ? newName : r.fromColumn,
            toColumn: r.toTable === table && r.toColumn === column ? newName : r.toColumn,
          }))
        : schema.refs,
  }
}

export function togglePk(schema: Schema, table: string, column: string): Schema {
  return mapTables(schema, (t) =>
    t.name === table
      ? {
          ...t,
          columns: t.columns.map((c) =>
            c.name === column
              ? { ...c, pk: !c.pk, notNull: !c.pk ? true : c.notNull }
              : c,
          ),
        }
      : t,
  )
}

export function addRef(schema: Schema, ref: Ref): Schema {
  // Ignore self-loops and exact duplicates.
  if (ref.fromTable === ref.toTable && ref.fromColumn === ref.toColumn) return schema
  const exists = schema.refs.some(
    (r) =>
      r.fromTable === ref.fromTable &&
      r.fromColumn === ref.fromColumn &&
      r.toTable === ref.toTable &&
      r.toColumn === ref.toColumn,
  )
  if (exists) return schema
  return { ...schema, refs: [...schema.refs, ref] }
}
