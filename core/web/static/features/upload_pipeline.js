const DEFAULT_PART_SIZE = 32 * 1024 * 1024;
const DEFAULT_CONCURRENCY = 4;
const DEFAULT_RETRIES = 3;

// TODOS:
// Change upload defaults to 4MB parts and 4 concurrency.
// Remove base64 from encrypted chunk flow and use Uint8Array/ArrayBuffer directly.
// Delete base64ToBytes() after switching to binary uploads.
// Null out large buffers (chunkBytes, encryptedBytes) after use.
// Add exponential retry backoff in uploadPartWithRetry().
// Cancel backend multipart session on abort/failure with cancelSession().
// Strengthen chunk AAD with file ID, session ID, part number, part size, and total parts.
// Keep encrypted chunk BLAKE3 hashes server-side for integrity/resume validation.
// Store plaintext BLAKE3 file hash only inside encrypted metadata for future verification/sync features.

function createUploadToken() {
  if (window.crypto && window.crypto.randomUUID) {
    return window.crypto.randomUUID();
  }
  return "upload-" + Date.now() + "-" + Math.random().toString(36).slice(2);
}

export class MultipartUploadPipeline {
  constructor(options) {
    this.apiBase = options.apiBase || "/api/uploads";
    this.partSize = options.partSize || DEFAULT_PART_SIZE;
    this.concurrency = options.concurrency || DEFAULT_CONCURRENCY;
    this.retries = options.retries || DEFAULT_RETRIES;
    this.onProgress = options.onProgress || function () {};
  }

  async upload(task) {
    const file = task.file;
    const uploadToken = createUploadToken();
    const partSize = this.partSize;
    const totalParts = Math.ceil(file.size / partSize);
    const metadata = {
      filename: file.name,
      mimeType: file.type || "application/octet-stream",
      size: file.size,
      lastModified: file.lastModified || 0,
    };

    if (
      !window.ArkiveVault ||
      !window.ArkiveVault.isUnlocked ||
      !window.ArkiveVault.isUnlocked()
    ) {
      throw new Error("Vault is locked");
    }

    task.uploadToken = uploadToken;
    task.abortController = new AbortController();
    task.totalBytes = file.size;
    task.uploadedBytes = 0;

    const prepared = await window.ArkiveVault.prepareUpload(
      uploadToken,
      metadata,
      totalParts,
    );
    const started = await this.startSession({
      encryptedMetadata: prepared.encryptedMetadata,
      encryptedFileKey: prepared.encryptedFileKey,
      originalSize: file.size,
      partSize: partSize,
      totalParts: totalParts,
      encryptionVersion: prepared.encryptionVersion || 1,
    });

    task.fileId = started.fileId;
    task.uploadSessionId = started.uploadSessionId;

    try {
      await this.uploadParts(task, partSize, totalParts);
      await this.completeSession(task.uploadSessionId);
    } finally {
      await window.ArkiveVault.finalizeUpload(uploadToken).catch(
        function () {},
      );
    }
  }

  abort(task) {
    if (task && task.abortController) {
      task.abortController.abort();
    }
  }

  async startSession(body) {
    return this.api("/start", {
      method: "POST",
      body: body,
    });
  }

  async presignPart(uploadSessionId, partNumber) {
    return this.api(
      "/" +
        encodeURIComponent(uploadSessionId) +
        "/parts/" +
        encodeURIComponent(String(partNumber)) +
        "/presign",
      {
        method: "POST",
      },
    );
  }

  async recordPart(uploadSessionId, body) {
    await this.api("/" + encodeURIComponent(uploadSessionId) + "/parts", {
      method: "POST",
      body: body,
    });
  }

  async completeSession(uploadSessionId) {
    await this.api("/" + encodeURIComponent(uploadSessionId) + "/complete", {
      method: "POST",
    });
  }

  async cancelSession(uploadSessionId) {
    await this.api("/" + encodeURIComponent(uploadSessionId) + "/cancel", {
      method: "POST",
    });
  }

  async uploadParts(task, partSize, totalParts) {
    let nextPart = 1;
    let active = 0;
    let finished = 0;
    let failed = null;

    return new Promise((resolve, reject) => {
      const launch = () => {
        if (failed) {
          reject(failed);
          return;
        }
        if (finished === totalParts && active === 0) {
          resolve();
          return;
        }
        while (active < this.concurrency && nextPart <= totalParts) {
          const partNumber = nextPart++;
          active++;
          this.uploadPartWithRetry(task, partNumber, partSize)
            .then((plaintextBytes) => {
              active--;
              finished++;
              task.uploadedBytes += plaintextBytes;
              this.onProgress(task, task.uploadedBytes, task.totalBytes);
              launch();
            })
            .catch((error) => {
              active--;
              failed = error;
              this.abort(task);
              launch();
            });
        }
      };

      launch();
    });
  }

  async uploadPartWithRetry(task, partNumber, partSize) {
    let lastError = null;
    for (let attempt = 1; attempt <= this.retries; attempt++) {
      try {
        return await this.uploadPart(task, partNumber, partSize);
      } catch (error) {
        lastError = error;
        if (task.abortController && task.abortController.signal.aborted) {
          break;
        }
      }
    }
    throw lastError || new Error("Part upload failed");
  }

  async uploadPart(task, partNumber, partSize) {
    const start = (partNumber - 1) * partSize;
    const end = Math.min(start + partSize, task.file.size);
    const chunkBytes = new Uint8Array(
      await task.file.slice(start, end).arrayBuffer(),
    );
    const aad =
      "arkive:file-part:v1:file:" + task.fileId + ":part:" + partNumber;
    const encrypted = await window.ArkiveVault.encryptUploadPart(
      task.uploadToken,
      chunkBytes,
      aad,
    );
    const presigned = await this.presignPart(task.uploadSessionId, partNumber);
    const response = await fetch(presigned.url, {
      method: "PUT",
      body: this.base64ToBytes(encrypted.encryptedChunk),
      signal: task.abortController.signal,
    });
    if (!response.ok) {
      throw new Error("Part upload failed");
    }

    const etag =
      response.headers.get("etag") || response.headers.get("ETag") || "";
    await this.recordPart(task.uploadSessionId, {
      partNumber: partNumber,
      encryptedSize: encrypted.encryptedSize,
      encryptedHash: encrypted.encryptedHash,
      etag: etag,
    });
    return end - start;
  }

  async api(path, options) {
    const response = await fetch(this.apiBase + path, {
      method: options.method || "POST",
      headers: options.body
        ? { "Content-Type": "application/json" }
        : undefined,
      body: options.body ? JSON.stringify(options.body) : undefined,
    });
    if (response.status === 204) {
      return null;
    }
    const data = await response.json();
    if (!response.ok) {
      const error = new Error((data && data.error) || "Request failed");
      error.status = response.status;
      error.data = data;
      throw error;
    }
    return data;
  }

  base64ToBytes(value) {
    const binary = atob(String(value || ""));
    const bytes = new Uint8Array(binary.length);
    for (let i = 0; i < binary.length; i++) {
      bytes[i] = binary.charCodeAt(i);
    }
    return bytes;
  }
}
