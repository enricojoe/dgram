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
    <div className="flex h-full flex-col">
      <div className="flex items-center gap-2 border-b border-gray-200 px-3 py-2">
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          className="rounded border border-gray-300 px-2 py-1 text-sm"
          aria-label="Diagram name"
        />
        <button
          type="button"
          onClick={save}
          disabled={saving}
          className="rounded bg-indigo-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-indigo-700 disabled:opacity-50"
        >
          {saving ? 'Saving…' : token ? 'Save' : 'Log in to save'}
        </button>

        <button
          type="button"
          onClick={exportSql}
          className="rounded border border-gray-300 px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-50"
        >
          Export SQL
        </button>

        {routeId && (
          <button
            type="button"
            onClick={() => share.mutate()}
            disabled={share.isPending}
            className="rounded border border-gray-300 px-3 py-1.5 text-sm text-gray-700 hover:bg-gray-50 disabled:opacity-50"
          >
            {share.isPending ? 'Sharing…' : 'Share'}
          </button>
        )}

        {shareUrl && (
          <span className="truncate text-xs text-green-700" title={shareUrl}>
            Link copied: {shareUrl}
          </span>
        )}

        {(create.isError || update.isError) && (
          <span className="text-xs text-red-600">Save failed</span>
        )}
      </div>

      <div className="grid min-h-0 flex-1 grid-cols-[minmax(340px,2fr)_5fr]">
        <DdlEditor />
        <DiagramCanvas />
      </div>
    </div>
  )
}
