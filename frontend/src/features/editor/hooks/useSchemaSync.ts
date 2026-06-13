import { useEffect } from 'react'
import { AxiosError } from 'axios'
import { parseDDL } from '@/api/schema'
import { emptySchema } from '@/types/schema'
import { useEditorStore } from '../store/editorStore'

const DEBOUNCE_MS = 500

/**
 * useSchemaSync watches the DDL + dialect in the store, and after a short debounce
 * sends them to the backend parser, writing the result back into the store. A
 * per-run `cancelled` flag drops stale responses so a slow earlier request can't
 * overwrite a newer one.
 */
export function useSchemaSync() {
  const ddl = useEditorStore((s) => s.ddl)
  const dialect = useEditorStore((s) => s.dialect)
  const setParseResult = useEditorStore((s) => s.setParseResult)
  const setError = useEditorStore((s) => s.setError)
  const setParsing = useEditorStore((s) => s.setParsing)

  useEffect(() => {
    // Only parse when the DDL pane drove the change. DDL produced by /generate
    // (origin 'diagram') must not be re-parsed, or it would clobber the edit.
    if (useEditorStore.getState().origin !== 'ddl') return

    if (!ddl.trim()) {
      setParseResult(emptySchema, [])
      return
    }

    let cancelled = false
    setParsing(true)

    const handle = setTimeout(async () => {
      try {
        const res = await parseDDL(dialect, ddl)
        if (cancelled) return
        setParseResult(res.schema, res.warnings)
      } catch (err) {
        if (cancelled) return
        setError(extractParseError(err))
      } finally {
        if (!cancelled) setParsing(false)
      }
    }, DEBOUNCE_MS)

    return () => {
      cancelled = true
      clearTimeout(handle)
    }
  }, [ddl, dialect, setParseResult, setError, setParsing])
}

/** Pulls the backend's error message out of an axios error, with a fallback. */
function extractParseError(err: unknown): string {
  if (err instanceof AxiosError) {
    return err.response?.data?.error ?? err.message
  }
  return 'Failed to parse DDL'
}
