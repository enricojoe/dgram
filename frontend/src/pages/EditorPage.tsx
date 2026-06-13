import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import DdlEditor from '@/features/editor/components/DdlEditor'
import DiagramCanvas from '@/features/editor/components/DiagramCanvas'
import { useSchemaSync } from '@/features/editor/hooks/useSchemaSync'
import { useDdlSync } from '@/features/editor/hooks/useDdlSync'
import { useEditorStore } from '@/features/editor/store/editorStore'
import { useAuthStore } from '@/features/auth/store/authStore'
import {
  useDiagram,
  useCreateDiagram,
  useUpdateDiagram,
} from '@/api/diagrams'
import { enableShare } from '@/api/share'
import { BurgerIcon } from '@/components/icons'

/**
 * EditorPage — the core workspace. Anonymous users can parse/edit freely;
 * saving requires login. When opened as /d/:id it loads that saved diagram.
 * The two sync hooks keep DDL and diagram in step in both directions.
 */
export default function EditorPage() {
  useSchemaSync()
  useDdlSync()

  const params = useParams()
  const navigate = useNavigate()
  const routeId = params.id ? Number(params.id) : undefined

  const token = useAuthStore((s) => s.accessToken)
  const loadDiagram = useEditorStore((s) => s.loadDiagram)

  const [name, setName] = useState('Untitled diagram')
  const [ddlCollapsed, setDdlCollapsed] = useState(false)
  const { data: loaded } = useDiagram(routeId)
  const create = useCreateDiagram()
  const update = useUpdateDiagram()

  // When a saved diagram is fetched, load it into the editor store.
  useEffect(() => {
    if (loaded) {
      setName(loaded.name)
      loadDiagram({
        ddl: loaded.ddl,
        dialect: loaded.dialect,
        layout: loaded.layout ?? {},
      })
    }
  }, [loaded, loadDiagram])

  const [shareUrl, setShareUrl] = useState<string | null>(null)
  const share = useMutation({
    mutationFn: () => enableShare(routeId as number),
    onSuccess: ({ shareId }) => {
      const url = `${window.location.origin}/s/${shareId}`
      setShareUrl(url)
      navigator.clipboard?.writeText(url).catch(() => {})
    },
  })

  const exportSql = () => {
    const { ddl } = useEditorStore.getState()
    const blob = new Blob([ddl], { type: 'application/sql' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${(name.trim() || 'schema').replace(/\s+/g, '_')}.sql`
    a.click()
    URL.revokeObjectURL(url)
  }

  const save = () => {
    if (!token) {
      navigate('/login')
      return
    }
    // Read the freshest values straight from the store.
    const { ddl, dialect, layout } = useEditorStore.getState()
    const input = { name: name.trim() || 'Untitled diagram', dialect, ddl, layout }

    if (routeId) {
      update.mutate({ id: routeId, input })
    } else {
      create.mutate(input, { onSuccess: (d) => navigate(`/d/${d.id}`) })
    }
  }

  const saving = create.isPending || update.isPending

  return (
    <div className="flex h-full flex-col bg-white text-gray-900 dark:bg-gray-950 dark:text-gray-100">
      <div className="flex items-center gap-3 border-b border-gray-200 px-3 py-2 dark:border-gray-800">
        <button
          type="button"
          onClick={() => setDdlCollapsed((v) => !v)}
          title={ddlCollapsed ? 'Show DDL editor' : 'Hide DDL editor'}
          aria-label={ddlCollapsed ? 'Show DDL editor' : 'Hide DDL editor'}
          aria-pressed={!ddlCollapsed}
          className="shrink-0 rounded p-1 text-gray-600 hover:bg-gray-100 dark:text-gray-300 dark:hover:bg-gray-800"
        >
          <BurgerIcon />
        </button>

        {/* Title sits on the left and reads like a heading until focused. */}
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Untitled diagram"
          className="min-w-0 flex-1 rounded border border-transparent bg-transparent px-2 py-1 text-sm font-semibold text-gray-900 hover:border-gray-300 focus:border-indigo-500 focus:bg-white focus:font-normal focus:outline-none dark:text-gray-100 dark:hover:border-gray-700 dark:focus:bg-gray-900"
          aria-label="Diagram name"
        />

        {/* Inline status, kept next to the actions it relates to. */}
        {shareUrl && (
          <span className="hidden truncate text-xs text-green-700 sm:inline dark:text-green-400" title={shareUrl}>
            Link copied
          </span>
        )}
        {(create.isError || update.isError) && (
          <span className="text-xs text-red-600 dark:text-red-400">Save failed</span>
        )}

        {/* Actions are right-aligned, secondary first and the primary Save last. */}
        <div className="flex shrink-0 items-center gap-2">
          <button
            type="button"
            onClick={exportSql}
            className="rounded border border-gray-300 px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-200 dark:hover:bg-gray-800"
          >
            Export SQL
          </button>

          {routeId && (
            <button
              type="button"
              onClick={() => share.mutate()}
              disabled={share.isPending}
              className="rounded border border-gray-300 px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-50 disabled:opacity-50 dark:border-gray-700 dark:text-gray-200 dark:hover:bg-gray-800"
            >
              {share.isPending ? 'Sharing…' : 'Share'}
            </button>
          )}

          <button
            type="button"
            onClick={save}
            disabled={saving}
            className="rounded bg-indigo-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
          >
            {saving ? 'Saving…' : token ? 'Save' : 'Log in to save'}
          </button>
        </div>
      </div>

      <div
        className={`grid min-h-0 flex-1 ${
          ddlCollapsed ? 'grid-cols-1' : 'grid-cols-[minmax(340px,2fr)_5fr]'
        }`}
      >
        {!ddlCollapsed && <DdlEditor />}
        <DiagramCanvas />
      </div>
    </div>
  )
}
