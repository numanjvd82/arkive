function bytesToBase64(bytes) {
  if (!(bytes instanceof Uint8Array)) {
    return "";
  }
  let binary = "";
  for (let i = 0; i < bytes.length; i += 1) {
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
  for (let i = 0; i < values.length; i += 1) {
    const current = binaryTransfer(values[i]);
    for (let j = 0; j < current.length; j += 1) {
      transfers.push(current[j]);
    }
  }
  return transfers;
}

let logoutBound = false;
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

export function clearSessionUnlock() {
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
  storeSessionUnlock(stored.sessionUnwrapKey, stored.wrappedMasterKey);
}

function normalizeText(value) {
  return String(value || "")
    .toLowerCase()
    .normalize("NFKD")
    .replace(/[^\p{L}\p{N}.]+/gu, " ")
    .trim();
}

function termsForFile(metadata) {
  const name = normalizeText(metadata && metadata.name);
  const mime = normalizeText(metadata && metadata.mime);
  const ext = name.includes(".") ? name.split(".").pop() : "";
  const words = name.replace(/\./g, " ").split(/\s+/).filter(Boolean);
  const terms = [];

  words.forEach(function(word) {
    terms.push({ term: word, field: "name", weight: 10 });
    if (word.length >= 3) {
      for (let i = 3; i <= Math.min(word.length, 32); i += 1) {
        terms.push({ term: word.slice(0, i), field: "prefix", weight: 1 });
      }
    }
  });
  if (ext) {
    terms.push({ term: ext, field: "ext", weight: 4 });
  }
  if (mime) {
    terms.push({ term: mime, field: "mime", weight: 2 });
  }
  return terms;
}

function termsForQuery(query) {
  const normalized = normalizeText(query).replace(/\./g, " ");
  const words = normalized.split(/\s+/).filter(Boolean);
  const terms = [];
  words.forEach(function(word) {
    terms.push(word);
    if (word.length >= 3) {
      for (let i = 3; i <= Math.min(word.length, 32); i += 1) {
        terms.push(word.slice(0, i));
      }
    }
  });
  return Array.from(new Set(terms)).slice(0, 32);
}

export function getSessionUnlock() {
  return loadSessionUnlock();
}

export function onSessionUnlock(callback) {
  if (typeof callback === "function") {
    callback(loadSessionUnlock());
  }
}

function ensureRestored() {
  if (!restorePromise) {
    restorePromise = restoreSessionUnlock().catch(function() {
      return { unlocked: false };
    });
  }
  return restorePromise;
}

function bindLogoutHandler() {
  if (logoutBound) {
    return;
  }
  logoutBound = true;
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
    callWorker("lockVault", {}).catch(function() {});
  });
}

