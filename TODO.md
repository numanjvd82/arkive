# Uploads TODO

## Not Implemented Yet
- Abuse safeguards: new account upload limits (age/verification gates) with progressive unlocks.
- Abuse safeguards: per-token and per-IP limits for public share downloads, with burst control.
- Abuse safeguards: heuristic detection for upload spikes, repeated failures, and high share churn.
- Periodic cleanup for stale multipart uploads (database records + R2 uploads/objects).
- Upload lifecycle metrics/logging (start/part/complete/abort/fail counters and durations).
- Automated tests for multipart edge cases (concurrent abort/complete, resume after partial failure, R2 transient errors).
- Integrity checks after completion (verify size/hash matches expected, fail if mismatch).
- Extract image metadata via `ffprobe` and store it alongside files.
- Generate video thumbnails during processing.
- Generate image thumbnails during processing.
- Multi-file and folder upload UI (queue, progress per item, folder support).
- Alert component for inline form and page messaging.
- Refactor share UI (dialog + public share view) for clearer components and state management.
- Add ad provider name to Privacy Policy and Cookie Policy once finalized.
- Add governing law and dispute venue to Terms of Service after company registration.
- Add company address and phone to Terms of Service after company registration.
- Add cross-links between Privacy Policy, Cookie Policy, Terms of Service, and AUP.
- Add revision dates to Terms of Service and AUP pages.
- Add rewards disclaimer to Privacy Policy (availability, fraud checks, subject to change).

## Done
- Install `ffprobe` on the server for video metadata extraction.
- Rate limiting for upload endpoints and background media processing.
- Add Cloudflare analytics script to main production layout.
- Refactor media view page (components, helpers) for better structure and maintainability.
- Strategy for private/public view and download expiry (stable media URLs, refresh flow, tokenized CDN).
- Improve toast visibility and messaging for rate-limited actions across the app.
