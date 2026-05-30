# Arkive Crypto

This is the core crypto crate for Arkive. Every encryption, key derivation, hashing, and key-wrapping operation lives here. It compiles to both native Rust and WASM — the browser runtime calls into the exact same primitives.

## Why we built it this way

We started with a simple question: what happens when a user changes their password?

If you encrypt files directly with a password-derived key, the answer is "re-encrypt everything." That's slow, risky, and doesn't scale past a few hundred megabytes. We wanted password changes to be a single re-wrap operation — one AES/GCM envelope swap, not a bulk re-encryption of every chunk.

That single requirement pulled us toward a key hierarchy. Everything else — recovery keys, per-file keys, chunk-level AEAD, AAD binding — followed naturally once we had the hierarchy in place.

## Key hierarchy

We keep a strict separation between low-entropy human input and high-entropy cryptographic keys:

```
password ──▶ password_kek ──▶ encrypted_master_key ──▶ master_key
                                                                │
                                           ┌────────────────────┤
                                           ▼                    ▼
                                   encrypted_file_key     recovery_kek
                                           │                    │
                                           ▼                    ▼
                                       file_key        encrypted_master_key_recovery
                                           │
                                           ▼
                                   encrypted chunks
```

Every arrow that crosses a trust boundary is an AEAD operation with domain-separated AAD. Nothing crosses without authentication.

### The keys, concretely

| Key | Size | Source | Lifetime |
|-----|------|--------|----------|
| `master_key` | 32 bytes | `generate_master_key()` | Per vault |
| `password_kek` | 32 bytes | `Argon2id(password, salt)` | Per session |
| `recovery_key` | 32 bytes | `generate_recovery_key()` | Per user, exportable |
| `recovery_kek` | 32 bytes | `HKDF-SHA256(recovery_key)` | Per recovery operation |
| `file_key` | 32 bytes | `generate_file_key()` | Per file |
| `share_key` | 32 bytes | `generate_share_key()` | Per share link |
| `salt` | 16 bytes | `generate_salt()` | Per user |

All random keys come from the OS CSPRNG (`rand::thread_rng`). We zeroize sensitive buffers on drop.

### What we get from this

- **Password changes are cheap.** Rotate the `password_kek` and re-wrap the `encrypted_master_key` blob. File data never moves.
- **Recovery is independent.** The `recovery_key` wraps `master_key` through a separate path. No password needed for recovery.
- **Per-file blast radius.** Compromise of one `file_key` exposes one file, not the vault.
- **Sharing is future-proof.** `share_key` gives us a dedicated key slot for multi-recipient and link-based sharing without touching file keys or master keys.

## Encryption format

Every encrypted object in Arkive uses the same binary envelope:

```
[version:1 byte][nonce:24 bytes][ciphertext || poly1305 tag]
```

Version `0x01` = XChaCha20-Poly1305.

We picked XChaCha20-Poly1305 over AES-GCM for three reasons:
1. **Large nonce space.** 192-bit random nonces mean we never worry about nonce reuse at scale. We don't need counters or state.
2. **Software performance.** It runs fast without AES-NI, which matters in browsers without hardware acceleration.
3. **Simplicity.** No padding, no block alignment, no special cases for empty input.

The version byte gives us an explicit migration path. If we ever switch to AEGIS or a post-quantum AEAD, version `0x02` slots in without guessing.

### What uses this envelope

Every function in `symmetric.rs` writes and reads this format:
- `encrypt_chunk` / `decrypt_chunk`
- `wrap_master_key` / `unwrap_master_key`
- `wrap_file_key` / `unwrap_file_key`
- `encrypt_file_metadata` / `decrypt_file_metadata`

## AAD: why we bind context everywhere

AEAD authenticates ciphertext, but it doesn't tell you *where* that ciphertext belongs. If we used the same key for chunk 7 and chunk 42 without context separation, an attacker could replay chunk 7's ciphertext in chunk 42's slot and decryption would succeed.

AAD prevents that. Every encrypt/decrypt call takes an `aad` parameter:

```rust
encrypt_chunk(&chunk, &file_key, b"file:abc123|chunk:7")?;
```

We recommend binding:
- File identity (`"file:<id>"`)
- Chunk index (`"chunk:<n>"`)  
- Operation type (`"wrapped-file-key"`, `"file-metadata"`)
- Future: tenant, vault, and schema version identifiers

