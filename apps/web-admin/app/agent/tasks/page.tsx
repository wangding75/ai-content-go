'use client'

import { useEffect, useState } from 'react'
import {
  fetchAgentTasks,
  pageErrorFromEnvelope,
  PageError,
  AgentTaskResponse,
} from '@/lib/api'

export default function AgentTasksPage() {
  const [tasks, setTasks] = useState<AgentTaskResponse[]>([])
  const [loading, setLoading] = useState(true)
  const [workflowRunId, setWorkflowRunId] = useState('')
  const [stepRunId, setStepRunId] = useState('')
  const [agentCode, setAgentCode] = useState('')
  const [status, setStatus] = useState('')
  const [error, setError] = useState<PageError | null>(null)

  const load = async () => {
    setLoading(true)
    try {
      const res = await fetchAgentTasks({
        workflow_run_id: workflowRunId,
        step_run_id: stepRunId,
        agent_code: agentCode,
        status,
      })
      if (!res.success || !res.data) {
        setError(pageErrorFromEnvelope(res, 'Failed to load agent tasks'))
        setTasks([])
        return
      }
      setError(null)
      setTasks(res.data.items)
    } catch {
      setError({ message: 'Failed to load agent tasks' })
      setTasks([])
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { load() }, [workflowRunId, stepRunId, agentCode, status])

  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold mb-4">Agent Tasks</h1>

      <div className="flex gap-4 mb-4">
        <input
          placeholder="workflow_run_id"
          value={workflowRunId}
          onChange={e => setWorkflowRunId(e.target.value)}
          className="border px-2 py-1 rounded"
        />
        <input
          placeholder="step_run_id"
          value={stepRunId}
          onChange={e => setStepRunId(e.target.value)}
          className="border px-2 py-1 rounded"
        />
        <input
          placeholder="agent_code"
          value={agentCode}
          onChange={e => setAgentCode(e.target.value)}
          className="border px-2 py-1 rounded"
        />
        <select value={status} onChange={e => setStatus(e.target.value)} className="border px-2 py-1 rounded">
          <option value="">All status</option>
          <option value="pending">pending</option>
          <option value="running">running</option>
          <option value="success">success</option>
          <option value="failed">failed</option>
          <option value="cancelled">cancelled</option>
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
        <p data-testid="agent-tasks-loading">Loading...</p>
      ) : tasks.length === 0 ? (
        <p data-testid="agent-tasks-empty">No agent tasks found.</p>
      ) : (
        <table className="w-full border-collapse">
          <thead>
            <tr>
              <th className="border px-4 py-2 text-left">ID</th>
              <th className="border px-4 py-2 text-left">workflow_run_id</th>
              <th className="border px-4 py-2 text-left">step_run_id</th>
              <th className="border px-4 py-2 text-left">agent_code</th>
              <th className="border px-4 py-2 text-left">status</th>
            </tr>
          </thead>
          <tbody>
            {tasks.map((task: AgentTaskResponse) => (
              <tr key={task.id}>
                <td className="border px-4 py-2">
                  <a href={`/agent/tasks/${task.id}`} className="text-blue-600 hover:underline">{task.id}</a>
                </td>
                <td className="border px-4 py-2">{task.workflow_run_id}</td>
                <td className="border px-4 py-2">{task.step_run_id}</td>
                <td className="border px-4 py-2">{task.agent_code}</td>
                <td className="border px-4 py-2">{task.status}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
