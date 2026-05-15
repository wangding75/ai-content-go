'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import {
  fetchWorkflowTemplate,
  fetchWorkflowVersions,
  createWorkflowVersion,
  publishWorkflowVersion,
  WorkflowTemplateVersionResponse,
} from '@/lib/api'

export default function WorkflowTemplateDetailPage() {
  const params = useParams()
  const id = params?.id as string
  const [template, setTemplate] = useState<any>(null)
  const [versions, setVersions] = useState<WorkflowTemplateVersionResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [creating, setCreating] = useState(false)
  const [newDefinition, setNewDefinition] = useState('')

  const load = async () => {
    setLoading(true)
    try {
      const [tmpl, vResp] = await Promise.all([
        fetchWorkflowTemplate(id),
        fetchWorkflowVersions(id),
      ])
      setTemplate(tmpl)
      setVersions(vResp.items ?? [])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { if (id) load() }, [id])

  const handleCreateVersion = async (e: React.FormEvent) => {
    e.preventDefault()
    const idempotencyKey = `create-version-${id}-${Date.now()}`
    await createWorkflowVersion(id, { definition: JSON.parse(newDefinition || '{}') }, { 'Idempotency-Key': idempotencyKey })
    setCreating(false)
    setNewDefinition('')
    load()
  }

  const handlePublish = async (versionId: string) => {
    const idempotencyKey = `publish-${versionId}-${Date.now()}`
    await publishWorkflowVersion(id, versionId, { 'Idempotency-Key': idempotencyKey })
    load()
  }

  if (loading) return <p className="p-6">Loading...</p>
  if (!template) return <p className="p-6">Template not found</p>

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-2">{template.name}</h1>
      <p className="text-gray-500 mb-6">
        {template.content_type} / {template.category} — {template.status}
      </p>

      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-semibold">Versions</h2>
        <button onClick={() => setCreating(true)} className="btn-primary">New Version</button>
      </div>

      {creating && (
        <form onSubmit={handleCreateVersion} className="mb-4 p-4 border rounded">
          <textarea
            placeholder='{"steps": []}'
            value={newDefinition}
            onChange={e => setNewDefinition(e.target.value)}
            className="border px-2 py-1 rounded w-full mb-2 h-24"
          />
          <button type="submit" className="btn-primary mr-2">Create Version</button>
          <button type="button" onClick={() => setCreating(false)}>Cancel</button>
        </form>
      )}

      <table className="w-full border-collapse">
        <thead>
          <tr>
            <th className="border px-4 py-2 text-left">ID</th>
            <th className="border px-4 py-2 text-left">Version</th>
            <th className="border px-4 py-2 text-left">Status</th>
            <th className="border px-4 py-2 text-left">Actions</th>
          </tr>
        </thead>
        <tbody>
          {versions.map((v: WorkflowTemplateVersionResponse) => (
            <tr key={v.id}>
              <td className="border px-4 py-2">{v.id}</td>
              <td className="border px-4 py-2">{v.version}</td>
              <td className="border px-4 py-2">{v.status}</td>
              <td className="border px-4 py-2">
                {v.status === 'draft' && (
                  <button
                    onClick={() => handlePublish(v.id)}
                    className="text-blue-600 hover:underline"
                  >
                    publishWorkflowVersion
                  </button>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
