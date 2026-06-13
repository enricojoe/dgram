import Dagre from '@dagrejs/dagre'
import type { Edge, Node } from '@xyflow/react'
import { MarkerType } from '@xyflow/react'
import type { Schema, Table } from '@/types/schema'

// Shared node dimensions so dagre's layout and the rendered TableNode agree.
export const NODE_WIDTH = 260
export const HEADER_HEIGHT = 36
export const ROW_HEIGHT = 26
const NODE_PADDING = 8

/** Data carried by each table node; consumed by the TableNode component. */
export interface TableNodeData extends Record<string, unknown> {
  table: Table
}

export type TableNode = Node<TableNodeData, 'table'>

/** Estimated rendered height of a table node, used for layout. */
function nodeHeight(table: Table): number {
  return HEADER_HEIGHT + table.columns.length * ROW_HEIGHT + NODE_PADDING
}

// Handle ids are derived from column names so foreign-key edges can attach to
// the exact column rows. Each column row renders both a target (left) and a
// source (right) handle.
export const targetHandle = (column: string) => `${column}-target`
export const sourceHandle = (column: string) => `${column}-source`

/**
 * schemaToFlow maps a parsed Schema into React Flow nodes + edges (without
 * positions — positions are assigned by layoutWithDagre). One node per table,
 * one edge per foreign-key column pair.
 */
export function schemaToFlow(schema: Schema): {
  nodes: TableNode[]
  edges: Edge[]
} {
  const nodes: TableNode[] = schema.tables.map((table) => ({
    id: table.name,
    type: 'table',
    position: { x: 0, y: 0 },
    data: { table },
    width: NODE_WIDTH,
    height: nodeHeight(table),
  }))

  const edges: Edge[] = schema.refs.map((ref) => ({
    id: `${ref.fromTable}.${ref.fromColumn}->${ref.toTable}.${ref.toColumn}`,
    source: ref.fromTable,
    target: ref.toTable,
    sourceHandle: sourceHandle(ref.fromColumn),
    targetHandle: targetHandle(ref.toColumn),
    type: 'smoothstep',
    label: ref.onDelete ? `ON DELETE ${ref.onDelete}` : undefined,
    markerEnd: { type: MarkerType.ArrowClosed },
  }))

  return { nodes, edges }
}

/**
 * layoutWithDagre assigns left-to-right positions to nodes based on the edge
 * graph. dagre returns center coordinates; React Flow wants top-left, so we
 * shift by half the node size.
 */
export function layoutWithDagre(nodes: TableNode[], edges: Edge[]): TableNode[] {
  const g = new Dagre.graphlib.Graph().setDefaultEdgeLabel(() => ({}))
  g.setGraph({ rankdir: 'LR', nodesep: 40, ranksep: 90 })

  nodes.forEach((n) =>
    g.setNode(n.id, { width: n.width ?? NODE_WIDTH, height: n.height ?? 0 }),
  )
  edges.forEach((e) => g.setEdge(e.source, e.target))

  Dagre.layout(g)

  return nodes.map((n) => {
    const { x, y, width, height } = g.node(n.id)
    return { ...n, position: { x: x - width / 2, y: y - height / 2 } }
  })
}

/**
 * layoutWithStored applies dagre as a baseline, then overrides positions for any
 * table that already has a saved (dragged) position. New/unsaved tables fall
 * back to the dagre position, so they appear in a sensible spot.
 */
export function layoutWithStored(
  nodes: TableNode[],
  edges: Edge[],
  stored: Record<string, { x: number; y: number }>,
): TableNode[] {
  return layoutWithDagre(nodes, edges).map((n) =>
    stored[n.id] ? { ...n, position: stored[n.id] } : n,
  )
}
