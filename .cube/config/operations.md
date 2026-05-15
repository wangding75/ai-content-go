# ai-content-go — Operations

## Observability
TBD. Requirements mandate `AgentTask` and `LLMCallLog` tracking for input, output, model, token usage, cost, and errors.

## Monitoring
No monitoring configuration exists yet.

## Alerts
TBD. n8n may be used for notifications, webhooks, external API sync, and alerts, but not core workflow orchestration.

## Runbooks
TBD.

## Operational Constraints
- Core workflow state must remain inside AI Content Factory.
- Human confirmation must be supported for review, publishing, and strategy execution.
