'use client'

import { useEffect, useState } from 'react'
import {
  fetchWorkflowTemplates,
  createWorkflowTemplate,
  WorkflowTemplateResponse,
} from '@/lib/api'

export default function WorkflowTemplatesPage() {
  const [templates, setTemplates] = useState<WorkflowTemplateResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [contentType, setContentType] = useState('')
  const [category, setCategory] = useState('')
  const [status, setStatus] = useState('')
  const [creating, setCreating] = useState(false)
  const [newName, setNewName] = useState('')
  const [newContentType, setNewContentType] = useState('')
  const [newCategory, setNewCategory] = useState('')

  const load = async () => {
    setLoading(true)
    try {
      const res = await fetchWorkflowTemplates({ content_type: contentType, category, status })
      setTemplates(res.items ?? [])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [contentType, category, status])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    await createWorkflowTemplate({ name: newName, content_type: newContentType, category: newCategory })
    setCreating(false)
    setNewName('')
    setNewContentType('')
    setNewCategory('')
    load()
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-4">
        <h1 className="text-2xl font-bold">Workflow Templates</h1>
        <button onClick={() => setCreating(true)} className="btn-primary">New Template</button>
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
          <button type="submit" className="btn-primary mr-2">Create</button>
          <button type="button" onClick={() => setCreating(false)}>Cancel</button>
        </form>
      )}

      {loading ? (
        <p>Loading...</p>
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
