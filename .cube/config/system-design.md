# ai-content-go — System Design

## Technology Stack
| Layer | Technology | Version |
|-------|-----------|---------|
| Language | Go | 1.22+ required; installed 1.22.3 |
| HTTP API | Chi recommended | TBD |
| Frontend | Next.js + TypeScript | TBD |
| Database | PostgreSQL | TBD |
| Cache / Queue | Redis / asynq | TBD |
| API Docs | OpenAPI 3.0 | TBD |

## Architecture Pattern
Greenfield layered architecture planned from requirements: Web Admin / Browser Extension → Go HTTP API → Content Business Service → Workflow Engine → Agent Runtime → Knowledge Memory → LLM Router → PostgreSQL / Redis / Object Storage.

## Layer Structure
```text
Web Admin / Browser Extension
  ↓
API Gateway / Go HTTP API
  ↓
Content Business Service
  ↓
Workflow Engine
  ↓
Agent Runtime
  ↓
Knowledge Memory
  ↓
LLM Router
  ↓
PostgreSQL / Redis / Object Storage
```

## Key Base Classes
| Class | Package | Purpose | Subclasses Must |
|-------|---------|---------|----------------|
| TBD | TBD | No source code exists yet | TBD |

## Public Components
| Component | Package | Purpose |
|-----------|---------|---------|
| API response wrapper | TBD | Must follow `success/data/error/request_id` contract |
| Pagination wrapper | TBD | Must expose `items` and `pagination` |

## Configuration
- Config location: TBD
- Profiles: TBD
- Database: PostgreSQL planned; connection pattern TBD
