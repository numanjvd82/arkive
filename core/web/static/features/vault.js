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

const SESSION_UNLOCK_KEY = "arkive:vault-session:v1";
const SESSION_UNLOCK_TTL_MS = 60 * 60 * 1000;

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

function binaryTransfer(value) {
  if (value instanceof Uint8Array) {
    const exactBuffer = value.buffer.slice(
      value.byteOffset,
      value.byteOffset + value.byteLength,
    );
    return [exactBuffer];
  }
  if (value instanceof ArrayBuffer) {
    return [value];
  }
  return [];
}

function transferList(values) {
  const transfers = [];
  for (let i = 0; i < values.length; i++) {
    const current = binaryTransfer(values[i]);
    for (let j = 0; j < current.length; j++) {
      transfers.push(current[j]);
    }
  }
  return transfers;
}

export function initVault() {
  if (window.ArkiveVault && window.ArkiveVault.__arkiveReady) {
    return;
  }

  let worker = null;
  let requestID = 0;
  let unlocked = false;
  let restorePromise = null;
  let sessionExpiryTimer = null;
  const pending = new Map();

  function emitVaultState() {
    window.dispatchEvent(new CustomEvent("arkive:vault-state", {
      detail: { unlocked: !!unlocked }
    }));
  }

  function clearSessionExpiryTimer() {
    if (!sessionExpiryTimer) {
      return;
    }
    window.clearTimeout(sessionExpiryTimer);
    sessionExpiryTimer = null;
  }

  function scheduleSessionExpiry(expiresAt) {
    clearSessionExpiryTimer();
    if (typeof expiresAt !== "number" || expiresAt <= Date.now()) {
      return;
    }
    sessionExpiryTimer = window.setTimeout(function() {
      clearSessionUnlock();
      unlocked = false;
      callWorker("lockVault", {}).catch(function() {});
      emitVaultState();
    }, Math.max(0, expiresAt - Date.now()));
  }

  function loadSessionUnlock() {
    try {
      const raw = sessionStorage.getItem(SESSION_UNLOCK_KEY);
      if (!raw) {
        return null;
      }
      const parsed = JSON.parse(raw);
      if (
        !parsed ||
        typeof parsed.sessionUnwrapKey !== "string" ||
        typeof parsed.wrappedMasterKey !== "string" ||
        typeof parsed.expiresAt !== "number"
      ) {
        return null;
      }
      if (parsed.expiresAt <= Date.now()) {
        clearSessionUnlock();
        return null;
      }
      scheduleSessionExpiry(parsed.expiresAt);
      return parsed;
    } catch (_) {
      return null;
    }
  }

  function storeSessionUnlock(sessionUnwrapKey, wrappedMasterKey) {
    try {
      const expiresAt = Date.now() + SESSION_UNLOCK_TTL_MS;
      sessionStorage.setItem(
        SESSION_UNLOCK_KEY,
        JSON.stringify({
          sessionUnwrapKey: String(sessionUnwrapKey || ""),
          wrappedMasterKey: String(wrappedMasterKey || ""),
          expiresAt: expiresAt,
        }),
      );
      scheduleSessionExpiry(expiresAt);
    } catch (_) {}
  }

  function clearSessionUnlock() {
    clearSessionExpiryTimer();
    try {
      sessionStorage.removeItem(SESSION_UNLOCK_KEY);
    } catch (_) {}
  }

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
      emitVaultState();
      if (event && event.preventDefault) {
        event.preventDefault();
      }
    });
    return worker;
  }

  function callWorker(method, params, transfer) {
    return new Promise(function(resolve, reject) {
      const activeWorker = ensureWorker();
      const id = ++requestID;
      pending.set(id, { resolve: resolve, reject: reject });
      activeWorker.postMessage({
        id: id,
        method: method,
        params: params || {}
      }, transfer || []);
    });
  }

  async function unlockWithEnvelope(password, salt, encryptedMasterKey) {
    const result = await callWorker("unlockVault", {
      password: String(password || ""),
      salt: String(salt || ""),
      encryptedMasterKey: String(encryptedMasterKey || ""),
      aad: "arkive:master-key:v1"
    });
    unlocked = !!(result && result.unlocked);
    emitVaultState();
    return result;
  }

  async function createSessionUnlock() {
    const sessionUnwrapKey = new Uint8Array(32);
    window.crypto.getRandomValues(sessionUnwrapKey);
    const result = await callWorker(
      "createSessionUnlock",
      {
        sessionUnwrapKey: sessionUnwrapKey,
        aad: "arkive:session-master-key:v1",
      },
      binaryTransfer(sessionUnwrapKey),
    );
    const sessionUnwrapKeyBase64 = bytesToBase64(sessionUnwrapKey);
    sessionUnwrapKey.fill(0);
    return {
      sessionUnwrapKey: sessionUnwrapKeyBase64,
      wrappedMasterKey: String((result && result.wrappedMasterKey) || ""),
    };
  }

  async function restoreSessionUnlock() {
    const stored = loadSessionUnlock();
    if (!stored) {
      unlocked = false;
      return { unlocked: false };
    }
    try {
      const result = await callWorker("restoreSessionUnlock", {
        sessionUnwrapKey: stored.sessionUnwrapKey,
        wrappedMasterKey: stored.wrappedMasterKey,
        aad: "arkive:session-master-key:v1",
      });
      unlocked = !!(result && result.unlocked);
      if (unlocked) {
        storeSessionUnlock(stored.sessionUnwrapKey, stored.wrappedMasterKey);
      }
      emitVaultState();
      return result;
    } catch (error) {
      clearSessionUnlock();
      unlocked = false;
      emitVaultState();
      throw error;
    }
  }

  function touchSessionUnlock() {
    const stored = loadSessionUnlock();
    if (!stored) {
      return;
    }
    storeSessionUnlock(
      stored.sessionUnwrapKey,
      stored.wrappedMasterKey,
    );
  }

  function getSessionUnlock() {
    return loadSessionUnlock();
  }

  function onSessionUnlock(callback) {
    if (typeof callback === "function") {
      callback(loadSessionUnlock());
    }
  }

  function ensureRestored() {
    if (!restorePromise) {
      restorePromise = restoreSessionUnlock().catch(function () {
        return { unlocked: false };
      });
    }
    return restorePromise;
  }

  document.addEventListener("submit", function(event) {
    const form = event.target;
    if (!(form instanceof HTMLFormElement)) {
      return;
    }
    const action = form.getAttribute("action") || "";
    if (action !== "/logout") {
      return;
    }
    clearSessionUnlock();
    unlocked = false;
    emitVaultState();
    callWorker("lockVault", {}).catch(function () {});
  });

  window.ArkiveVault = {
    __arkiveReady: true,
    waitUntilReady: function() {
      return ensureRestored();
    },
    unlockVault: async function(password, salt, encryptedMasterKey) {
      const result = await unlockWithEnvelope(password, salt, encryptedMasterKey);
      restorePromise = Promise.resolve(result);
      return result;
    },
    persistSessionUnlock: async function() {
      const sessionState = await createSessionUnlock();
      storeSessionUnlock(
        sessionState.sessionUnwrapKey,
        sessionState.wrappedMasterKey,
      );
      restorePromise = Promise.resolve({ unlocked: true });
      return sessionState;
    },
    clearSessionUnlock: function() {
      clearSessionUnlock();
    },
    lock: async function() {
      unlocked = false;
      clearSessionUnlock();
      await callWorker("lockVault", {});
      emitVaultState();
    },
    isUnlocked: function() {
      return unlocked;
    },
    generateFileKey: function() {
      touchSessionUnlock();
      return callWorker("generateFileKey", {});
    },
    prepareShare: function(record, token) {
      touchSessionUnlock();
      return callWorker("prepareShare", {
        vaultId: String((record && record.vaultId) || ""),
        fileId: String((record && record.fileId) || ""),
        token: String(token || ""),
        encryptedFileKey: String((record && record.encryptedFileKey) || ""),
        fileKeyAad: "arkive:file-key:v1:" + String((record && record.vaultId) || "") + ":" + String((record && record.fileId) || ""),
        shareKeyAad: "arkive:share-key:v1:" + String(token || ""),
        shareFileKeyAad: "arkive:share-file-key:v1:" + String((record && record.fileId) || "") + ":" + String(token || ""),
      });
    },
    openShareKey: function(encryptedShareKey, token) {
      touchSessionUnlock();
      return callWorker("openShareKey", {
        encryptedShareKey: String(encryptedShareKey || ""),
        shareKeyAad: "arkive:share-key:v1:" + String(token || ""),
      });
    },
    prepareUpload: function(uploadToken, vaultId, fileId, metadata, totalParts) {
      touchSessionUnlock();
      return callWorker("prepareUpload", {
        uploadToken: String(uploadToken || ""),
        vaultId: String(vaultId || ""),
        fileId: String(fileId || ""),
        metadata: metadata || {},
        totalParts: Number(totalParts || 0),
        metadataAad: "arkive:file-metadata:v1:" + String(vaultId || "") + ":" + String(fileId || ""),
        fileKeyAad: "arkive:file-key:v1:" + String(vaultId || "") + ":" + String(fileId || "")
      });
    },
    encryptUploadPart: function(uploadToken, chunkBytes, aad) {
      touchSessionUnlock();
      const payload =
        chunkBytes instanceof Uint8Array
          ? chunkBytes.slice()
          : chunkBytes instanceof ArrayBuffer
            ? chunkBytes.slice(0)
            : chunkBytes;
      return callWorker("encryptUploadPart", {
        uploadToken: String(uploadToken || ""),
        chunkBytes: payload,
        aad: aad || ""
      }, binaryTransfer(payload));
    },
    finalizeUpload: function(uploadToken, vaultId, fileId, manifest, partHashes) {
      touchSessionUnlock();
      return callWorker("finalizeUpload", {
        uploadToken: String(uploadToken || ""),
        manifest: manifest || {},
        partHashes: partHashes || [],
        manifestAad: "arkive:file-manifest:v1:" + String(vaultId || "") + ":" + String(fileId || "")
      });
    },
    clearUploadContext: function(uploadToken) {
      return callWorker("clearUploadContext", {
        uploadToken: String(uploadToken || ""),
      });
    },
    openFileContext: function(contextId, record) {
      touchSessionUnlock();
      return callWorker("openFileContext", {
        contextId: String(contextId || ""),
        encryptedFileKey: String((record && record.encryptedFileKey) || ""),
        encryptedMetadata: String((record && record.encryptedMetadata) || ""),
        encryptedManifest: String((record && record.encryptedManifest) || ""),
        fileKeyAad: "arkive:file-key:v1:" + String((record && record.vaultId) || "") + ":" + String((record && record.fileId) || ""),
        metadataAad: "arkive:file-metadata:v1:" + String((record && record.vaultId) || "") + ":" + String((record && record.fileId) || ""),
        manifestAad: "arkive:file-manifest:v1:" + String((record && record.vaultId) || "") + ":" + String((record && record.fileId) || ""),
      });
    },
    closeFileContext: function(contextId) {
      return callWorker("closeFileContext", {
        contextId: String(contextId || ""),
      });
    },
    decryptFileChunk: function(contextId, encryptedChunk, aad, expectedHash) {
      touchSessionUnlock();
      const payload =
        encryptedChunk instanceof Uint8Array
          ? encryptedChunk.slice()
          : encryptedChunk instanceof ArrayBuffer
            ? encryptedChunk.slice(0)
            : encryptedChunk;
      return callWorker("decryptFileChunk", {
        contextId: String(contextId || ""),
        encryptedChunk: payload,
        aad: aad || "",
        expectedHash: expectedHash || "",
      }, transferList([payload]));
    },
    encryptFileMetadata: function(metadata, masterKey) {
      touchSessionUnlock();
      return callWorker("encryptFileMetadata", {
        metadata: metadata || {},
        masterKey: toBase64(masterKey),
        aad: "arkive:file-metadata:v1"
      });
    },
    encryptFileKey: function(fileKey, masterKey, aad) {
      touchSessionUnlock();
      return callWorker("encryptFileKey", {
        fileKey: toBase64(fileKey),
        masterKey: toBase64(masterKey),
        aad: aad || "arkive:file-key:v1"
      });
    },
    encryptChunk: function(chunkBytes, fileKey, aad) {
      touchSessionUnlock();
      return callWorker("encryptChunk", {
        chunkBytes: toBase64(chunkBytes),
        fileKey: toBase64(fileKey),
        aad: aad || ""
      });
    },
    decryptChunk: function(encryptedChunk, fileKey, aad) {
      touchSessionUnlock();
      return callWorker("decryptChunk", {
        encryptedChunk: toBase64(encryptedChunk),
        fileKey: toBase64(fileKey),
        aad: aad || ""
      });
    },
    getSessionUnlock: getSessionUnlock,
    onSessionUnlock: onSessionUnlock
  };

  ensureRestored();
}
