'use client'

import { useEffect, useState } from 'react'
import {
  fetchWorkflowRuns,
  createWorkflowRun,
  pageErrorFromEnvelope,
  PageError,
  WorkflowRunResponse,
} from '@/lib/api'

export default function WorkflowRunsPage() {
  const [runs, setRuns] = useState<WorkflowRunResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [projectId, setProjectId] = useState('')
  const [templateVersionId, setTemplateVersionId] = useState('')
  const [creating, setCreating] = useState(false)
  const [newProjectId, setNewProjectId] = useState('')
  const [newTemplateVersionId, setNewTemplateVersionId] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [success, setSuccess] = useState('')
  const [error, setError] = useState<PageError | null>(null)

  const load = async () => {
    setLoading(true)
    try {
      const res = await fetchWorkflowRuns({ project_id: projectId, template_version_id: templateVersionId })
      if (!res.success || !res.data) {
        setError(pageErrorFromEnvelope(res, 'Failed to load workflow runs'))
        setRuns([])
        return
      }
      setError(null)
      setRuns(res.data.items)
    } catch {
      setError({ message: 'Failed to load workflow runs' })
      setRuns([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [projectId, templateVersionId])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setSuccess('')
    try {
      const res = await createWorkflowRun({ project_id: newProjectId, template_version_id: newTemplateVersionId })
      if (!res.success) {
        setError(pageErrorFromEnvelope(res, 'Failed to create workflow run'))
        return
      }
      setError(null)
      setSuccess(`Created workflow run ${res.data?.workflow_run_id ?? ''}`.trim())
    } catch {
      setError({ message: 'Failed to create workflow run' })
      return
    } finally {
      setSubmitting(false)
    }
    setCreating(false)
    setNewProjectId('')
    setNewTemplateVersionId('')
    load()
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-4">
        <h1 className="text-2xl font-bold">Workflow Runs</h1>
        <button onClick={() => setCreating(true)} disabled={submitting} className="btn-primary">Trigger Run</button>
      </div>

      <div className="flex gap-4 mb-4">
        <input
          placeholder="project_id"
          value={projectId}
          onChange={e => setProjectId(e.target.value)}
          className="border px-2 py-1 rounded"
        />
        <input
          placeholder="template_version_id"
          value={templateVersionId}
          onChange={e => setTemplateVersionId(e.target.value)}
          className="border px-2 py-1 rounded"
        />
      </div>

      {creating && (
        <form onSubmit={handleCreate} className="mb-4 p-4 border rounded">
          <input
            placeholder="project_id"
            value={newProjectId}
            onChange={e => setNewProjectId(e.target.value)}
            className="border px-2 py-1 rounded mr-2"
          />
          <input
            placeholder="template_version_id"
            value={newTemplateVersionId}
            onChange={e => setNewTemplateVersionId(e.target.value)}
            required
            className="border px-2 py-1 rounded mr-2"
          />
          <button type="submit" disabled={submitting} className="btn-primary mr-2">
            {submitting ? 'Creating...' : 'Create Run'}
          </button>
          <button type="button" disabled={submitting} onClick={() => setCreating(false)}>Cancel</button>
        </form>
      )}

      {success && <p role="status" className="mb-4 text-green-600">{success}</p>}
      {error && (
        <div role="alert" className="mb-4 text-red-600">
          <p>{error.message}</p>
          {error.code && <p>code: {error.code}</p>}
          {error.request_id && <p>request_id: {error.request_id}</p>}
        </div>
      )}

      {loading ? (
        <p data-testid="workflow-runs-loading">Loading...</p>
      ) : runs.length === 0 ? (
        <p data-testid="workflow-runs-empty">No workflow runs found.</p>
      ) : (
        <table className="w-full border-collapse">
          <thead>
            <tr>
              <th className="border px-4 py-2 text-left">ID</th>
              <th className="border px-4 py-2 text-left">project_id</th>
              <th className="border px-4 py-2 text-left">template_version_id</th>
              <th className="border px-4 py-2 text-left">Status</th>
            </tr>
          </thead>
          <tbody>
            {runs.map((r: WorkflowRunResponse) => (
              <tr key={r.id}>
                <td className="border px-4 py-2">
                  <a href={`/workflow/runs/${r.id}`} className="text-blue-600 hover:underline">{r.id}</a>
                </td>
                <td className="border px-4 py-2">{r.project_id}</td>
                <td className="border px-4 py-2">{r.template_version_id}</td>
                <td className="border px-4 py-2">{r.status}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
