'use client'

import { useEffect, useState } from 'react'
import {
  fetchWorkflowTemplates,
  createWorkflowTemplate,
  pageErrorFromEnvelope,
  PageError,
  WorkflowTemplateResponse,
} from '@/lib/api'

export default function WorkflowTemplatesPage() {
  const [templates, setTemplates] = useState<WorkflowTemplateResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [contentType, setContentType] = useState('')
  const [category, setCategory] = useState('')
  const [status, setStatus] = useState('')
  const [creating, setCreating] = useState(false)
  const [newCode, setNewCode] = useState('')
  const [newName, setNewName] = useState('')
  const [newContentType, setNewContentType] = useState('')
  const [newCategory, setNewCategory] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [success, setSuccess] = useState('')
  const [error, setError] = useState<PageError | null>(null)

  const load = async () => {
    setLoading(true)
    try {
      const res = await fetchWorkflowTemplates({ content_type: contentType, category, status })
      if (!res.success || !res.data) {
        setError(pageErrorFromEnvelope(res, 'Failed to load workflow templates'))
        setTemplates([])
        return
      }
      setError(null)
      setTemplates(res.data.items)
    } catch {
      setError({ message: 'Failed to load workflow templates' })
      setTemplates([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [contentType, category, status])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setSuccess('')
    try {
      const res = await createWorkflowTemplate({ code: newCode, name: newName, content_type: newContentType, category: newCategory })
      if (!res.success) {
        setError(pageErrorFromEnvelope(res, 'Failed to create workflow template'))
        return
      }
      setError(null)
      setSuccess(`Created workflow template ${res.data?.workflow_template_id ?? newCode}`)
    } catch {
      setError({ message: 'Failed to create workflow template' })
      return
    } finally {
      setSubmitting(false)
    }
    setCreating(false)
    setNewCode('')
    setNewName('')
    setNewContentType('')
    setNewCategory('')
    load()
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-4">
        <h1 className="text-2xl font-bold">Workflow Templates</h1>
        <button onClick={() => setCreating(true)} disabled={submitting} className="btn-primary">New Template</button>
      </div>

      <div className="flex gap-4 mb-4">
        <input
          placeholder="content_type"
          value={contentType}
          onChange={e => setContentType(e.target.value)}
          className="border px-2 py-1 rounded"
        />
        <input
          placeholder="category"
          value={category}
          onChange={e => setCategory(e.target.value)}
          className="border px-2 py-1 rounded"
        />
        <select value={status} onChange={e => setStatus(e.target.value)} className="border px-2 py-1 rounded">
          <option value="">All status</option>
          <option value="draft">draft</option>
          <option value="active">active</option>
          <option value="archived">archived</option>
        </select>
      </div>

      {creating && (
        <form onSubmit={handleCreate} className="mb-4 p-4 border rounded">
          <input
            placeholder="code"
            value={newCode}
            onChange={e => setNewCode(e.target.value)}
            required
            className="border px-2 py-1 rounded mr-2"
          />
          <input
            placeholder="Name"
            value={newName}
            onChange={e => setNewName(e.target.value)}
            required
            className="border px-2 py-1 rounded mr-2"
          />
          <input
            placeholder="content_type"
            value={newContentType}
            onChange={e => setNewContentType(e.target.value)}
            className="border px-2 py-1 rounded mr-2"
          />
          <input
            placeholder="category"
            value={newCategory}
            onChange={e => setNewCategory(e.target.value)}
            className="border px-2 py-1 rounded mr-2"
          />
          <button type="submit" disabled={submitting} className="btn-primary mr-2">
            {submitting ? 'Creating...' : 'Create'}
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
        <p data-testid="workflow-templates-loading">Loading...</p>
      ) : templates.length === 0 ? (
        <p data-testid="workflow-templates-empty">No workflow templates found.</p>
      ) : (
        <table className="w-full border-collapse">
          <thead>
            <tr>
              <th className="border px-4 py-2 text-left">Name</th>
              <th className="border px-4 py-2 text-left">content_type</th>
              <th className="border px-4 py-2 text-left">category</th>
              <th className="border px-4 py-2 text-left">status</th>
            </tr>
          </thead>
          <tbody>
            {templates.map((t: WorkflowTemplateResponse) => (
              <tr key={t.id}>
                <td className="border px-4 py-2">
                  <a href={`/workflow/templates/${t.id}`} className="text-blue-600 hover:underline">{t.name}</a>
                </td>
                <td className="border px-4 py-2">{t.content_type}</td>
                <td className="border px-4 py-2">{t.category}</td>
                <td className="border px-4 py-2">{t.status}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
