# e-timologisi-utils ergani and payments 
## logging

Shared structured logging for every service (one JSON line per event, request id, whole
payloads with credentials masked, CloudWatch-friendly). Set `SERVICE_NAME` per service.
Fields: `timestamp, level, service, component, event, request_id, msg` + event fields.
Use `logging.Middleware` on the HTTP router and pass `r.Context()` down.
