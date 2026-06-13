import { useCallback, useEffect, useMemo } from 'react'
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  Panel,
  useNodesState,
  useEdgesState,
  useReactFlow,
  getNodesBounds,
  getViewportForBounds,
  type Connection,
  type Node,
  type NodeTypes,
} from '@xyflow/react'
import { toPng } from 'html-to-image'
import { useEditorStore } from '../store/editorStore'
import { EditorModeProvider } from '../editorMode'
import { schemaToFlow, layoutWithStored } from '@/lib/schemaFlow'
import { addRef, addTable } from '@/lib/schemaEdits'
import TableNode from './TableNode'

const nodeTypes: NodeTypes = { table: TableNode }

const columnFromHandle = (handle: string | null | undefined, suffix: string) =>
  handle?.endsWith(suffix) ? handle.slice(0, -suffix.length) : handle ?? ''

/**
 * DiagramCanvas is the right pane. It derives nodes/edges from the store schema,
 * preserves dragged positions, and (unless readOnly) supports editing: drag to
 * reposition, connect column handles to create a foreign key, and Add-table.
 * readOnly is used by the public share view.
 */
export default function DiagramCanvas({ readOnly = false }: { readOnly?: boolean }) {
  const schema = useEditorStore((s) => s.schema)
  const applySchemaEdit = useEditorStore((s) => s.applySchemaEdit)
  const setNodePosition = useEditorStore((s) => s.setNodePosition)

  const laidOut = useMemo(() => {
    const { nodes, edges } = schemaToFlow(schema)
    const stored = useEditorStore.getState().layout
    return { nodes: layoutWithStored(nodes, edges, stored), edges }
  }, [schema])

  const [nodes, setNodes, onNodesChange] = useNodesState(laidOut.nodes)
  const [edges, setEdges, onEdgesChange] = useEdgesState(laidOut.edges)

  useEffect(() => {
    setNodes(laidOut.nodes)
    setEdges(laidOut.edges)
  }, [laidOut, setNodes, setEdges])

  const onNodeDragStop = useCallback(
    (_: unknown, node: Node) => setNodePosition(node.id, node.position),
    [setNodePosition],
  )

  const onConnect = useCallback(
    (c: Connection) => {
      if (!c.source || !c.target) return
      applySchemaEdit((s) =>
        addRef(s, {
          fromTable: c.source!,
          fromColumn: columnFromHandle(c.sourceHandle, '-source'),
          toTable: c.target!,
          toColumn: columnFromHandle(c.targetHandle, '-target'),
        }),
      )
    },
    [applySchemaEdit],
  )

  return (
    <EditorModeProvider value={{ readOnly }}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeDragStop={readOnly ? undefined : onNodeDragStop}
        onConnect={readOnly ? undefined : onConnect}
        nodesConnectable={!readOnly}
        nodesDraggable={!readOnly}
        nodeTypes={nodeTypes}
        fitView
        proOptions={{ hideAttribution: true }}
      >
        <Panel position="top-right" className="flex gap-2">
          {!readOnly && (
            <button
              type="button"
              onClick={() => applySchemaEdit(addTable)}
              className="rounded bg-indigo-600 px-3 py-1.5 text-sm font-medium text-white shadow hover:bg-indigo-700"
            >
              + Add table
            </button>
          )}
          <ExportPngButton />
        </Panel>
        <Background />
        <Controls />
        <MiniMap pannable zoomable />
      </ReactFlow>
    </EditorModeProvider>
  )
}

/** Exports the current diagram as a PNG. Must render inside <ReactFlow>. */
function ExportPngButton() {
  const { getNodes } = useReactFlow()

  const onExport = useCallback(() => {
    const bounds = getNodesBounds(getNodes())
    if (!bounds.width || !bounds.height) return
    const width = bounds.width + 80
    const height = bounds.height + 80
    const vp = getViewportForBounds(bounds, width, height, 0.5, 2, 40)
    const el = document.querySelector('.react-flow__viewport') as HTMLElement | null
    if (!el) return

    toPng(el, {
      backgroundColor: '#ffffff',
      width,
      height,
      style: {
        width: `${width}px`,
        height: `${height}px`,
        transform: `translate(${vp.x}px, ${vp.y}px) scale(${vp.zoom})`,
      },
    }).then((dataUrl) => {
      const a = document.createElement('a')
      a.download = 'diagram.png'
      a.href = dataUrl
      a.click()
    })
  }, [getNodes])

  return (
    <button
      type="button"
      onClick={onExport}
      className="rounded border border-gray-300 bg-white px-3 py-1.5 text-sm font-medium text-gray-700 shadow hover:bg-gray-50"
    >
      Export PNG
    </button>
  )
}
