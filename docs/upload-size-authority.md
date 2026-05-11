# Upload Size Authority

Arkive must use server-counted bytes, not browser-reported size, as upload source of truth.

## Rule

- Browser `file.size` is hint only.
- Authoritative size is bytes actually received/stored by backend or storage provider.
- Zero-knowledge is not affected. This only uses byte counts, not file contents.

## Why

Client-reported size is easy to spoof.

Examples:

- Browser says `100 MB`, sends `6 GB`
- Browser declares size that fits quota, real stored object exceeds quota

Server must enforce limits from real bytes, not declared bytes.

## Local Storage Flow

For local storage, count bytes while writing.

Use `http.MaxBytesReader` to cap request body, then count bytes via `io.Copy`.

```go
r := http.MaxBytesReader(w, req.Body, maxAllowedEncryptedBytes)
written, err := io.Copy(dst, r)
```

Backend should compare `written` against:

- configured max upload size
- remaining user quota
- declared encrypted size, if present

If browser underreports size, server must still stop at byte limit.

## S3 / R2 Multipart Flow

For S3-compatible multipart uploads, backend must not trust browser-declared size.

Flow:

1. Client declares intended plaintext/encrypted size.
2. Backend checks declared size against configured limits and reserves quota.
3. Backend creates `upload_session` in `pending`/active state.
4. Client uploads encrypted parts to presigned URLs.
5. Client requests completion.
6. Backend completes multipart upload.
7. Backend asks storage provider for final object size.
8. Backend uses provider-reported object size as authoritative.
9. Backend commits quota with actual stored size.
10. If actual stored size exceeds allowed size/quota, backend deletes object, releases reservation, marks upload failed.

Example with `HeadObject`:

```go
out, err := s3Client.HeadObject(ctx, &s3.HeadObjectInput{
    Bucket: aws.String(bucket),
    Key:    aws.String(objectKey),
})

actualEncryptedSize := *out.ContentLength
```

`actualEncryptedSize` is real stored object size.

## Quota Reservation Policy

Reserve first. Finalize with actual stored bytes later.

Example:

- User quota: `10 GB`
- User used: `8 GB`
- Client declares: `1 GB`
- Backend reserves: `1 GB`
- Upload completes
- Storage provider reports: `1.2 GB`

If `8 GB + 1.2 GB <= 10 GB`:

- commit actual usage `+1.2 GB`

Else:

- delete object
- release reservation
- fail upload

This blocks abuse where many uploads start before real usage is committed.

## What To Store

Recommended fields for `files` / `upload_sessions`:

- `declared_plaintext_size`
- `declared_encrypted_size`
- `actual_encrypted_size`
- `chunk_size`
- `total_parts`

Quota/storage enforcement must use:

- `actual_encrypted_size`

Do not rely on `declared_plaintext_size` for quota enforcement.

## Zero-Knowledge Note

This remains zero-knowledge.

Server only learns:

- encrypted object byte length

Server still does not learn:

- file contents
- file keys
- plaintext metadata unless explicitly exposed

Plaintext size is slightly more sensitive than encrypted size. Prefer using encrypted size for enforcement.

## Arkive Policy

### Core

- Quota uses actual encrypted bytes stored
- Max upload size uses actual encrypted bytes stored

### Cloud

- Quota uses actual encrypted bytes stored
- Billing uses actual encrypted bytes stored
- Fair-use bandwidth uses actual transferred bytes

## Implementation One-Liner

- Local storage: count bytes while writing.
- S3 / R2: `HeadObject` completed object and use `Content-Length`.
