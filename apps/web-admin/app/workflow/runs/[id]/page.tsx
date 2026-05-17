'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import {
  fetchWorkflowRun,
  fetchWorkflowRunSteps,
  cancelWorkflowRun,
  retryWorkflowRun,
  pageErrorFromEnvelope,
  redactSensitive,
  PageError,
  WorkflowRunResponse,
  WorkflowStepRunResponse,
} from '@/lib/api'

export default function WorkflowRunDetailPage() {
  const params = useParams()
  const id = params?.id as string
  const [run, setRun] = useState<(WorkflowRunResponse & { input: Record<string, unknown>; output: Record<string, unknown>; error?: string; step_count: number; agent_task_count: number }) | null>(null)
  const [steps, setSteps] = useState<WorkflowStepRunResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [action, setAction] = useState('')
  const [success, setSuccess] = useState('')
  const [error, setError] = useState<PageError | null>(null)

  const load = async () => {
    setLoading(true)
    try {
      const [r, stepsResp] = await Promise.all([
        fetchWorkflowRun(id),
        fetchWorkflowRunSteps(id),
      ])
      if (!r.success || !r.data) {
        setError(pageErrorFromEnvelope(r, 'Failed to load workflow run'))
        setRun(null)
        setSteps([])
        return
      }
      if (!stepsResp.success || !stepsResp.data) {
        setError(pageErrorFromEnvelope(stepsResp, 'Failed to load workflow run steps'))
        setRun(null)
        setSteps([])
        return
      }
      setError(null)
      setRun(r.data)
      setSteps(stepsResp.data.items)
    } catch {
      setError({ message: 'Failed to load workflow run' })
      setRun(null)
      setSteps([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { if (id) load() }, [id])

  const handleCancel = async () => {
    setAction('cancel')
    setSuccess('')
    try {
      const res = await cancelWorkflowRun(id, { reason: 'user_request' })
      if (!res.success) {
        setError(pageErrorFromEnvelope(res, 'Failed to cancel workflow run'))
        return
      }
      setError(null)
      setSuccess(`Cancelled workflow run ${id}`)
    } catch {
      setError({ message: 'Failed to cancel workflow run' })
      return
    } finally {
      setAction('')
    }
    load()
  }

  const handleRetry = async () => {
    setAction('retry')
    setSuccess('')
    try {
      const res = await retryWorkflowRun(id, { reason: 'user_request' })
      if (!res.success) {
        setError(pageErrorFromEnvelope(res, 'Failed to retry workflow run'))
        return
      }
      setError(null)
      setSuccess(`Created retry workflow run ${res.data?.new_workflow_run_id ?? ''}`.trim())
    } catch {
      setError({ message: 'Failed to retry workflow run' })
      return
    } finally {
      setAction('')
    }
    load()
  }

  if (loading) return <p className="p-6">Loading...</p>
  if (!run) {
    return (
      <div className="p-6">
        {error ? (
          <div className="mb-4 text-red-600">
            <p>{error.message}</p>
            {error.code && <p>code: {error.code}</p>}
            {error.request_id && <p>request_id: {error.request_id}</p>}
          </div>
        ) : (
          <p>Run not found</p>
        )}
      </div>
    )
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-4">
        <div>
          <h1 className="text-2xl font-bold">Run {run.id}</h1>
          <p className="text-gray-500">
            step_count: {run.step_count} | agent_task_count: {run.agent_task_count} | Status: {run.status}
          </p>
        </div>
        <div className="flex gap-2">
          {run.status === 'running' && (
            <button onClick={handleCancel} disabled={Boolean(action)} className="btn-danger">
              {action === 'cancel' ? 'Cancelling...' : 'cancelWorkflowRun'}
            </button>
          )}
          {run.status === 'failed' && (
            <button onClick={handleRetry} disabled={Boolean(action)} className="btn-primary">
              {action === 'retry' ? 'Retrying...' : 'retryWorkflowRun'}
            </button>
          )}
        </div>
      </div>

      {success && <p className="mb-4 text-green-600">{success}</p>}
      {error && (
        <div className="mb-4 text-red-600">
          <p>{error.message}</p>
          {error.code && <p>code: {error.code}</p>}
          {error.request_id && <p>request_id: {error.request_id}</p>}
        </div>
      )}

      <div className="grid grid-cols-2 gap-4 mb-6">
        <section className="border rounded p-4">
          <h2 className="font-semibold mb-2">Input</h2>
          <pre className="whitespace-pre-wrap text-sm">{JSON.stringify(redactSensitive(run.input), null, 2)}</pre>
        </section>
        <section className="border rounded p-4">
          <h2 className="font-semibold mb-2">Output</h2>
          <pre className="whitespace-pre-wrap text-sm">{JSON.stringify(redactSensitive(run.output), null, 2)}</pre>
        </section>
      </div>
      {run.error && <p className="mb-4 text-red-600">Error: {String(redactSensitive(run.error))}</p>}

      <h2 className="text-xl font-semibold mb-3">Steps</h2>
      <table className="w-full border-collapse">
        <thead>
          <tr>
            <th className="border px-4 py-2 text-left">Step ID</th>
            <th className="border px-4 py-2 text-left">step_template_id</th>
            <th className="border px-4 py-2 text-left">Status</th>
          </tr>
        </thead>
        <tbody>
          {steps.map((s: WorkflowStepRunResponse) => (
            <tr key={s.id}>
              <td className="border px-4 py-2">{s.id}</td>
              <td className="border px-4 py-2">{s.step_template_id}</td>
              <td className="border px-4 py-2">{s.status}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
