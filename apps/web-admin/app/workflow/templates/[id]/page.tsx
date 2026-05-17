'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import {
  fetchWorkflowTemplate,
  fetchWorkflowVersions,
  createWorkflowVersion,
  publishWorkflowVersion,
  pageErrorFromEnvelope,
  PageError,
  WorkflowTemplateResponse,
  WorkflowTemplateVersionResponse,
} from '@/lib/api'

export default function WorkflowTemplateDetailPage() {
  const params = useParams()
  const id = params?.id as string
  const [template, setTemplate] = useState<WorkflowTemplateResponse | null>(null)
  const [versions, setVersions] = useState<WorkflowTemplateVersionResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [creating, setCreating] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [publishingID, setPublishingID] = useState('')
  const [success, setSuccess] = useState('')
  const [newDefinition, setNewDefinition] = useState('')
  const [error, setError] = useState<PageError | null>(null)

  const load = async () => {
    setLoading(true)
    try {
      const [tmpl, vResp] = await Promise.all([
        fetchWorkflowTemplate(id),
        fetchWorkflowVersions(id),
      ])
      if (!tmpl.success || !tmpl.data) {
        setError(pageErrorFromEnvelope(tmpl, 'Failed to load workflow template'))
        setTemplate(null)
        setVersions([])
        return
      }
      if (!vResp.success || !vResp.data) {
        setError(pageErrorFromEnvelope(vResp, 'Failed to load workflow template versions'))
        setTemplate(null)
        setVersions([])
        return
      }
      setError(null)
      setTemplate(tmpl.data)
      setVersions(vResp.data.items)
    } catch {
      setError({ message: 'Failed to load workflow template' })
      setTemplate(null)
      setVersions([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { if (id) load() }, [id])

  const handleCreateVersion = async (e: React.FormEvent) => {
    e.preventDefault()
    let definition: { steps?: Array<{ step_code: string; step_type: string; agent_code?: string; order_index: number }> }
    try {
      definition = JSON.parse(newDefinition || '{"steps": []}') as { steps?: Array<{ step_code: string; step_type: string; agent_code?: string; order_index: number }> }
    } catch {
      setError({ message: 'Version definition must be valid JSON' })
      return
    }
    setSubmitting(true)
    setSuccess('')
    try {
      const res = await createWorkflowVersion(id, { steps: definition.steps ?? [] })
      if (!res.success) {
        setError(pageErrorFromEnvelope(res, 'Failed to create workflow version'))
        return
      }
      setError(null)
      setSuccess(`Created workflow version ${res.data?.template_version_id ?? ''}`.trim())
    } catch {
      setError({ message: 'Failed to create workflow version' })
      return
    } finally {
      setSubmitting(false)
    }
    setCreating(false)
    setNewDefinition('')
    load()
  }

  const handlePublish = async (versionId: string) => {
    const headers = { 'Idempotency-Key': `publish-${versionId}-${Date.now()}` }
    setPublishingID(versionId)
    setSuccess('')
    try {
      const res = await publishWorkflowVersion(versionId, {}, headers['Idempotency-Key'])
      if (!res.success) {
        setError(pageErrorFromEnvelope(res, 'Failed to publish workflow version'))
        return
      }
      setError(null)
      setSuccess(`Published workflow version ${versionId}`)
    } catch {
      setError({ message: 'Failed to publish workflow version' })
      return
    } finally {
      setPublishingID('')
    }
    load()
  }

  if (loading) return <p className="p-6">Loading...</p>
  if (!template) {
    return (
      <div className="p-6">
        {error ? (
          <div className="mb-4 text-red-600">
            <p>{error.message}</p>
            {error.code && <p>code: {error.code}</p>}
            {error.request_id && <p>request_id: {error.request_id}</p>}
          </div>
        ) : (
          <p>Template not found</p>
        )}
      </div>
    )
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-2">{template.name}</h1>
      <p className="text-gray-500 mb-6">
        {template.content_type} / {template.category} — {template.status}
      </p>

      {success && <p className="mb-4 text-green-600">{success}</p>}
      {error && (
        <div className="mb-4 text-red-600">
          <p>{error.message}</p>
          {error.code && <p>code: {error.code}</p>}
          {error.request_id && <p>request_id: {error.request_id}</p>}
        </div>
      )}

      <div className="flex justify-between items-center mb-4">
        <h2 className="text-xl font-semibold">Versions</h2>
        <button onClick={() => setCreating(true)} disabled={submitting || Boolean(publishingID)} className="btn-primary">New Version</button>
      </div>

      {creating && (
        <form onSubmit={handleCreateVersion} className="mb-4 p-4 border rounded">
          <textarea
            placeholder='{"steps": []}'
            value={newDefinition}
            onChange={e => setNewDefinition(e.target.value)}
            className="border px-2 py-1 rounded w-full mb-2 h-24"
          />
          <button type="submit" disabled={submitting} className="btn-primary mr-2">
            {submitting ? 'Creating...' : 'Create Version'}
          </button>
          <button type="button" disabled={submitting} onClick={() => setCreating(false)}>Cancel</button>
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
                    disabled={Boolean(publishingID)}
                    className="text-blue-600 hover:underline disabled:text-gray-400"
                  >
                    {publishingID === v.id ? 'Publishing...' : 'publishWorkflowVersion'}
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
