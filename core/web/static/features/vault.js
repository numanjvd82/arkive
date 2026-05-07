function bytesToBase64(bytes) {
  if (!(bytes instanceof Uint8Array)) {
    return "";
  }
  let binary = "";
  for (let i = 0; i < bytes.length; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}

function toBase64(value) {
  if (!value) {
    return "";
  }
  if (typeof value === "string") {
    return value;
  }
  if (value instanceof Uint8Array) {
    return bytesToBase64(value);
  }
  if (value instanceof ArrayBuffer) {
    return bytesToBase64(new Uint8Array(value));
  }
  return "";
}

export function initVault() {
  if (window.ArkiveVault && window.ArkiveVault.__arkiveReady) {
    return;
  }

  let worker = null;
  let requestID = 0;
  let unlocked = false;
  const pending = new Map();

  function ensureWorker() {
    if (worker) {
      return worker;
    }

    worker = new Worker("/static/workers/crypto_worker.js", { type: "module" });
    worker.addEventListener("message", function(event) {
      const data = event.data || {};
      const entry = pending.get(data.id);
      if (!entry) {
        return;
      }
      pending.delete(data.id);
      if (data.ok) {
        entry.resolve(data.result || {});
        return;
      }
      entry.reject(new Error(data.error || "Vault worker failed"));
    });
    worker.addEventListener("error", function(event) {
      pending.forEach(function(entry) {
        entry.reject(new Error("Vault worker failed"));
      });
      pending.clear();
      worker = null;
      unlocked = false;
      if (event && event.preventDefault) {
        event.preventDefault();
      }
    });
    return worker;
  }

  function callWorker(method, params) {
    return new Promise(function(resolve, reject) {
      const activeWorker = ensureWorker();
      const id = ++requestID;
      pending.set(id, { resolve: resolve, reject: reject });
      activeWorker.postMessage({
        id: id,
        method: method,
        params: params || {}
      });
    });
  }

  window.ArkiveVault = {
    __arkiveReady: true,
    unlockVault: async function(password, salt, encryptedMasterKey) {
      const result = await callWorker("unlockVault", {
        password: String(password || ""),
        salt: String(salt || ""),
        encryptedMasterKey: String(encryptedMasterKey || ""),
        aad: "arkive:master-key:v1"
      });
      unlocked = !!(result && result.unlocked);
      return result;
    },
    lock: async function() {
      unlocked = false;
      await callWorker("lockVault", {});
    },
    isUnlocked: function() {
      return unlocked;
    },
    generateFileKey: function() {
      return callWorker("generateFileKey", {});
    },
    prepareUpload: function(uploadToken, metadata, totalParts) {
      return callWorker("prepareUpload", {
        uploadToken: String(uploadToken || ""),
        metadata: metadata || {},
        totalParts: Number(totalParts || 0),
        metadataAad: "arkive:file-metadata:v1",
        fileKeyAad: "arkive:file-key:v1"
      });
    },
    encryptUploadPart: function(uploadToken, chunkBytes, aad) {
      return callWorker("encryptUploadPart", {
        uploadToken: String(uploadToken || ""),
        chunkBytes: toBase64(chunkBytes),
        aad: aad || ""
      });
    },
    finalizeUpload: function(uploadToken) {
      return callWorker("finalizeUpload", {
        uploadToken: String(uploadToken || "")
      });
    },
    encryptFileMetadata: function(metadata, masterKey) {
      return callWorker("encryptFileMetadata", {
        metadata: metadata || {},
        masterKey: toBase64(masterKey),
        aad: "arkive:file-metadata:v1"
      });
    },
    encryptFileKey: function(fileKey, masterKey, aad) {
      return callWorker("encryptFileKey", {
        fileKey: toBase64(fileKey),
        masterKey: toBase64(masterKey),
        aad: aad || "arkive:file-key:v1"
      });
    },
    encryptChunk: function(chunkBytes, fileKey, aad) {
      return callWorker("encryptChunk", {
        chunkBytes: toBase64(chunkBytes),
        fileKey: toBase64(fileKey),
        aad: aad || ""
      });
    },
    decryptChunk: function(encryptedChunk, fileKey, aad) {
      return callWorker("decryptChunk", {
        encryptedChunk: toBase64(encryptedChunk),
        fileKey: toBase64(fileKey),
        aad: aad || ""
      });
    }
  };
}
