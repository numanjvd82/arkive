import { fetchEncryptedChunk } from "./reader/chunk_fetcher.js";
import { buildChunkMap } from "./reader/chunk_map.js";
import { downloadFile } from "./reader/download_controller.js";

function createContextId() {
  if (window.crypto && window.crypto.randomUUID) {
    return window.crypto.randomUUID();
  }
  return "share-reader-" + Date.now() + "-" + Math.random().toString(36).slice(2);
}

function decodeBase64(value) {
  const normalized = String(value || "").trim();
  if (!normalized) {
    return new Uint8Array();
  }
  const binary = atob(normalized);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}

function bytesToBase64(bytes) {
  let binary = "";
  for (let i = 0; i < bytes.length; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}

function bytesToText(bytes) {
  return new TextDecoder().decode(bytes);
}

export class ArkiveShareReader {
  constructor(options) {
    this.token = String((options && options.token) || "");
    this.shareSecret = String((options && options.shareSecret) || "");
    this.contextId = createContextId();
    this.record = null;
    this.metadata = null;
    this.manifest = null;
    this.fileKey = null;
    this.chunkMap = [];
    this.loadPromise = null;
  }

  async load() {
    if (this.loadPromise) {
      return this.loadPromise;
    }
    this.loadPromise = this.loadInternal();
    return this.loadPromise;
  }

  async loadInternal() {
    if (!this.token) {
      throw new Error("Missing share token");
    }
    if (!this.shareSecret) {
      throw new Error("Missing share secret");
    }
    if (!window.ArkiveCrypto || typeof window.ArkiveCrypto.ready !== "function") {
      throw new Error("Crypto is unavailable");
    }

    const crypto = await window.ArkiveCrypto.ready();
    const response = await fetch("/api/public/shares/" + encodeURIComponent(this.token), {
      method: "GET",
      headers: { "Content-Type": "application/json" },
    });
    const data = await response.json();
    if (!response.ok) {
      throw new Error((data && data.error) || "Failed to load share");
    }

    this.record = data;
    const shareKey = decodeBase64(this.shareSecret);
    const encryptedFileKeyForShare = decodeBase64(data.encryptedFileKeyForShare);
    const encryptedMetadata = decodeBase64(data.encryptedMetadata);
    const encryptedManifest = decodeBase64(data.encryptedManifest);

    try {
      const fileKey = crypto.unwrap_file_key(
        encryptedFileKeyForShare,
        shareKey,
        new TextEncoder().encode(String(data.shareFileKeyAad || "")),
      );
      this.fileKey = fileKey.slice();

      const metadataBytes = crypto.decrypt_chunk(
        encryptedMetadata,
        fileKey,
        new TextEncoder().encode(String(data.metadataAad || "")),
      );
      const manifestBytes = crypto.decrypt_chunk(
        encryptedManifest,
        fileKey,
        new TextEncoder().encode(String(data.manifestAad || "")),
      );

      try {
        this.metadata = JSON.parse(bytesToText(metadataBytes) || "{}");
        this.manifest = JSON.parse(bytesToText(manifestBytes) || "{}");
      } finally {
        crypto.zeroize(metadataBytes);
        crypto.zeroize(manifestBytes);
      }
    } finally {
      crypto.zeroize(shareKey);
      crypto.zeroize(encryptedFileKeyForShare);
      crypto.zeroize(encryptedMetadata);
      crypto.zeroize(encryptedManifest);
    }

    this.chunkMap = buildChunkMap(
      this.manifest,
      this.record.chunkSize,
      Number((this.metadata && this.metadata.size) || this.record.plaintextSize || 0),
    );
    return this;
  }

  async dispose() {
    if (!this.fileKey || !window.ArkiveCrypto || typeof window.ArkiveCrypto.ready !== "function") {
      return;
    }
    const fileKey = this.fileKey;
    this.fileKey = null;
    const crypto = await window.ArkiveCrypto.ready();
    crypto.zeroize(fileKey);
  }

  getMetadata() {
    return this.metadata || {};
  }

  getManifest() {
    return this.manifest || {};
  }

  getChunkMap() {
    return this.chunkMap.slice();
  }

  chunkAAD(index) {
    return (
      "arkive:file-chunk:v1:" +
      String(this.record.vaultId || "") +
      ":" +
      String(this.record.fileId || "") +
      ":" +
      String(index + 1) +
      ":" +
      String(this.record.chunkSize || 0) +
      ":" +
      String(this.record.totalChunks || 0)
    );
  }

  async decryptChunk(mappedChunk, encryptedBytes) {
    const crypto = await window.ArkiveCrypto.ready();
    try {
      const expectedHash = String(mappedChunk.hash || "");
      if (expectedHash) {
        const actualHash = crypto.hash_bytes_blake3(encryptedBytes);
        try {
          const hashEncoding = String((this.manifest && this.manifest.hash_encoding) || "base64").toLowerCase();
          const actual = hashEncoding === "hex"
            ? Array.from(actualHash).map(function(byte) { return byte.toString(16).padStart(2, "0"); }).join("")
            : bytesToBase64(actualHash);
          if (actual !== expectedHash) {
            throw new Error("Encrypted chunk hash mismatch");
          }
        } finally {
          crypto.zeroize(actualHash);
        }
      }

      const chunkBytes = crypto.decrypt_chunk(
        encryptedBytes,
        this.fileKey,
        new TextEncoder().encode(mappedChunk.aad || this.chunkAAD(mappedChunk.index)),
      );
      try {
        return chunkBytes.slice();
      } finally {
        crypto.zeroize(chunkBytes);
      }
    } finally {
      crypto.zeroize(encryptedBytes);
    }
  }

  async readChunk(index) {
    await this.load();
    const chunk = this.chunkMap[index];
    if (!chunk) {
      throw new Error("Chunk out of range");
    }
    const encryptedBytes = await fetchEncryptedChunk(this.record.sourceUrl, chunk, null);
    return this.decryptChunk(chunk, encryptedBytes);
  }

  async readRange(start, end) {
    await this.load();
    const chunkSize = Number(this.record.chunkSize || 0);
    if (chunkSize <= 0) {
      throw new Error("Invalid chunk size");
    }
    const firstChunk = Math.floor(start / chunkSize);
    const lastChunk = Math.floor((end - 1) / chunkSize);
    const parts = [];
    let total = 0;
    for (let index = firstChunk; index <= lastChunk; index++) {
      const chunkBytes = await this.readChunk(index);
      let sliceStart = 0;
      let sliceEnd = chunkBytes.length;
      if (index === firstChunk) {
        sliceStart = start - index * chunkSize;
      }
      if (index === lastChunk) {
        sliceEnd = end - index * chunkSize;
      }
      const slice = chunkBytes.slice(sliceStart, sliceEnd);
      parts.push(slice);
      total += slice.length;
    }
    const output = new Uint8Array(total);
    let offset = 0;
    for (let i = 0; i < parts.length; i++) {
      output.set(parts[i], offset);
      offset += parts[i].length;
    }
    return output;
  }

  async createBlob() {
    await this.load();
    const parts = [];
    for (let i = 0; i < this.chunkMap.length; i++) {
      parts.push(await this.readChunk(i));
    }
    return new Blob(parts, {
      type: (this.metadata && this.metadata.mime) || "application/octet-stream",
    });
  }

  async download(options) {
    const settings = options || {};
    if (!this.record || !this.metadata || !this.manifest || !this.chunkMap.length) {
      await this.load();
    }
    const metadata = this.metadata || {};
    const record = Object.assign({}, this.record || {}, {
      filename: metadata.name || "download.bin",
      mimeType: metadata.mime || "application/octet-stream",
      plaintextSize: Number(metadata.size || (this.record && this.record.plaintextSize) || 0),
    });

    return downloadFile({
      record: record,
      chunkMap: this.chunkMap,
      decryptChunk: this.decryptChunk.bind(this),
      filename: record.filename,
      warningContainer: settings.warningContainer || null,
      onProgress: settings.onProgress,
      signal: settings.signal,
    });
  }

  async textPreview(maxBytes) {
    await this.load();
    const limit = Math.min(
      Number(maxBytes || 2 * 1024 * 1024),
      Number((this.metadata && this.metadata.size) || this.record.plaintextSize || 0),
    );
    const bytes = await this.readRange(0, limit);
    return bytesToText(bytes);
  }
}
