# Rate Limits

This document lists the per-route rate limits enforced by `core/middleware/ratelimit.go`
and configured in `core/router/router.go`. Limits are expressed as requests per minute
with a burst allowance. Premium users receive higher limits (not skipped).

## Rationale
- Upload endpoints have the highest abuse potential, so they are tighter.
- Share settings are moderate to keep the UI responsive without encouraging spam.
- Downloads/media are higher for authenticated users to support normal usage.
- Public share pages are separate and stay conservative to limit scraping.

## Uploads (authenticated)
- `POST /api/uploads/start`: 6/min, burst 10
  - Premium: 30/min, burst 60
- `POST /api/uploads/:id/next`: 120/min, burst 240
  - Premium: 600/min, burst 1200
  - Chunk size is 10MB+ (`pkg/storage/upload.go`), so 120/min is sufficient.
- `POST /api/uploads/:id/complete`: 10/min, burst 20
- `POST /api/uploads/:id/cancel`: 10/min, burst 20

## Files + Shares (authenticated)
- `GET /api/files/:id/share`: 60/min, burst 120
- `POST /api/files/:id/share`: 10/min, burst 20
- `DELETE /api/files/:id`: 30/min, burst 60
- `POST /api/shares/:id/revoke`: 20/min, burst 40
- `PATCH /api/shares/:id`: 20/min, burst 40
- `DELETE /api/shares/:id`: 20/min, burst 40

## Downloads + Media (authenticated)
- `GET /api/files/:id/download`: 120/min, burst 240
  - Premium: 600/min, burst 1200
- `GET /api/files/:id/media`: 120/min, burst 240
  - Premium: 600/min, burst 1200

## Public Share Pages (anonymous)
- `GET /s/:token`: 2/min, burst 2
- `POST /s/:token`: 2/min, burst 2

## Notes
- There are no anonymous `/api/files/:id/download` or `/api/files/:id/media` routes.
  Public access uses `/s/:token`, which is rate-limited separately.
- Premium handling is configured per-route via `PremiumRPM` and `PremiumBurst`.
- The 429 penalty box rule is planned but not implemented yet. When enabled:
  - If a key triggers 429 more than 5 times in a 60-second window, it is blocked.
  - Block duration is 2 minutes.
  - Responses remain 429 with the standard `Retry-After` header.
