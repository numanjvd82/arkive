# Uploads TODO

## Not Implemented Yet
- Periodic cleanup for stale multipart uploads (database records + R2 uploads/objects).
- Upload lifecycle metrics/logging (start/part/complete/abort/fail counters and durations).
- Automated tests for multipart edge cases (concurrent abort/complete, resume after partial failure, R2 transient errors).
- Integrity checks after completion (verify size/hash matches expected, fail if mismatch).
- Strategy for private/public view and download expiry (stable media URLs, refresh flow, tokenized CDN).
