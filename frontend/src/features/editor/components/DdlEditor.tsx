import CodeMirror from '@uiw/react-codemirror'
import { sql, PostgreSQL, MySQL } from '@codemirror/lang-sql'
import { useEditorStore } from '../store/editorStore'
import type { Dialect } from '@/types/schema'

/**
 * DdlEditor is the left pane: a CodeMirror SQL editor bound to the store's DDL,
 * plus a dialect selector. Typing updates the store, which triggers re-parsing
 * via useSchemaSync.
 */
export default function DdlEditor() {
  const ddl = useEditorStore((s) => s.ddl)
  const dialect = useEditorStore((s) => s.dialect)
  const parsing = useEditorStore((s) => s.parsing)
  const generating = useEditorStore((s) => s.generating)
  const error = useEditorStore((s) => s.error)
  const warnings = useEditorStore((s) => s.warnings)
  const setDdl = useEditorStore((s) => s.setDdl)
  const setDialect = useEditorStore((s) => s.setDialect)

  return (
    <div className="flex h-full flex-col border-r border-gray-200">
      <div className="flex items-center justify-between border-b border-gray-200 px-3 py-2">
        <span className="text-sm font-semibold text-gray-700">DDL</span>
        <div className="flex items-center gap-2">
          {parsing && <span className="text-xs text-gray-400">parsing…</span>}
          {generating && <span className="text-xs text-gray-400">updating DDL…</span>}
          <select
            value={dialect}
            onChange={(e) => setDialect(e.target.value as Dialect)}
            className="rounded border border-gray-300 px-2 py-1 text-sm"
          >
            <option value="postgres">PostgreSQL</option>
            <option value="mysql">MySQL</option>
          </select>
        </div>
      </div>

      <div className="min-h-0 flex-1 overflow-auto">
        <CodeMirror
          value={ddl}
          height="100%"
          extensions={[
            sql({ dialect: dialect === 'mysql' ? MySQL : PostgreSQL }),
          ]}
          onChange={setDdl}
        />
      </div>

      {(error || warnings.length > 0) && (
        <div className="max-h-32 overflow-auto border-t border-gray-200 p-2 text-xs">
          {error && <p className="text-red-600">{error}</p>}
          {warnings.map((w, i) => (
            <p key={i} className="text-amber-600">
              ⚠ {w}
            </p>
          ))}
        </div>
      )}
    </div>
  )
}
