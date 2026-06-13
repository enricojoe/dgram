import { useState } from 'react'
import { Handle, Position, type NodeProps } from '@xyflow/react'
import {
  ROW_HEIGHT,
  HEADER_HEIGHT,
  NODE_WIDTH,
  sourceHandle,
  targetHandle,
  type TableNode as TableNodeType,
} from '@/lib/schemaFlow'
import { useEditorStore } from '../store/editorStore'
import { useEditorMode } from '../editorMode'
import { columnTypes } from '@/lib/columnTypes'
import {
  addColumn,
  deleteColumn,
  deleteTable,
  renameTable,
  togglePk,
  updateColumn,
} from '@/lib/schemaEdits'

/**
 * TableNode renders one editable database table. Double-click the title or a
 * column name/type to edit; click the key to toggle PK; hover a row for delete.
 * Each column row exposes left (target) and right (source) handles keyed by
 * column name, so foreign-key edges connect specific columns. All edits go
 * through applySchemaEdit, which regenerates the DDL.
 */
export default function TableNode({ data }: NodeProps<TableNodeType>) {
  const { table } = data
  const applySchemaEdit = useEditorStore((s) => s.applySchemaEdit)
  const dialect = useEditorStore((s) => s.dialect)
  const { readOnly } = useEditorMode()

  return (
    <div
      className="overflow-hidden rounded-md border border-gray-300 bg-white shadow-sm"
      style={{ width: NODE_WIDTH }}
    >
      <div
        className="flex items-center justify-between bg-indigo-600 px-3 font-semibold text-white"
        style={{ height: HEADER_HEIGHT }}
      >
        <EditableText
          value={table.name}
          disabled={readOnly}
          onCommit={(name) => applySchemaEdit((s) => renameTable(s, table.name, name))}
          className="bg-transparent text-white placeholder-indigo-200"
        />
        {!readOnly && (
          <button
            type="button"
            title="Delete table"
            className="nodrag ml-2 text-indigo-200 hover:text-white"
            onClick={() => applySchemaEdit((s) => deleteTable(s, table.name))}
          >
            ✕
          </button>
        )}
      </div>

      {table.columns.map((col) => (
        <div
          key={col.name}
          className="group relative flex items-center gap-1 border-t border-gray-100 px-3 text-sm"
          style={{ height: ROW_HEIGHT }}
        >
          <Handle
            type="target"
            position={Position.Left}
            id={targetHandle(col.name)}
            className="!h-2 !w-2 !border-gray-400 !bg-white"
          />

          <button
            type="button"
            title="Toggle primary key"
            disabled={readOnly}
            className="nodrag w-4 shrink-0 text-left"
            onClick={() =>
              !readOnly && applySchemaEdit((s) => togglePk(s, table.name, col.name))
            }
          >
            {col.pk ? '🔑' : <span className="text-gray-300">·</span>}
          </button>

          <EditableText
            value={col.name}
            disabled={readOnly}
            onCommit={(name) =>
              applySchemaEdit((s) => updateColumn(s, table.name, col.name, { name }))
            }
            className={`min-w-0 flex-1 ${col.pk ? 'font-medium' : ''}`}
          />

          <TypeSelect
            value={col.type}
            options={columnTypes(dialect)}
            disabled={readOnly}
            onChange={(type) =>
              applySchemaEdit((s) => updateColumn(s, table.name, col.name, { type }))
            }
          />

          {!readOnly && (
            <span className="flex w-4 shrink-0 justify-center">
              <button
                type="button"
                title="Delete column"
                className="nodrag invisible text-gray-400 hover:text-red-600 group-hover:visible"
                onClick={() => applySchemaEdit((s) => deleteColumn(s, table.name, col.name))}
              >
                ✕
              </button>
            </span>
          )}

          <Handle
            type="source"
            position={Position.Right}
            id={sourceHandle(col.name)}
            className="!h-2 !w-2 !border-gray-400 !bg-white"
          />
        </div>
      ))}

      {!readOnly && (
        <button
          type="button"
          className="nodrag block w-full border-t border-gray-100 py-1 text-xs text-indigo-600 hover:bg-indigo-50"
          onClick={() => applySchemaEdit((s) => addColumn(s, table.name))}
        >
          + column
        </button>
      )}
    </div>
  )
}

/**
 * TypeSelect is the column type picker. In edit mode it's a fixed-width dropdown
 * of the dialect's common types; the column's current type is always included so
 * parameterized/custom types (e.g. `varchar(100)`) are never lost. In read-only
 * (share) view it renders as plain text.
 */
function TypeSelect({
  value,
  options,
  onChange,
  disabled = false,
}: {
  value: string
  options: string[]
  onChange: (next: string) => void
  disabled?: boolean
}) {
  if (disabled) {
    return (
      <span className="w-20 shrink-0 truncate text-right text-xs text-gray-500">
        {value}
      </span>
    )
  }

  const opts = options.includes(value) ? options : [value, ...options]

  return (
    <select
      value={value}
      onChange={(e) => e.target.value !== value && onChange(e.target.value)}
      title={value}
      className="nodrag w-20 shrink-0 cursor-pointer truncate rounded bg-transparent text-right text-xs text-gray-500 outline-none hover:text-gray-700"
    >
      {opts.map((t) => (
        <option key={t} value={t}>
          {t}
        </option>
      ))}
    </select>
  )
}

/**
 * EditableText shows text that becomes an input on double-click; commits on
 * Enter/blur, cancels on Escape. `nodrag` keeps clicks from starting a node drag.
 */
function EditableText({
  value,
  onCommit,
  className = '',
  disabled = false,
}: {
  value: string
  onCommit: (next: string) => void
  className?: string
  disabled?: boolean
}) {
  const [editing, setEditing] = useState(false)
  const [draft, setDraft] = useState(value)

  if (disabled || !editing) {
    return (
      <span
        className={`block truncate ${disabled ? '' : 'nodrag cursor-text'} ${className}`}
        onDoubleClick={
          disabled
            ? undefined
            : () => {
                setDraft(value)
                setEditing(true)
              }
        }
      >
        {value}
      </span>
    )
  }

  const commit = () => {
    setEditing(false)
    if (draft.trim() && draft !== value) onCommit(draft.trim())
  }

  return (
    <input
      autoFocus
      value={draft}
      onChange={(e) => setDraft(e.target.value)}
      onBlur={commit}
      onKeyDown={(e) => {
        if (e.key === 'Enter') commit()
        if (e.key === 'Escape') setEditing(false)
      }}
      className={`nodrag w-full rounded border border-indigo-300 px-1 outline-none ${className}`}
    />
  )
}
