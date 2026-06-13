import { Link, useNavigate } from 'react-router-dom'
import { useDiagrams, useDeleteDiagram } from '@/api/diagrams'

/**
 * DashboardPage lists the signed-in user's saved diagrams with open/delete
 * actions, plus a button to start a new one.
 */
export default function DashboardPage() {
  const navigate = useNavigate()
  const { data: diagrams, isLoading, error } = useDiagrams()
  const del = useDeleteDiagram()

  return (
    <div className="mx-auto max-w-3xl p-6">
      <div className="mb-4 flex items-center justify-between">
        <h1 className="text-xl font-semibold text-gray-800">My Diagrams</h1>
        <button
          type="button"
          onClick={() => navigate('/')}
          className="rounded bg-indigo-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-indigo-700"
        >
          + New diagram
        </button>
      </div>

      {isLoading && <p className="text-sm text-gray-500">Loading…</p>}
      {error && <p className="text-sm text-red-600">Failed to load diagrams.</p>}

      {diagrams && diagrams.length === 0 && (
        <p className="text-sm text-gray-500">
          No diagrams yet. Create one and hit Save.
        </p>
      )}

      <ul className="divide-y divide-gray-100 rounded border border-gray-200">
        {diagrams?.map((d) => (
          <li key={d.id} className="flex items-center justify-between px-4 py-3">
            <Link to={`/d/${d.id}`} className="min-w-0">
              <span className="font-medium text-gray-800 hover:text-indigo-700">
                {d.name}
              </span>
              <span className="ml-2 text-xs uppercase text-gray-400">
                {d.dialect}
              </span>
            </Link>
            <button
              type="button"
              onClick={() => {
                if (confirm(`Delete "${d.name}"?`)) del.mutate(d.id)
              }}
              className="text-sm text-gray-400 hover:text-red-600"
            >
              Delete
            </button>
          </li>
        ))}
      </ul>
    </div>
  )
}