export const vault = {
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
  clearSessionUnlock: clearSessionUnlock,
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
  openPublicShareContext: function(contextId, record, shareSecret) {
    return callWorker("openPublicShareContext", {
      contextId: String(contextId || ""),
      shareSecret: String(shareSecret || ""),
      encryptedFileKeyForShare: String((record && record.encryptedFileKeyForShare) || ""),
      encryptedMetadata: String((record && record.encryptedMetadata) || ""),
      encryptedManifest: String((record && record.encryptedManifest) || ""),
      shareFileKeyAad: String((record && record.shareFileKeyAad) || ""),
      metadataAad: String((record && record.metadataAad) || ""),
      manifestAad: String((record && record.manifestAad) || ""),
    });
  },
  closePublicShareContext: function(contextId) {
    return callWorker("closePublicShareContext", {
      contextId: String(contextId || ""),
    });
  },
  decryptPublicShareChunk: function(contextId, encryptedChunk, aad, expectedHash, hashEncoding) {
    const payload =
      encryptedChunk instanceof Uint8Array
        ? encryptedChunk.slice()
        : encryptedChunk instanceof ArrayBuffer
          ? encryptedChunk.slice(0)
          : encryptedChunk;
    return callWorker("decryptPublicShareChunk", {
      contextId: String(contextId || ""),
      encryptedChunk: payload,
      aad: aad || "",
      expectedHash: expectedHash || "",
      hashEncoding: hashEncoding || "base64",
    }, transferList([payload]));
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
  encryptFileMetadataForFile: function(metadata, vaultId, fileId, masterKey) {
    touchSessionUnlock();
    return callWorker("encryptFileMetadata", {
      metadata: metadata || {},
      masterKey: toBase64(masterKey),
      aad: "arkive:file-metadata:v1:" + String(vaultId || "") + ":" + String(fileId || "")
    });
  },
  encryptFileMetadataInContext: function(contextId, metadata, vaultId, fileId) {
    touchSessionUnlock();
    return callWorker("encryptFileMetadataInContext", {
      contextId: String(contextId || ""),
      metadata: metadata || {},
      aad: "arkive:file-metadata:v1:" + String(vaultId || "") + ":" + String(fileId || "")
    });
  },
  decryptFileMetadata: function(encryptedMetadata, masterKey) {
    touchSessionUnlock();
    return callWorker("decryptFileMetadata", {
      encryptedMetadata: toBase64(encryptedMetadata),
      masterKey: toBase64(masterKey),
      aad: "arkive:file-metadata:v1"
    });
  },
  encryptFolderName: function(metadata, masterKey) {
    touchSessionUnlock();
    return callWorker("encryptFolderMetadata", {
      metadata: metadata || {},
      masterKey: toBase64(masterKey),
      aad: "arkive:folder-name:v1"
    });
  },
  encryptFolderMetadata: function(metadata, masterKey) {
    touchSessionUnlock();
    return callWorker("encryptFolderMetadata", {
      metadata: metadata || {},
      masterKey: toBase64(masterKey),
      aad: "arkive:folder-metadata:v1"
    });
  },
  decryptFolderName: function(encryptedMetadata, masterKey) {
    touchSessionUnlock();
    return callWorker("decryptFolderMetadata", {
      encryptedMetadata: toBase64(encryptedMetadata),
      masterKey: toBase64(masterKey),
      aad: "arkive:folder-name:v1"
    });
  },
  decryptFolderMetadata: function(encryptedMetadata, masterKey) {
    touchSessionUnlock();
    return callWorker("decryptFolderMetadata", {
      encryptedMetadata: toBase64(encryptedMetadata),
      masterKey: toBase64(masterKey),
      aad: "arkive:folder-metadata:v1"
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
  createSearchTokens: async function(query, vaultId) {
    touchSessionUnlock();
    const terms = termsForQuery(query);
    if (!terms.length) {
      return [];
    }
    const result = await callWorker("createSearchTokens", {
      vaultId: String(vaultId || document.body.getAttribute("data-vault-id") || ""),
      terms: terms,
    });
    return Array.isArray(result && result.tokens) ? result.tokens : [];
  },
  createSearchTokenEntries: async function(vaultId, metadata) {
    touchSessionUnlock();
    const items = termsForFile(metadata).slice(0, 128);
    if (!items.length) {
      return [];
    }
    const result = await callWorker("createSearchTokens", {
      vaultId: String(vaultId || document.body.getAttribute("data-vault-id") || ""),
      terms: items.map(function(item) { return item.term; }),
    });
    const tokens = Array.isArray(result && result.tokens) ? result.tokens : [];
    return items.slice(0, tokens.length).map(function(item, index) {
      return {
        token: tokens[index],
        field: item.field,
        weight: item.weight,
      };
    });
  },
  decryptSearchFile: function(file) {
    touchSessionUnlock();
    return callWorker("decryptSearchFile", {
      fileId: String((file && file.id) || ""),
      vaultId: String((file && file.vaultId) || ""),
      encryptedMetadata: String((file && file.encryptedMetadata) || ""),
      encryptedFileKey: String((file && file.encryptedFileKey) || ""),
      fileKeyAad: "arkive:file-key:v1:" + String((file && file.vaultId) || "") + ":" + String((file && file.id) || ""),
      metadataAad: "arkive:file-metadata:v1:" + String((file && file.vaultId) || "") + ":" + String((file && file.id) || ""),
    });
  },
  getSessionUnlock: getSessionUnlock,
  onSessionUnlock: onSessionUnlock,
};

export function waitUntilReady() {
  return vault.waitUntilReady();
}

export function unlockVault(password, salt, encryptedMasterKey) {
  return vault.unlockVault(password, salt, encryptedMasterKey);
}

export function persistSessionUnlock() {
  return vault.persistSessionUnlock();
}

export function lockVault() {
  return vault.lock();
}

export function getVaultState() {
  return { unlocked: vault.isUnlocked() };
}

export async function requireUnlockedVault() {
  await waitUntilReady();
  if (!vault.isUnlocked()) {
    throw new Error("Vault is locked.");
  }
  return vault;
}

export function onVaultUnlock(callback) {
  if (typeof callback !== "function") {
    return function() {};
  }
  function handler(event) {
    const detail = event && event.detail ? event.detail : {};
    if (detail.unlocked) {
      callback(detail);
    }
  }
  window.addEventListener("arkive:vault-state", handler);
  if (vault.isUnlocked()) {
    callback({ unlocked: true });
  }
  return function() {
    window.removeEventListener("arkive:vault-state", handler);
  };
}

export function onVaultLock(callback) {
  if (typeof callback !== "function") {
    return function() {};
  }
  function handler(event) {
    const detail = event && event.detail ? event.detail : {};
    if (!detail.unlocked) {
      callback(detail);
    }
  }
  window.addEventListener("arkive:vault-state", handler);
  return function() {
    window.removeEventListener("arkive:vault-state", handler);
  };
}

export function initVault() {
  bindLogoutHandler();
  ensureRestored();
  window.ArkiveVault = vault;
}
