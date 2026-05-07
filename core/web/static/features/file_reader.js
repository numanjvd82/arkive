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
    this.cipherOffsets = [];
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
    this.cipherOffsets = [];
    let offset = 0;
    const chunks = this.manifest.chunks || [];
    for (let i = 0; i < chunks.length; i++) {
      this.cipherOffsets[i] = offset;
      offset += Number(chunks[i].cipher_size || 0);
    }
    return this;
  }

  async dispose() {
    await window.ArkiveVault.closeFileContext(this.contextId).catch(function () {});
  }

  getMetadata() {
    return this.metadata || {};
  }

  getManifest() {
    return this.manifest || {};
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

  async fetchEncryptedRange(start, endExclusive) {
    const response = await fetch(this.record.sourceUrl, {
      method: "GET",
      headers: {
        Range: "bytes=" + String(start) + "-" + String(endExclusive - 1),
      },
    });
    if (!response.ok && response.status !== 206) {
      throw new Error("Failed to fetch encrypted chunk");
    }
    return new Uint8Array(await response.arrayBuffer());
  }

  async readChunk(index) {
    await this.load();
    const chunk = (this.manifest.chunks || [])[index];
    if (!chunk) {
      throw new Error("Chunk out of range");
    }
    const cipherStart = this.cipherOffsets[index];
    const cipherEnd = cipherStart + Number(chunk.cipher_size || 0);
    const encryptedBytes = await this.fetchEncryptedRange(cipherStart, cipherEnd);
    return window.ArkiveVault.decryptFileChunk(
      this.contextId,
      encryptedBytes,
      this.chunkAAD(index),
      chunk.hash || "",
    ).then(function (result) {
      return result.chunkBytes;
    });
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
    const chunks = this.manifest.chunks || [];
    const parts = [];
    for (let i = 0; i < chunks.length; i++) {
      parts.push(await this.readChunk(i));
    }
    return new Blob(parts, {
      type: (this.metadata && this.metadata.mime) || "application/octet-stream",
    });
  }

  async download() {
    await this.load();
    const filename = (this.metadata && this.metadata.name) || "download.bin";
    if (window.showSaveFilePicker) {
      const handle = await window.showSaveFilePicker({
        suggestedName: filename,
      });
      const writable = await handle.createWritable();
      try {
        const chunks = this.manifest.chunks || [];
        for (let i = 0; i < chunks.length; i++) {
          await writable.write(await this.readChunk(i));
        }
      } finally {
        await writable.close();
      }
      return;
    }
    const blob = await this.createBlob();
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = filename;
    anchor.click();
    window.setTimeout(function () {
      URL.revokeObjectURL(url);
    }, 1000);
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
