'use client'

import { useEffect, useState } from 'react'
import {
  fetchLLMCallLogs,
  pageErrorFromEnvelope,
  PageError,
  LLMCallLogResponse,
} from '@/lib/api'

export default function LLMLogsPage() {
  const [logs, setLogs] = useState<LLMCallLogResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [workflowRunId, setWorkflowRunId] = useState('')
  const [agentTaskId, setAgentTaskId] = useState('')
  const [provider, setProvider] = useState('')
  const [model, setModel] = useState('')
  const [status, setStatus] = useState('')
  const [error, setError] = useState<PageError | null>(null)

  const load = async () => {
    setLoading(true)
    try {
      const res = await fetchLLMCallLogs({
        workflow_run_id: workflowRunId,
        agent_task_id: agentTaskId,
        provider,
        model,
        status,
      })
      if (!res.success || !res.data) {
        setError(pageErrorFromEnvelope(res, 'Failed to load LLM call logs'))
        setLogs([])
        return
      }
      setError(null)
      setLogs(res.data.items)
    } catch {
      setError({ message: 'Failed to load LLM call logs' })
      setLogs([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [workflowRunId, agentTaskId, provider, model, status])

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-4">LLM Call Logs</h1>

      <div className="flex gap-4 mb-4">
        <input
          placeholder="workflow_run_id"
          value={workflowRunId}
          onChange={e => setWorkflowRunId(e.target.value)}
          className="border px-2 py-1 rounded"
        />
        <input
          placeholder="agent_task_id"
          value={agentTaskId}
          onChange={e => setAgentTaskId(e.target.value)}
          className="border px-2 py-1 rounded"
        />
        <input
          placeholder="provider"
          value={provider}
          onChange={e => setProvider(e.target.value)}
          className="border px-2 py-1 rounded"
        />
        <input
          placeholder="model"
          value={model}
          onChange={e => setModel(e.target.value)}
          className="border px-2 py-1 rounded"
        />
        <select value={status} onChange={e => setStatus(e.target.value)} className="border px-2 py-1 rounded">
          <option value="">All status</option>
          <option value="success">success</option>
          <option value="failed">failed</option>
        </select>
      </div>

      {error && (
        <div role="alert" className="mb-4 text-red-600">
          <p>{error.message}</p>
          {error.code && <p>code: {error.code}</p>}
          {error.request_id && <p>request_id: {error.request_id}</p>}
        </div>
      )}

      {loading ? (
        <p data-testid="llm-logs-loading">Loading...</p>
      ) : logs.length === 0 ? (
        <p data-testid="llm-logs-empty">No LLM call logs found.</p>
      ) : (
        <table className="w-full border-collapse">
          <thead>
            <tr>
              <th className="border px-4 py-2 text-left">ID</th>
              <th className="border px-4 py-2 text-left">provider</th>
              <th className="border px-4 py-2 text-left">model</th>
              <th className="border px-4 py-2 text-left">input_tokens</th>
              <th className="border px-4 py-2 text-left">output_tokens</th>
              <th className="border px-4 py-2 text-left">cost</th>
              <th className="border px-4 py-2 text-left">latency_ms</th>
              <th className="border px-4 py-2 text-left">status</th>
            </tr>
          </thead>
          <tbody>
            {logs.map((log: LLMCallLogResponse) => (
              <tr key={log.id}>
                <td className="border px-4 py-2">{log.id}</td>
                <td className="border px-4 py-2">{log.provider}</td>
                <td className="border px-4 py-2">{log.model}</td>
                <td className="border px-4 py-2">{log.input_tokens}</td>
                <td className="border px-4 py-2">{log.output_tokens}</td>
                <td className="border px-4 py-2">{log.cost}</td>
                <td className="border px-4 py-2">{log.latency_ms}</td>
                <td className="border px-4 py-2">{log.status}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
