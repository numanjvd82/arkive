const DEFAULT_PART_SIZE = 4 * 1024 * 1024;
const DEFAULT_CONCURRENCY = 4;
const DEFAULT_RETRIES = 3;

// TODOS:
// Store plaintext BLAKE3 file hash only inside encrypted metadata for future verification/sync features.

function createUploadToken() {
  if (window.crypto && window.crypto.randomUUID) {
    return window.crypto.randomUUID();
  }
  return "upload-" + Date.now() + "-" + Math.random().toString(36).slice(2);
}

function fileExtension(name) {
  const value = String(name || "");
  const dot = value.lastIndexOf(".");
  if (dot <= 0 || dot === value.length - 1) {
    return "";
  }
  return value.slice(dot + 1).toLowerCase();
}

export class MultipartUploadPipeline {
  constructor(options) {
    this.apiBase = options.apiBase || "/api/uploads";
    this.partSize = options.partSize || DEFAULT_PART_SIZE;
    this.concurrency = options.concurrency || DEFAULT_CONCURRENCY;
    this.retries = options.retries || DEFAULT_RETRIES;
    this.retryBaseDelayMs = options.retryBaseDelayMs || 400;
    this.onProgress = options.onProgress || function () {};
  }

  async upload(task) {
    if (
      window.ArkiveVault &&
      typeof window.ArkiveVault.waitUntilReady === "function"
    ) {
      await window.ArkiveVault.waitUntilReady();
    }

    const file = task.file;
    const uploadToken = createUploadToken();
    const requestedPartSize = task.chunkSize || this.partSize;
    const requestedTotalParts = Math.ceil(file.size / requestedPartSize);
    const metadata = {
      schema: "arkive.file.metadata",
      version: 1,
      name: file.name,
      mime: file.type || "application/octet-stream",
      extension: fileExtension(file.name),
      size: file.size,
      created_at_client: new Date().toISOString(),
      modified_at_client: file.lastModified
        ? new Date(file.lastModified).toISOString()
        : null,
      preview: {
        thumbnail_file_id: null,
        has_thumbnail: false,
      },
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

    const started = await this.startSession({
      originalSize: file.size,
      partSize: requestedPartSize,
      totalParts: requestedTotalParts,
      encryptionVersion: 1,
    });

    task.fileId = started.fileId;
    task.vaultId = started.vaultId;
    task.uploadSessionId = started.uploadSessionId;
    task.chunkSize = started.partSize || requestedPartSize;
    task.totalParts = started.totalParts || requestedTotalParts;
    task.partRecords = [];

    const prepared = await window.ArkiveVault.prepareUpload(
      uploadToken,
      task.vaultId,
      task.fileId,
      metadata,
      task.totalParts,
    );

    let completed = false;

    try {
      await this.uploadParts(task, task.chunkSize, task.totalParts);
      const finalized = await window.ArkiveVault.finalizeUpload(
        uploadToken,
        task.vaultId,
        task.fileId,
        this.buildManifest(task, metadata),
        task.partRecords.map(function (part) {
          return part.encryptedHash;
        }),
      );
      await this.completeSession(task.uploadSessionId, {
        encryptedMetadata: prepared.encryptedMetadata,
        encryptedFileKey: prepared.encryptedFileKey,
        encryptedManifest: finalized.encryptedManifest,
        encryptedHash: finalized.encryptedHash,
      });
      completed = true;
    } catch (error) {
      if (task.uploadSessionId) {
        await this.cancelSession(task.uploadSessionId).catch(function () {});
      }
      throw error;
    } finally {
      await window.ArkiveVault.clearUploadContext(uploadToken).catch(
        function () {},
      );
      if (!completed) {
        task.uploadedBytes = 0;
      }
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

  async completeSession(uploadSessionId, body) {
    await this.api("/" + encodeURIComponent(uploadSessionId) + "/complete", {
      method: "POST",
      body: body,
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
        if (attempt < this.retries) {
          await this.waitForRetry(attempt, task.abortController.signal);
        }
      }
    }
    throw lastError || new Error("Part upload failed");
  }

  async uploadPart(task, partNumber, partSize) {
    const start = (partNumber - 1) * partSize;
    const end = Math.min(start + partSize, task.file.size);
    let chunkBytes = null;
    let encryptedBytes = null;

    try {
      chunkBytes = new Uint8Array(await task.file.slice(start, end).arrayBuffer());
      const encrypted = await window.ArkiveVault.encryptUploadPart(
        task.uploadToken,
        chunkBytes,
        this.buildPartAAD(task, partNumber, partSize),
      );
      encryptedBytes = encrypted.encryptedChunk;

      const presigned = await this.presignPart(task.uploadSessionId, partNumber);
      const response = await fetch(presigned.url, {
        method: "PUT",
        body: encryptedBytes,
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
      task.partRecords[partNumber - 1] = {
        n: partNumber - 1,
        offset: start,
        plain_size: end - start,
        cipher_size: encrypted.encryptedSize,
        hash: encrypted.encryptedHash,
      };
      return end - start;
    } finally {
      if (chunkBytes) {
        chunkBytes = null;
      }
      if (encryptedBytes) {
        try {
          encryptedBytes.fill(0);
        } catch (_) {}
        encryptedBytes = null;
      }
    }
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

  buildPartAAD(task, partNumber, partSize) {
    return (
      "arkive:file-chunk:v1:" +
      task.vaultId +
      ":" +
      task.fileId +
      ":" +
      partNumber +
      ":" +
      partSize +
      ":" +
      task.totalParts
    );
  }

  buildManifest(task, metadata) {
    return {
      schema: "arkive.file.manifest",
      version: 1,
      file_id: task.fileId,
      name: metadata.name,
      mime: metadata.mime,
      extension: metadata.extension,
      size: metadata.size,
      chunk_size: task.chunkSize,
      chunks: task.partRecords.map(function (part) {
        return {
          n: part.n,
          offset: part.offset,
          plain_size: part.plain_size,
          cipher_size: part.cipher_size,
          hash: part.hash,
        };
      }),
    };
  }

  async waitForRetry(attempt, signal) {
    const jitter = Math.floor(Math.random() * 150);
    const delay = this.retryBaseDelayMs * Math.pow(2, attempt - 1) + jitter;
    await new Promise(function (resolve, reject) {
      let timer = 0;
      let onAbort = null;
      const finish = function () {
        if (timer) {
          window.clearTimeout(timer);
          timer = 0;
        }
        if (signal && onAbort) {
          signal.removeEventListener("abort", onAbort);
        }
      };
      if (!signal) {
        timer = window.setTimeout(resolve, delay);
        return;
      }
      onAbort = function () {
        finish();
        reject(new Error("Upload aborted"));
      };
      timer = window.setTimeout(function () {
        finish();
        resolve();
      }, delay);
      signal.addEventListener("abort", onAbort, { once: true });
    });
  }
}
