# Arkive Crypto Design

## Goals

Arkive uses a simple key hierarchy with explicit trust boundaries:

- A random 32-byte `master_key` is the root secret for a vault.
- User passwords are never used directly for encryption. They are only used to derive a 32-byte `password_kek` with Argon2id.
- A separate random 32-byte `recovery_key` is generated for recovery.
- Each file gets its own random 32-byte `file_key`.
- Large files are encrypted chunk-by-chunk with AEAD so we can stream, resume, and verify each chunk independently.

This design separates low-entropy user input from high-entropy encryption keys. That gives us better offline attack resistance, simpler rotation paths, and a cleaner recovery story than using the password directly as the file-encryption key.

## Architecture

### Key hierarchy

1. Generate `master_key` randomly.
2. Generate a per-user 16-byte salt.
3. Derive `password_kek = Argon2id(password, salt)`.
4. `wrap_master_key(master_key, password_kek, aad)`.
5. Generate a separate random `recovery_key`.
6. Derive a `recovery_kek` from `recovery_key`.
7. Wrap `master_key` with `recovery_kek` for recovery.
8. Generate a random `file_key` for each file.
9. Wrap `file_key` with `master_key`, `share_key`, or a recipient key.
10. Encrypt file chunks with `file_key`.

This structure gives us:

- Password changes without re-encrypting file data.
- Recovery can be rotated independently of the master key.
- Per-file blast-radius reduction.
- Clear room for future sharing and re-wrapping flows.
- A single symmetric envelope format across wrapped keys and encrypted content.

### Encryption format

Arkive currently uses XChaCha20-Poly1305 with a versioned binary envelope:

`[version:1][nonce:24][ciphertext || tag]`

Version `0x01` means:

- Algorithm: XChaCha20-Poly1305
- Nonce: random 24 bytes
- Tag: 16 bytes, appended by the AEAD

Why this format:

- XChaCha20 gives a large nonce space, which is safer for random nonce generation at scale.
- The version byte lets us migrate algorithms later without heuristic parsing.
- The same envelope works for bytes, chunks, wrapped master keys, and wrapped file keys.

### AAD support

AEAD is strongest when we bind context, not just ciphertext. Arkive exports only the architecture-specific AAD-bound workflows: password KEK to master key, master key to file key, and file key to chunks.

AAD is not encrypted, but it is authenticated. If it changes, decryption fails.

Recommended AAD uses:

- File identity
- Chunk index
- Chunk size
- Object type like `wrapped-file-key`
- Future tenant, vault, or schema identifiers

Why this is better:

- It prevents valid ciphertext from being replayed in the wrong context.
- It lets us bind metadata without copying it into plaintext.
- It keeps the encryption envelope stable while strengthening integrity.

## Validation and Error Model

The WASM boundary now returns explicit errors instead of panicking.

We validate:

- Symmetric keys must be 32 bytes.
- Salts must be 16 bytes.
- Encrypted blobs must be at least 41 bytes.
- Envelope version must be recognized.
- Recovery-key payloads must match version and checksum.

Why this matters:

- `unwrap()` and `assert!()` inside WASM exports turn ordinary bad input into traps. That is the wrong failure mode for untrusted browser input.
- Explicit errors make JS integration predictable.
- Length checks stop malformed inputs before they reach crypto primitives.

## Recovery Key Format

Arkive recovery keys are a human-portable encoding of a random 32-byte `recovery_key`, not of the `master_key`.

Text format:

`ARK-RK1-XXXXXXXX-XXXXXXXX-XXXXXXXX-XXXXXXXX-XXXXXXXX-XXXXXXXX-XXXXXXXX-XXXXXXXX-XXXXXXXX`

Details:

- `ARK-RK1` is the textual prefix and schema version.
- The payload is uppercase hex.
- Binary payload layout is:
  - `version:1`
  - `recovery_key:32`
  - `checksum:3`
- The checksum is the first 3 bytes of:
  - `SHA256("arkive-recovery-key-v1" || version || recovery_key)`

Recovery flow:

- `recovery_key` is random and user-portable.
- `recovery_kek = HKDF-SHA256(recovery_key, "arkive-recovery-kek-v1")`
- `wrap_master_key_for_recovery(master_key, recovery_key)`

Why this format:

- It is stable and easy to transcribe.
- It is case-insensitive after normalization.
- It has an explicit version.
- It detects most entry errors before we accept a recovery secret.

Important note:

- The recovery key is not a password reset token.
- It is not the master key itself.
- Anyone with the recovery key plus the stored recovery-wrapped master-key blob can recover the vault.

## Why This Design Is Better

Compared with a direct `password -> encrypt everything` model, this architecture is better because:

- Password changes are cheap.
- File encryption keys stay high-entropy and random.
- Metadata can be integrity-bound with AAD.
- The master key stays internal while recovery remains user-portable.
- The root secret can be recovered independently of the password.
- The envelope is versioned and migration-friendly.
- The browser side gets explicit, inspectable failures instead of traps.

## Recommended Hierarchy

- `password -> kek -> encrypted_master_key -> master_key`
- `master_key -> encrypted_file_key -> file_key`
- `file_key -> encrypted chunks`
- `recovery_key -> recovery_kek -> encrypted_master_key_recovery -> master_key`
- `share_key -> encrypted_file_key_for_share -> file_key`

The result is a design that is easier to reason about operationally and safer to evolve over time.
