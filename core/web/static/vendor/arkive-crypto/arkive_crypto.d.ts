/* tslint:disable */
/* eslint-disable */

export class Blake3Hasher {
    free(): void;
    [Symbol.dispose](): void;
    digest(): Uint8Array;
    digest_hex(): string;
    finalize(): Uint8Array;
    finalize_hex(): string;
    constructor();
    update(data: Uint8Array): void;
}

export class Sha256Hasher {
    free(): void;
    [Symbol.dispose](): void;
    digest(): Uint8Array;
    digest_hex(): string;
    finalize(): Uint8Array;
    finalize_hex(): string;
    constructor();
    update(data: Uint8Array): void;
}

export function decrypt_chunk(encrypted: Uint8Array, key: Uint8Array, aad: Uint8Array): Uint8Array;

export function decrypt_file_metadata(encrypted_metadata: Uint8Array, master_key: Uint8Array, aad: Uint8Array): Uint8Array;

export function derive_password_kek(password: string, salt: Uint8Array): Uint8Array;

export function derive_search_key(master_key: Uint8Array): Uint8Array;

export function encrypt_chunk(chunk: Uint8Array, key: Uint8Array, aad: Uint8Array): Uint8Array;

export function encrypt_file_metadata(metadata_json: Uint8Array, master_key: Uint8Array, aad: Uint8Array): Uint8Array;

export function format_recovery_key(recovery_key: Uint8Array): string;

export function generate_file_key(): Uint8Array;

export function generate_master_key(): Uint8Array;

export function generate_recovery_key(): Uint8Array;

export function generate_salt(): Uint8Array;

export function generate_share_key(): Uint8Array;

export function hash_blake3(data: Uint8Array): Uint8Array;

export function hash_bytes_blake3(data: Uint8Array): Uint8Array;

export function hash_bytes_blake3_hex(data: Uint8Array): string;

export function hash_bytes_sha256(data: Uint8Array): Uint8Array;

export function hash_bytes_sha256_hex(data: Uint8Array): string;

export function hmac_sha256(key: Uint8Array, data: Uint8Array): Uint8Array;

export function parse_recovery_key(recovery_key: string): Uint8Array;

export function recover_master_key(encrypted_master_key: Uint8Array, recovery_key: Uint8Array): Uint8Array;

export function unwrap_file_key(encrypted_file_key: Uint8Array, master_key: Uint8Array, aad: Uint8Array): Uint8Array;

export function unwrap_master_key(encrypted_master_key: Uint8Array, kek: Uint8Array, aad: Uint8Array): Uint8Array;

export function unwrap_master_key_with_recovery_key(encrypted_master_key: Uint8Array, recovery_key: Uint8Array, user_id: string): Uint8Array;

export function wrap_file_key(file_key: Uint8Array, master_key: Uint8Array, aad: Uint8Array): Uint8Array;

export function wrap_master_key(master_key: Uint8Array, kek: Uint8Array, aad: Uint8Array): Uint8Array;

export function wrap_master_key_for_recovery(master_key: Uint8Array, recovery_key: Uint8Array): Uint8Array;

export function wrap_master_key_with_password(master_key: Uint8Array, password: string, salt: Uint8Array, user_id: string): Uint8Array;

export function wrap_master_key_with_recovery_key(master_key: Uint8Array, recovery_key: Uint8Array, user_id: string): Uint8Array;

export function zeroize(bytes: Uint8Array): void;

export type InitInput = RequestInfo | URL | Response | BufferSource | WebAssembly.Module;

