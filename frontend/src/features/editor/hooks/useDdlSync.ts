import { useEffect } from 'react'
import { AxiosError } from 'axios'
import { generateDDL } from '@/api/schema'
import { useEditorStore } from '../store/editorStore'

const DEBOUNCE_MS = 300

/**
 * useDdlSync is the diagram → DDL direction. When a diagram edit changes the
 * schema (origin 'diagram'), it debounces, regenerates DDL via the backend, and
 * writes it back into the editor via setGeneratedDdl (which keeps origin
 * 'diagram', so useSchemaSync does not re-parse it). A stale-guard drops
 * out-of-order responses.
 */
export function useDdlSync() {
  const schema = useEditorStore((s) => s.schema)
  const dialect = useEditorStore((s) => s.dialect)
  const setGeneratedDdl = useEditorStore((s) => s.setGeneratedDdl)
  const setError = useEditorStore((s) => s.setError)
  const setGenerating = useEditorStore((s) => s.setGenerating)

  useEffect(() => {
    // Only regenerate when the diagram drove the change.
    if (useEditorStore.getState().origin !== 'diagram') return

    let cancelled = false
    setGenerating(true)

    const handle = setTimeout(async () => {
      try {
        const ddl = await generateDDL(dialect, schema)
        if (cancelled) return
        setGeneratedDdl(ddl)
        setError(null)
      } catch (err) {
        if (cancelled) return
        setError(extractError(err))
      } finally {
        if (!cancelled) setGenerating(false)
      }
    }, DEBOUNCE_MS)

    return () => {
      cancelled = true
      clearTimeout(handle)
    }
  }, [schema, dialect, setGeneratedDdl, setError, setGenerating])
}

function extractError(err: unknown): string {
  if (err instanceof AxiosError) {
    return err.response?.data?.error ?? err.message
  }
  return 'Failed to generate DDL'
}
