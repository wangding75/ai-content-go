# ai-content-go — API Specification

## API Style
REST JSON API planned under `/api/v1`.

## Base URL Pattern
```text
/api/v1/{resource}
```

## Response Format
Successful responses must follow:
```json
{
  "success": true,
  "data": {},
  "error": null,
  "request_id": "req_20260514100000001"
}
```

Failed responses must follow:
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request body.",
    "details": []
  },
  "request_id": "req_20260514100000002"
}
```

## Error Code Convention
Common codes include `VALIDATION_ERROR`, `UNAUTHORIZED`, `FORBIDDEN`, `NOT_FOUND`, `CONFLICT`, `IDEMPOTENCY_CONFLICT`, `WORKFLOW_RUN_FAILED`, `AGENT_OUTPUT_INVALID`, `LLM_PROVIDER_ERROR`, `EXTERNAL_AUTOMATION_ERROR`, and `INTERNAL_ERROR`.

## Authentication
Bearer token authentication is required by the API contract; implementation is TBD.
