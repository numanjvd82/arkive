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

export function initVault() {
  if (window.ArkiveVault && window.ArkiveVault.__arkiveReady) {
    return;
  }

  let worker = null;
  let requestID = 0;
  let unlocked = false;
  let restorePromise = null;
  const pending = new Map();

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
      return parsed;
    } catch (_) {
      return null;
    }
  }

  function storeSessionUnlock(sessionUnwrapKey, wrappedMasterKey) {
    try {
      sessionStorage.setItem(
        SESSION_UNLOCK_KEY,
        JSON.stringify({
          sessionUnwrapKey: String(sessionUnwrapKey || ""),
          wrappedMasterKey: String(wrappedMasterKey || ""),
          expiresAt: Date.now() + SESSION_UNLOCK_TTL_MS,
        }),
      );
    } catch (_) {}
  }

  function clearSessionUnlock() {
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
      return result;
    } catch (error) {
      clearSessionUnlock();
      unlocked = false;
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
    },
    isUnlocked: function() {
      return unlocked;
    },
    generateFileKey: function() {
      touchSessionUnlock();
      return callWorker("generateFileKey", {});
    },
    prepareUpload: function(uploadToken, metadata, totalParts) {
      touchSessionUnlock();
      return callWorker("prepareUpload", {
        uploadToken: String(uploadToken || ""),
        metadata: metadata || {},
        totalParts: Number(totalParts || 0),
        metadataAad: "arkive:file-metadata:v1",
        fileKeyAad: "arkive:file-key:v1"
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
    finalizeUpload: function(uploadToken) {
      touchSessionUnlock();
      return callWorker("finalizeUpload", {
        uploadToken: String(uploadToken || "")
      });
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
    }
  };

  ensureRestored();
}
