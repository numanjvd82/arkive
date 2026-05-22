import { fetchEncryptedChunk } from "./reader/chunk_fetcher.js";
import { buildChunkMap } from "./reader/chunk_map.js";
import { downloadFile } from "./reader/download_controller.js";

function createContextId() {
  if (window.crypto && window.crypto.randomUUID) {
    return window.crypto.randomUUID();
  }
  return "share-reader-" + Date.now() + "-" + Math.random().toString(36).slice(2);
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
    this.chunkMap = [];
    this.loadPromise = null;
    this.chunkCache = new Map();
    this.chunkCacheBytes = 0;
    this.maxChunkCacheBytes = 32 * 1024 * 1024;
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
    if (!window.ArkiveVault || typeof window.ArkiveVault.openPublicShareContext !== "function") {
      throw new Error("Share crypto is unavailable");
    }

    const response = await fetch("/api/public/shares/" + encodeURIComponent(this.token), {
      method: "GET",
      headers: { "Content-Type": "application/json" },
    });
    const data = await window.ArkiveAPI.readJSON(response, "Failed to load share");

    this.record = data;
    const opened = await window.ArkiveVault.openPublicShareContext(
      this.contextId,
      data,
      this.shareSecret,
    );
    this.metadata = JSON.parse(opened.metadata || "{}");
    this.manifest = JSON.parse(opened.manifest || "{}");

    this.chunkMap = buildChunkMap(
      this.manifest,
      this.record.chunkSize,
      Number((this.metadata && this.metadata.size) || this.record.plaintextSize || 0),
    );
    return this;
  }

  async dispose() {
    this.clearChunkCache();
    if (!window.ArkiveVault || typeof window.ArkiveVault.closePublicShareContext !== "function") {
      return;
    }
    await window.ArkiveVault.closePublicShareContext(this.contextId).catch(function() {});
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
    const result = await window.ArkiveVault.decryptPublicShareChunk(
      this.contextId,
      encryptedBytes,
      mappedChunk.aad || this.chunkAAD(mappedChunk.index),
      mappedChunk.hash || "",
      String((this.manifest && this.manifest.hash_encoding) || "base64"),
    );
    return result.chunkBytes;
  }

  clearChunkCache() {
    this.chunkCache.clear();
    this.chunkCacheBytes = 0;
  }

  cachedChunk(index) {
    const entry = this.chunkCache.get(index);
    if (!entry) {
      return null;
    }
    this.chunkCache.delete(index);
    this.chunkCache.set(index, entry);
    return entry.bytes;
  }

  storeChunk(index, bytes) {
    if (!(bytes instanceof Uint8Array)) {
      return;
    }
    const existing = this.chunkCache.get(index);
    if (existing) {
      this.chunkCacheBytes -= existing.size;
      this.chunkCache.delete(index);
    }
    this.chunkCache.set(index, { bytes: bytes, size: bytes.length });
    this.chunkCacheBytes += bytes.length;

    while (this.chunkCacheBytes > this.maxChunkCacheBytes && this.chunkCache.size > 1) {
      const oldestKey = this.chunkCache.keys().next().value;
      const oldest = this.chunkCache.get(oldestKey);
      this.chunkCache.delete(oldestKey);
      this.chunkCacheBytes -= oldest ? oldest.size : 0;
    }
  }

  async readChunk(index) {
    await this.load();
    const cached = this.cachedChunk(index);
    if (cached) {
      return cached;
    }
    const chunk = this.chunkMap[index];
    if (!chunk) {
      throw new Error("Chunk out of range");
    }
    const encryptedBytes = await fetchEncryptedChunk(this.record.sourceUrl, chunk, null);
    const chunkBytes = await this.decryptChunk(chunk, encryptedBytes);
    this.storeChunk(index, chunkBytes);
    return chunkBytes;
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
      reader: this,
      record: record,
      chunkMap: this.chunkMap,
      decryptChunk: this.decryptChunk.bind(this),
      filename: record.filename,
      warningContainer: settings.warningContainer || null,
      onProgress: settings.onProgress,
      signal: settings.signal,
      readAhead: true,
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
