import { Link, useNavigate } from 'react-router-dom'
import { useDiagrams, useDeleteDiagram } from '@/api/diagrams'

/**
 * DashboardPage lists the signed-in user's saved diagrams as a responsive card
 * grid with open/delete actions, plus a button to start a new one.
 */
export default function DashboardPage() {
  const navigate = useNavigate()
  const { data: diagrams, isLoading, error } = useDiagrams()
  const del = useDeleteDiagram()

  const formatDate = (iso: string) =>
    new Date(iso).toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })

  return (
    <div className="mx-auto max-w-5xl p-6">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold text-gray-900 dark:text-gray-100">
            My Diagrams
          </h1>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            Your saved schema diagrams.
          </p>
        </div>
        <button
          type="button"
          onClick={() => navigate('/')}
          className="rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700"
        >
          + New diagram
        </button>
      </div>

      {isLoading && (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div
              key={i}
              className="h-28 animate-pulse rounded-lg border border-gray-200 bg-gray-50 dark:border-gray-800 dark:bg-gray-900"
            />
          ))}
        </div>
      )}

      {error && (
        <p className="text-sm text-red-600 dark:text-red-400">
          Failed to load diagrams.
        </p>
      )}

      {diagrams && diagrams.length === 0 && (
        <div className="rounded-lg border border-dashed border-gray-300 p-12 text-center dark:border-gray-700">
          <p className="text-sm text-gray-500 dark:text-gray-400">
            No diagrams yet.
          </p>
          <button
            type="button"
            onClick={() => navigate('/')}
            className="mt-3 rounded-md bg-indigo-600 px-4 py-2 text-sm font-medium text-white hover:bg-indigo-700"
          >
            Create your first diagram
          </button>
        </div>
      )}

      {diagrams && diagrams.length > 0 && (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {diagrams.map((d) => (
            <div
              key={d.id}
              className="group flex flex-col justify-between rounded-lg border border-gray-200 bg-white p-4 shadow-sm transition hover:shadow-md dark:border-gray-800 dark:bg-gray-900"
            >
              <Link to={`/d/${d.id}`} className="min-w-0">
                <h2 className="truncate font-medium text-gray-900 group-hover:text-indigo-700 dark:text-gray-100 dark:group-hover:text-indigo-400">
                  {d.name}
                </h2>
                <span className="mt-1 inline-block rounded bg-gray-100 px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide text-gray-500 dark:bg-gray-800 dark:text-gray-400">
                  {d.dialect}
                </span>
              </Link>
              <div className="mt-4 flex items-center justify-between">
                <span className="text-xs text-gray-400 dark:text-gray-500">
                  Updated {formatDate(d.updatedAt)}
                </span>
                <div className="flex items-center gap-3 text-sm">
                  <Link
                    to={`/d/${d.id}`}
                    className="text-indigo-600 hover:underline dark:text-indigo-400"
                  >
                    Open
                  </Link>
                  <button
                    type="button"
                    onClick={() => {
                      if (confirm(`Delete "${d.name}"?`)) del.mutate(d.id)
                    }}
                    className="text-gray-400 hover:text-red-600 dark:hover:text-red-400"
                  >
                    Delete
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
