import { fetchEncryptedChunk } from "./chunk_fetcher.js";
import { buildChunkMap } from "./chunk_map.js";
import { downloadFile } from "./download_controller.js";

function createContextId() {
  if (window.crypto && window.crypto.randomUUID) {
    return window.crypto.randomUUID();
  }
  return "file-reader-" + Date.now() + "-" + Math.random().toString(36).slice(2);
}

function bytesToText(bytes) {
  return new TextDecoder().decode(bytes);
}

export class ArkiveFileReader {
  constructor(options) {
    this.fileId = String((options && options.fileId) || "");
    this.apiBase = (options && options.apiBase) || "/api/files";
    this.contextId = createContextId();
    this.record = null;
    this.metadata = null;
    this.manifest = null;
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
    if (!this.fileId) {
      throw new Error("Missing file ID");
    }
    await window.ArkiveVault.waitUntilReady();
    const response = await fetch(
      this.apiBase + "/" + encodeURIComponent(this.fileId) + "/record",
      {
        method: "GET",
        headers: { "Content-Type": "application/json" },
      },
    );
    const data = await response.json();
    if (!response.ok) {
      throw new Error((data && data.error) || "Failed to load file");
    }
    this.record = data;
    const opened = await window.ArkiveVault.openFileContext(this.contextId, data);
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
    await window.ArkiveVault.closeFileContext(this.contextId).catch(function() {});
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
    const result = await window.ArkiveVault.decryptFileChunk(
      this.contextId,
      encryptedBytes,
      mappedChunk.aad || this.chunkAAD(mappedChunk.index),
      mappedChunk.hash || "",
    );
    return result.chunkBytes;
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
