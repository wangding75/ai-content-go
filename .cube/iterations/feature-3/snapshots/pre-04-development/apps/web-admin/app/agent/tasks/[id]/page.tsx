'use client'

import { useEffect, useState } from 'react'
import { useParams } from 'next/navigation'
import {
  fetchAgentTask,
  pageErrorFromEnvelope,
  redactSensitive,
  PageError,
  AgentTaskDetailResponse,
} from '@/lib/api'

export default function AgentTaskDetailPage() {
  const params = useParams()
  const id = params?.id as string
  const [task, setTask] = useState<AgentTaskDetailResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<PageError | null>(null)

  const load = async () => {
    setLoading(true)
    try {
      const res = await fetchAgentTask(id)
      if (!res.success || !res.data) {
        setError(pageErrorFromEnvelope(res, 'Failed to load agent task'))
        setTask(null)
        return
      }
      setError(null)
      setTask(res.data)
    } catch {
      setError({ message: 'Failed to load agent task' })
      setTask(null)
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { if (id) load() }, [id])

  if (loading) return <p className="p-6">Loading...</p>
  if (!task) {
    return (
      <div className="p-6">
        {error ? (
          <div className="mb-4 text-red-600">
            <p>{error.message}</p>
            {error.code && <p>code: {error.code}</p>}
            {error.request_id && <p>request_id: {error.request_id}</p>}
          </div>
        ) : (
          <p>Agent task not found</p>
        )}
      </div>
    )
  }

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-2">Agent Task {task.id}</h1>
      <p className="text-gray-500 mb-6">
        workflow_run_id: {task.workflow_run_id} | step_run_id: {task.step_run_id} | agent_code: {task.agent_code} | status: {task.status}
      </p>

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
          <pre className="whitespace-pre-wrap text-sm">{JSON.stringify(redactSensitive(task.input), null, 2)}</pre>
        </section>
        <section className="border rounded p-4">
          <h2 className="font-semibold mb-2">Output</h2>
          <pre className="whitespace-pre-wrap text-sm">{JSON.stringify(redactSensitive(task.output), null, 2)}</pre>
        </section>
      </div>

      {task.error && <p className="mb-4 text-red-600">Error: {String(redactSensitive(task.error))}</p>}
      <p className="mb-4">started_at: {task.started_at ?? '-'} | finished_at: {task.finished_at ?? '-'}</p>

      <h2 className="text-xl font-semibold mb-3">LLM Call Logs</h2>
      {task.llm_call_log_ids.length === 0 ? (
        <p>No llm_call_log_ids.</p>
      ) : (
        <ul className="list-disc pl-5">
          {task.llm_call_log_ids.map(logID => (
            <li key={logID}>{logID}</li>
          ))}
        </ul>
      )}
    </div>
  )
}
