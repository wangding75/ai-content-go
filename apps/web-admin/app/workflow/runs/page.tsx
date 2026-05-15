'use client'

import { useEffect, useState } from 'react'
import {
  fetchWorkflowRuns,
  createWorkflowRun,
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

  const load = async () => {
    setLoading(true)
    try {
      const res = await fetchWorkflowRuns({ project_id: projectId, template_version_id: templateVersionId })
      setRuns(res.items ?? [])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [projectId, templateVersionId])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    await createWorkflowRun({ project_id: newProjectId, template_version_id: newTemplateVersionId })
    setCreating(false)
    setNewProjectId('')
    setNewTemplateVersionId('')
    load()
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-4">
        <h1 className="text-2xl font-bold">Workflow Runs</h1>
        <button onClick={() => setCreating(true)} className="btn-primary">Trigger Run</button>
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
          <button type="submit" className="btn-primary mr-2">Create Run</button>
          <button type="button" onClick={() => setCreating(false)}>Cancel</button>
        </form>
      )}

      {loading ? (
        <p>Loading...</p>
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