export interface InitOutput {
    readonly memory: WebAssembly.Memory;
    readonly __wbg_blake3hasher_free: (a: number, b: number) => void;
    readonly __wbg_sha256hasher_free: (a: number, b: number) => void;
    readonly blake3hasher_digest: (a: number) => [number, number, number, number];
    readonly blake3hasher_digest_hex: (a: number) => [number, number, number, number];
    readonly blake3hasher_finalize: (a: number) => [number, number, number, number];
    readonly blake3hasher_finalize_hex: (a: number) => [number, number, number, number];
    readonly blake3hasher_new: () => number;
    readonly blake3hasher_update: (a: number, b: number, c: number) => [number, number];
    readonly decrypt_chunk: (a: number, b: number, c: number, d: number, e: number, f: number) => [number, number, number, number];
    readonly decrypt_file_metadata: (a: number, b: number, c: number, d: number, e: number, f: number) => [number, number, number, number];
    readonly derive_password_kek: (a: number, b: number, c: number, d: number) => [number, number, number, number];
    readonly derive_search_key: (a: number, b: number) => [number, number, number, number];
    readonly encrypt_chunk: (a: number, b: number, c: number, d: number, e: number, f: number) => [number, number, number, number];
    readonly encrypt_file_metadata: (a: number, b: number, c: number, d: number, e: number, f: number) => [number, number, number, number];
    readonly format_recovery_key: (a: number, b: number) => [number, number, number, number];
    readonly generate_file_key: () => [number, number];
    readonly generate_master_key: () => [number, number];
    readonly generate_recovery_key: () => [number, number];
    readonly generate_salt: () => [number, number];
    readonly generate_share_key: () => [number, number];
    readonly hash_blake3: (a: number, b: number) => [number, number, number, number];
    readonly hash_bytes_blake3: (a: number, b: number) => [number, number];
    readonly hash_bytes_blake3_hex: (a: number, b: number) => [number, number];
    readonly hash_bytes_sha256: (a: number, b: number) => [number, number];
    readonly hash_bytes_sha256_hex: (a: number, b: number) => [number, number];
    readonly hmac_sha256: (a: number, b: number, c: number, d: number) => [number, number, number, number];
    readonly parse_recovery_key: (a: number, b: number) => [number, number, number, number];
    readonly recover_master_key: (a: number, b: number, c: number, d: number) => [number, number, number, number];
    readonly sha256hasher_digest: (a: number) => [number, number, number, number];
    readonly sha256hasher_digest_hex: (a: number) => [number, number, number, number];
    readonly sha256hasher_finalize: (a: number) => [number, number, number, number];
    readonly sha256hasher_finalize_hex: (a: number) => [number, number, number, number];
    readonly sha256hasher_new: () => number;
    readonly sha256hasher_update: (a: number, b: number, c: number) => [number, number];
    readonly unwrap_file_key: (a: number, b: number, c: number, d: number, e: number, f: number) => [number, number, number, number];
    readonly unwrap_master_key: (a: number, b: number, c: number, d: number, e: number, f: number) => [number, number, number, number];
    readonly unwrap_master_key_with_recovery_key: (a: number, b: number, c: number, d: number, e: number, f: number) => [number, number, number, number];
    readonly wrap_file_key: (a: number, b: number, c: number, d: number, e: number, f: number) => [number, number, number, number];
    readonly wrap_master_key: (a: number, b: number, c: number, d: number, e: number, f: number) => [number, number, number, number];
    readonly wrap_master_key_for_recovery: (a: number, b: number, c: number, d: number) => [number, number, number, number];
    readonly wrap_master_key_with_password: (a: number, b: number, c: number, d: number, e: number, f: number, g: number, h: number) => [number, number, number, number];
    readonly wrap_master_key_with_recovery_key: (a: number, b: number, c: number, d: number, e: number, f: number) => [number, number, number, number];
    readonly zeroize: (a: number, b: number, c: any) => void;
    readonly __wbindgen_exn_store: (a: number) => void;
    readonly __externref_table_alloc: () => number;
    readonly __wbindgen_externrefs: WebAssembly.Table;
    readonly __externref_table_dealloc: (a: number) => void;
    readonly __wbindgen_free: (a: number, b: number, c: number) => void;
    readonly __wbindgen_malloc: (a: number, b: number) => number;
    readonly __wbindgen_realloc: (a: number, b: number, c: number, d: number) => number;
    readonly __wbindgen_start: () => void;
}

export type SyncInitInput = BufferSource | WebAssembly.Module;

/**
 * Instantiates the given `module`, which can either be bytes or
 * a precompiled `WebAssembly.Module`.
 *
 * @param {{ module: SyncInitInput }} module - Passing `SyncInitInput` directly is deprecated.
 *
 * @returns {InitOutput}
 */
export function initSync(module: { module: SyncInitInput } | SyncInitInput): InitOutput;

/**
 * If `module_or_path` is {RequestInfo} or {URL}, makes a request and
 * for everything else, calls `WebAssembly.instantiate` directly.
 *
 * @param {{ module_or_path: InitInput | Promise<InitInput> }} module_or_path - Passing `InitInput` directly is deprecated.
 *
 * @returns {Promise<InitOutput>}
 */
export default function __wbg_init (module_or_path?: { module_or_path: InitInput | Promise<InitInput> } | InitInput | Promise<InitInput>): Promise<InitOutput>;
