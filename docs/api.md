# API

The full OpenAPI spec lives at [`../internal/api/http/docs/openapi.yaml`](../internal/api/http/docs/openapi.yaml).

When the service is running, it also serves the spec and Swagger UI:

- `GET /openapi.yaml` - raw spec
- `GET /docs` - Swagger UI

Local URLs:

- <http://localhost:8080/openapi.yaml>
- <http://localhost:8080/docs>

See [Quickstart](quickstart.md) for curl examples.

## Endpoint Overview

All API endpoints below are under `/api/v1`.

```text
GET    /auth/me
POST   /auth/login
POST   /auth/logout

GET    /environments
POST   /environments
GET    /environments/{id}
PUT    /environments/{id}
DELETE /environments/{id}

GET    /flags
POST   /flags
GET    /flags/{key}
DELETE /flags/{key}

GET    /environments/{environment_id}/flags
POST   /environments/{environment_id}/flags
GET    /environments/{environment_id}/flags/{key}
DELETE /environments/{environment_id}/flags/{key}
POST   /environments/{environment_id}/flags/{key}/evaluate

GET    /flags/{key}/rules
POST   /flags/{key}/rules
POST   /flags/{key}/rules/reorder
GET    /flags/{key}/rules/{id}
PUT    /flags/{key}/rules/{id}
DELETE /flags/{key}/rules/{id}

GET    /environments/{environment_id}/flags/{key}/rules
POST   /environments/{environment_id}/flags/{key}/rules
POST   /environments/{environment_id}/flags/{key}/rules/reorder
GET    /environments/{environment_id}/flags/{key}/rules/{id}
PUT    /environments/{environment_id}/flags/{key}/rules/{id}
DELETE /environments/{environment_id}/flags/{key}/rules/{id}

GET    /contexts
POST   /contexts
GET    /contexts/{id}
PUT    /contexts/{id}
DELETE /contexts/{id}

POST   /eval
POST   /eval/{key}
GET    /eval/definitions

GET    /api-keys
POST   /api-keys
DELETE /api-keys/{id}
```

Evaluation API keys are scoped to one environment. SDK and `/eval` callers do
not pass an environment in the request; the bearer token selects it. Existing
`/flags...` admin endpoints remain as compatibility aliases for the default
`production` environment.
