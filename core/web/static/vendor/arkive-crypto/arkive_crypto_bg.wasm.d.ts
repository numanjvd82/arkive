/* tslint:disable */
/* eslint-disable */
export const memory: WebAssembly.Memory;
export const format_recovery_key: (a: number, b: number) => [number, number, number, number];
export const generate_recovery_key: () => [number, number];
export const parse_recovery_key: (a: number, b: number) => [number, number, number, number];
export const recover_master_key: (a: number, b: number, c: number, d: number) => [number, number, number, number];
export const wrap_master_key_for_recovery: (a: number, b: number, c: number, d: number) => [number, number, number, number];
export const decrypt_chunk: (a: number, b: number, c: number, d: number, e: number, f: number) => [number, number, number, number];
export const encrypt_chunk: (a: number, b: number, c: number, d: number, e: number, f: number) => [number, number, number, number];
export const unwrap_file_key: (a: number, b: number, c: number, d: number, e: number, f: number) => [number, number, number, number];
export const unwrap_master_key: (a: number, b: number, c: number, d: number, e: number, f: number) => [number, number, number, number];
export const wrap_file_key: (a: number, b: number, c: number, d: number, e: number, f: number) => [number, number, number, number];
export const wrap_master_key: (a: number, b: number, c: number, d: number, e: number, f: number) => [number, number, number, number];
export const derive_password_kek: (a: number, b: number, c: number, d: number) => [number, number, number, number];
export const generate_file_key: () => [number, number];
export const generate_salt: () => [number, number];
export const zeroize: (a: number, b: number, c: any) => void;
export const generate_master_key: () => [number, number];
export const generate_share_key: () => [number, number];
export const __wbindgen_exn_store: (a: number) => void;
export const __externref_table_alloc: () => number;
export const __wbindgen_externrefs: WebAssembly.Table;
export const __wbindgen_malloc: (a: number, b: number) => number;
export const __externref_table_dealloc: (a: number) => void;
export const __wbindgen_free: (a: number, b: number, c: number) => void;
export const __wbindgen_realloc: (a: number, b: number, c: number, d: number) => number;
export const __wbindgen_start: () => void;
