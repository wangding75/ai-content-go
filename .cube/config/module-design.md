# ai-content-go — Module Design

## Module List
| Module | Package | Purpose | Key Classes |
|--------|---------|---------|-------------|
| content | TBD | ContentProject, ContentType, ContentItem, ContentAsset | TBD |
| workflow | TBD | WorkflowTemplate, WorkflowRun, StepRun, Schedule, ProductionPlan | TBD |
| agent | TBD | AgentTask and prompt execution tracking | TBD |
| llm | TBD | Provider routing, fallback, token/cost logs | TBD |
| memory | TBD | Static context, dynamic state, recent content, style guides | TBD |
| publish | TBD | PublishTarget and PublishJob | TBD |
| metrics | TBD | MetricRecord and dashboard data | TBD |
| strategy | TBD | StrategySuggestion workflow | TBD |

## Module Dependencies
Planned flow: content modules drive workflow runs; workflow invokes agent runtime; agent runtime calls LLM router and memory; publishing, metrics, and strategy consume content and run state.

## Data Model
| Entity | Table | Key Fields |
|--------|-------|------------|
| ContentProject | TBD | id, name, content_type_id, target_platform, status |
| ContentType | TBD | id, code, name, pack |
| ContentItem | TBD | id, project_id, type, status, content |
| WorkflowRun | TBD | id, template_id, project_id, status |
| AgentTask | TBD | id, run_id, input, output, model, tokens, cost, error |
| LLMCallLog | TBD | id, provider, model, prompt_tokens, completion_tokens, cost |
| PublishJob | TBD | id, content_item_id, target_id, status |
| MetricRecord | TBD | id, project_id, metric_key, value, recorded_at |
| StrategySuggestion | TBD | id, project_id, action, status |
