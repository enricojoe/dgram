import { createContext, useContext } from 'react'

/**
 * EditorMode lets the diagram components render read-only (e.g. the public share
 * view) without threading a prop through every React Flow node. DiagramCanvas
 * provides it; TableNode consumes it to hide editing controls.
 */
interface EditorMode {
  readOnly: boolean
}

const EditorModeContext = createContext<EditorMode>({ readOnly: false })

export const EditorModeProvider = EditorModeContext.Provider

export function useEditorMode() {
  return useContext(EditorModeContext)
}
