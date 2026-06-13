import type { Dialect } from '@/types/schema'

// Common column types offered in the diagram's type dropdown, per dialect.
// Not exhaustive — a column's existing (possibly parameterized, e.g.
// `varchar(100)`) type is always preserved as a selectable option, and the DDL
// editor remains the path for anything more exotic.
const POSTGRES_TYPES = [
  'serial',
  'bigserial',
  'smallint',
  'int',
  'bigint',
  'numeric',
  'real',
  'double precision',
  'boolean',
  'varchar(255)',
  'text',
  'char',
  'uuid',
  'date',
  'time',
  'timestamp',
  'timestamptz',
  'json',
  'jsonb',
]

const MYSQL_TYPES = [
  'tinyint',
  'smallint',
  'int',
  'bigint',
  'decimal',
  'float',
  'double',
  'boolean',
  'varchar(255)',
  'text',
  'char',
  'date',
  'time',
  'datetime',
  'timestamp',
  'json',
]

/** Common column types for the given dialect, used by the type dropdown. */
export function columnTypes(dialect: Dialect): string[] {
  return dialect === 'mysql' ? MYSQL_TYPES : POSTGRES_TYPES
}
