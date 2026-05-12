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

function encodeBase64(bytes) {
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
    const crypto = await window.ArkiveCrypto.ready();
    const cipherStart = this.cipherOffsets[index];
    const cipherEnd = cipherStart + Number(chunk.cipher_size || 0);
    const encryptedBytes = await this.fetchEncryptedRange(cipherStart, cipherEnd);
    try {
      const expectedHash = String(chunk.hash || "");
      if (expectedHash) {
        const actualHash = crypto.hash_bytes_blake3(encryptedBytes);
        try {
          const hashEncoding = String((this.manifest && this.manifest.hash_encoding) || "base64").toLowerCase();
          const actual = hashEncoding === "hex"
            ? Array.from(actualHash).map(function(byte) { return byte.toString(16).padStart(2, "0"); }).join("")
            : encodeBase64(actualHash);
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
        new TextEncoder().encode(this.chunkAAD(index)),
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
    const size = Number((this.metadata && this.metadata.size) || this.record.plaintextSize || 0);
    const blobDownloadMaxBytes = 64 * 1024 * 1024;
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
    if (size > blobDownloadMaxBytes) {
      throw new Error("Browser download is disabled for large files without direct-save support.");
    }
    const blob = await this.createBlob();
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = filename;
    anchor.click();
    window.setTimeout(function() {
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