AAD is not secret and not stored in the ciphertext. It's presented again at decryption time. If it doesn't match, decryption fails cleanly — no silent corruption, no partial plaintext.

## Hashing

We expose both Blake3 and SHA-256. Blake3 is our primary hash — it's fast, parallel, and the tree-hash mode is future-proof for large files. SHA-256 is for interop (recovery key checksums, HKDF).

### One-shot

```rust
hash_bytes_blake3(data)       // → [u8; 32]
hash_bytes_blake3_hex(data)   // → String
hash_bytes_sha256(data)       // → [u8; 32]
hash_bytes_sha256_hex(data)   // → String
```

### Streaming

For chunk-by-chunk hashing of large files:

```rust
let mut hasher = Blake3Hasher::new();
hasher.update(&chunk_1)?;
hasher.update(&chunk_2)?;
let hash = hasher.finalize()?;
```

`digest()` and `digestHex()` give you intermediate snapshots without closing the hasher. `finalize()` and `finalizeHex()` close it. Calling any method after finalize returns an error — no panics, no browser traps.

## Metadata encryption

File metadata (filenames, sizes, MIME types, custom tags) is JSON we encrypt with the master key:

```rust
encrypt_file_metadata(&metadata_json, &master_key, b"file:abc123")?;
```

Same envelope format as chunks. Same AAD binding. If you encrypt metadata for file X but try to decrypt it with file Y's AAD, it fails — even if the master key matches.

## Recovery keys

We wanted recovery to be user-portable without tying it to the password. So we generate a separate 32-byte `recovery_key`, wrap the `master_key` with it, and give the user a human-readable encoding:

```
ARK-RK1-XXXXXXXX-XXXXXXXX-XXXXXXXX-XXXXXXXX-XXXXXXXX-XXXXXXXX-XXXXXXXX-XXXXXXXX-XXXXXXXX
```

The binary payload under the hood:
```
[version:1 byte][recovery_key:32 bytes][checksum:3 bytes]
```

Checksum is `SHA256("arkive-recovery-key-v1" || version || key)[..3]`. It catches most typos before we attempt a recovery operation.

The `recovery_key` derives a `recovery_kek` via `HKDF-SHA256` with domain-separated info. That KEK wraps the master key. The recovery key itself is never used as a direct encryption key.

A few things the recovery key is not:
- It's not the master key itself — it wraps the master key, so it can be rotated independently.
- It's not a password reset — you need both the recovery key and the stored recovery-wrapped master-key blob.
- It's not optional — if you lose the recovery key and forget your password, the vault is gone.

## Error model

Every WASM export returns `Result<_, JsValue>`. We never `unwrap()`, `assert!()`, or `panic!()` across the WASM boundary. The browser gets inspectable error strings for:

- Wrong key length
- Wrong salt length  
- Empty or truncated blobs
- Unrecognized envelope version
- AEAD authentication failure (wrong key, wrong AAD, tampered data)
- Recovery key checksum mismatch
- Double-finalize on hashers

This matters because a panic in WASM is an opaque trap — the JS side can't catch it, can't inspect it, and can't recover. Explicit errors let the UI show meaningful messages and retry paths.

## Crate structure

```
src/
├── lib.rs          # Module registration and re-exports
├── errors.rs       # CryptoError enum with Display/Into<JsValue>
├── utils.rs        # Random generation (keys, salts) and zeroize
├── kdf.rs          # Argon2id password-to-KEK derivation
├── symmetric.rs    # XChaCha20-Poly1305 encrypt/decrypt, key wrapping
├── recovery.rs     # Recovery key encode/decode, master key recovery
├── hash.rs         # Blake3 and SHA-256 (one-shot + streaming)
└── metadata.rs     # Encrypted file metadata (JSON under master key)
```

## Testing

We test at the Rust level with `cargo test --lib`. Every module has its own `#[cfg(test)] mod tests`. Currently 66 tests covering:

- Roundtrip encrypt/decrypt for every workflow
- Wrong key, wrong AAD, and tampered ciphertext rejection
- Key length and salt length validation
- Version byte and blob format validation
- Recovery key encode/decode and checksum detection
- Argon2id determinism and salt isolation
- Hasher state machine (digest, finalize, double-finalize edge cases)
- Zeroize correctness

No WASM-specific tests yet those live in the JS integration layer. But the core logic is fully exercised natively.
