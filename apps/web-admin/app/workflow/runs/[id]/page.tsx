'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import {
  fetchWorkflowRun,
  fetchWorkflowRunSteps,
  cancelWorkflowRun,
  retryWorkflowRun,
  WorkflowStepRunResponse,
} from '@/lib/api'

export default function WorkflowRunDetailPage() {
  const params = useParams()
  const id = params?.id as string
  const [run, setRun] = useState<any>(null)
  const [steps, setSteps] = useState<WorkflowStepRunResponse[]>([])
  const [loading, setLoading] = useState(true)

  const load = async () => {
    setLoading(true)
    try {
      const [r, stepsResp] = await Promise.all([
        fetchWorkflowRun(id),
        fetchWorkflowRunSteps(id),
      ])
      setRun(r)
      setSteps(stepsResp.items ?? [])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { if (id) load() }, [id])

  const handleCancel = async () => {
    await cancelWorkflowRun(id)
    load()
  }

  const handleRetry = async () => {
    await retryWorkflowRun(id)
    load()
  }

  if (loading) return <p className="p-6">Loading...</p>
  if (!run) return <p className="p-6">Run not found</p>

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
            <button onClick={handleCancel} className="btn-danger">cancelWorkflowRun</button>
          )}
          {run.status === 'failed' && (
            <button onClick={handleRetry} className="btn-primary">retryWorkflowRun</button>
          )}
        </div>
      </div>

      <h2 className="text-xl font-semibold mb-3">Steps</h2>
      <table className="w-full border-collapse">
        <thead>
          <tr>
            <th className="border px-4 py-2 text-left">Step ID</th>
            <th className="border px-4 py-2 text-left">Type</th>
            <th className="border px-4 py-2 text-left">Status</th>
          </tr>
        </thead>
        <tbody>
          {steps.map((s: WorkflowStepRunResponse) => (
            <tr key={s.id}>
              <td className="border px-4 py-2">{s.id}</td>
              <td className="border px-4 py-2">{s.step_type}</td>
              <td className="border px-4 py-2">{s.status}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
