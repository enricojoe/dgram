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
  const commitLayout = useEditorStore((s) => s.commitLayout)

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
      // Snapshot current positions so the re-layout triggered by the new edge
      // keeps every table where it is instead of letting dagre reflow them.
      commitLayout(Object.fromEntries(nodes.map((n) => [n.id, n.position])))
      applySchemaEdit((s) =>
        addRef(s, {
          fromTable: c.source!,
          fromColumn: columnFromHandle(c.sourceHandle, '-source'),
          toTable: c.target!,
          toColumn: columnFromHandle(c.targetHandle, '-target'),
        }),
      )
    },
    [applySchemaEdit, commitLayout, nodes],
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

// Properties whose computed value can be an oklch() color.
const COLOR_PROPS = [
  'color',
  'backgroundColor',
  'borderTopColor',
  'borderRightColor',
  'borderBottomColor',
  'borderLeftColor',
  'outlineColor',
  'fill',
  'stroke',
] as const

// Reuse one canvas: assigning any CSS color to fillStyle normalizes it to
// rgb()/#hex, which is how we convert oklch() to something html-to-image reads.
const colorCanvas = document.createElement('canvas').getContext('2d')
const toRgb = (value: string): string => {
  if (!colorCanvas) return value
  colorCanvas.fillStyle = '#000'
  colorCanvas.fillStyle = value
  return colorCanvas.fillStyle
}

/**
 * Tailwind v4's default palette uses the oklch() color space, which
 * html-to-image cannot parse and renders as black. Walk the subtree, convert
 * any computed oklch colors to rgb, and pin them as inline styles so they win
 * during capture. Returns a function that restores the original inline styles.
 */
function neutralizeOklch(root: HTMLElement): () => void {
  const restores: Array<() => void> = []
  for (const el of [root, ...Array.from(root.querySelectorAll<HTMLElement>('*'))]) {
    const computed = getComputedStyle(el)
    for (const prop of COLOR_PROPS) {
      const value = computed[prop as keyof CSSStyleDeclaration] as string
      if (typeof value === 'string' && value.includes('oklch')) {
        const previous = el.style.getPropertyValue(prop)
        const priority = el.style.getPropertyPriority(prop)
        el.style.setProperty(prop, toRgb(value))
        restores.push(() => el.style.setProperty(prop, previous, priority))
      }
    }
  }
  return () => restores.forEach((restore) => restore())
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

    const restore = neutralizeOklch(el)
    toPng(el, {
      backgroundColor: '#ffffff',
      width,
      height,
      style: {
        width: `${width}px`,
        height: `${height}px`,
        transform: `translate(${vp.x}px, ${vp.y}px) scale(${vp.zoom})`,
      },
    })
      .then((dataUrl) => {
        const a = document.createElement('a')
        a.download = 'diagram.png'
        a.href = dataUrl
        a.click()
      })
      .finally(restore)
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
