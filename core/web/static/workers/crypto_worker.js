import initArkiveCrypto, * as arkiveCrypto from "../vendor/arkive-crypto/arkive_crypto.js";

let readyPromise = null;
let unlockedMasterKey = null;
let activeUploads = new Map();
let activeReaders = new Map();
let activeShareReaders = new Map();

function ensureCrypto() {
  if (readyPromise) {
    return readyPromise;
  }
  readyPromise = initArkiveCrypto({
    module_or_path: "/static/vendor/arkive-crypto/arkive_crypto_bg.wasm",
  }).then(function () {
    return arkiveCrypto;
  });
  return readyPromise;
}

function encodeBase64(bytes) {
  let binary = "";
  for (let i = 0; i < bytes.length; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary);
}

function encodeBase64URL(bytes) {
  return encodeBase64(bytes).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/g, "");
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

function toUint8Array(value) {
  if (value instanceof Uint8Array) {
    return value;
  }
  if (value instanceof ArrayBuffer) {
    return new Uint8Array(value);
  }
  if (ArrayBuffer.isView(value)) {
    return new Uint8Array(value.buffer, value.byteOffset, value.byteLength);
  }
  return decodeBase64(value);
}

function aadBytes(value) {
  if (!value) {
    return new Uint8Array();
  }
  if (value instanceof Uint8Array) {
    return value;
  }
  return new TextEncoder().encode(String(value));
}

function lockVault(crypto) {
  if (unlockedMasterKey) {
    crypto.zeroize(unlockedMasterKey);
    unlockedMasterKey = null;
  }
  activeUploads.forEach(function(fileKey) {
    if (fileKey && fileKey.fileKey) {
      crypto.zeroize(fileKey.fileKey);
    }
  });
  activeUploads.clear();
  activeReaders.forEach(function(fileKey) {
    crypto.zeroize(fileKey);
  });
  activeReaders.clear();
  activeShareReaders.forEach(function(fileKey) {
    crypto.zeroize(fileKey);
  });
  activeShareReaders.clear();
}

function activeMasterKey(supplied) {
  if (supplied) {
    return {
      bytes: decodeBase64(supplied),
      temporary: true,
    };
  }
  if (!unlockedMasterKey) {
    throw new Error("Vault is locked");
  }
  return {
    bytes: unlockedMasterKey,
    temporary: false,
  };
}

async function deriveSearchKey(masterKey) {
  const hkdfKey = await globalThis.crypto.subtle.importKey(
    "raw",
    masterKey,
    "HKDF",
    false,
    ["deriveBits"],
  );
  const bits = await globalThis.crypto.subtle.deriveBits(
    {
      name: "HKDF",
      hash: "SHA-256",
      salt: new Uint8Array(),
      info: new TextEncoder().encode("arkive-search-v1"),
    },
    hkdfKey,
    256,
  );
  return new Uint8Array(bits);
}

async function handleMessage(message) {
  const crypto = await ensureCrypto();

  switch (message.method) {
    case "unlockVault": {
      const salt = decodeBase64(message.params.salt);
      const encryptedMasterKey = decodeBase64(
        message.params.encryptedMasterKey,
      );
      const kek = crypto.derive_password_kek(
        String(message.params.password || ""),
        salt,
      );
      const masterKey = crypto.unwrap_master_key(
        encryptedMasterKey,
        kek,
        aadBytes(message.params.aad),
      );
      lockVault(crypto);
      unlockedMasterKey = masterKey.slice();
      crypto.zeroize(kek);
      crypto.zeroize(masterKey);
      return { unlocked: true };
    }
    case "lockVault": {
      lockVault(crypto);
      return { unlocked: false };
    }
    case "createSessionUnlock": {
      if (!unlockedMasterKey) {
        throw new Error("Vault is locked");
      }
      const sessionUnwrapKey = toUint8Array(message.params.sessionUnwrapKey);
      try {
        const wrappedMasterKey = crypto.wrap_master_key(
          unlockedMasterKey,
          sessionUnwrapKey,
          aadBytes(message.params.aad),
        );
        try {
          return {
            wrappedMasterKey: encodeBase64(wrappedMasterKey),
          };
        } finally {
          crypto.zeroize(wrappedMasterKey);
        }
      } finally {
        crypto.zeroize(sessionUnwrapKey);
      }
    }
    case "restoreSessionUnlock": {
      const sessionUnwrapKey = toUint8Array(message.params.sessionUnwrapKey);
      const wrappedMasterKey = decodeBase64(message.params.wrappedMasterKey);
      try {
        const masterKey = crypto.unwrap_master_key(
          wrappedMasterKey,
          sessionUnwrapKey,
          aadBytes(message.params.aad),
        );
        try {
          lockVault(crypto);
          unlockedMasterKey = masterKey.slice();
          return { unlocked: true };
        } finally {
          crypto.zeroize(masterKey);
        }
      } finally {
        crypto.zeroize(sessionUnwrapKey);
        crypto.zeroize(wrappedMasterKey);
      }
    }
    case "generateFileKey": {
      const fileKey = crypto.generate_file_key();
      try {
        return { fileKey: encodeBase64(fileKey) };
      } finally {
        crypto.zeroize(fileKey);
      }
    }
    case "prepareShare": {
      if (!unlockedMasterKey) {
        throw new Error("Vault is locked");
      }
      const encryptedFileKey = decodeBase64(message.params.encryptedFileKey);
      try {
        const fileKey = crypto.unwrap_file_key(
          encryptedFileKey,
          unlockedMasterKey,
          aadBytes(message.params.fileKeyAad),
        );
        try {
          const shareKey = crypto.generate_share_key();
          try {
            const encryptedShareKey = crypto.wrap_file_key(
              shareKey,
              unlockedMasterKey,
              aadBytes(message.params.shareKeyAad),
            );
            try {
              const encryptedFileKeyForShare = crypto.wrap_file_key(
                fileKey,
                shareKey,
                aadBytes(message.params.shareFileKeyAad),
              );
              try {
                return {
                  encryptedShareKey: encodeBase64(encryptedShareKey),
                  encryptedFileKeyForShare: encodeBase64(encryptedFileKeyForShare),
                  shareSecret: encodeBase64(shareKey),
                  cryptoVersion: 1,
                };
              } finally {
                crypto.zeroize(encryptedFileKeyForShare);
              }
            } finally {
              crypto.zeroize(encryptedShareKey);
            }
          } finally {
            crypto.zeroize(shareKey);
          }
        } finally {
          crypto.zeroize(fileKey);
        }
      } finally {
        crypto.zeroize(encryptedFileKey);
      }
    }
    case "openShareKey": {
      if (!unlockedMasterKey) {
        throw new Error("Vault is locked");
      }
      const encryptedShareKey = decodeBase64(message.params.encryptedShareKey);
      try {
        const shareKey = crypto.unwrap_file_key(
          encryptedShareKey,
          unlockedMasterKey,
          aadBytes(message.params.shareKeyAad),
        );
        try {
          return { shareSecret: encodeBase64(shareKey) };
        } finally {
          crypto.zeroize(shareKey);
        }
      } finally {
        crypto.zeroize(encryptedShareKey);
      }
    }
    case "prepareUpload": {
      const uploadToken = String(message.params.uploadToken || "");
      if (!uploadToken) {
        throw new Error("Missing upload token");
      }
      if (!unlockedMasterKey) {
        throw new Error("Vault is locked");
      }
      const metadata = new TextEncoder().encode(
        JSON.stringify(message.params.metadata || {}),
      );
      const fileKey = crypto.generate_file_key();
      try {
        const encryptedMetadata = crypto.encrypt_chunk(
          metadata,
          fileKey,
          aadBytes(message.params.metadataAad),
        );
        try {
          const encryptedFileKey = crypto.wrap_file_key(
            fileKey,
            unlockedMasterKey,
            aadBytes(message.params.fileKeyAad),
          );
          try {
            activeUploads.set(uploadToken, {
              fileKey: fileKey.slice(),
            });
            return {
              encryptedMetadata: encodeBase64(encryptedMetadata),
              encryptedFileKey: encodeBase64(encryptedFileKey),
              totalParts: Number(message.params.totalParts || 0),
              encryptionVersion: 1,
            };
          } finally {
            crypto.zeroize(encryptedFileKey);
          }
        } finally {
          crypto.zeroize(encryptedMetadata);
        }
      } finally {
        crypto.zeroize(metadata);
        crypto.zeroize(fileKey);
      }
    }
    case "encryptUploadPart": {
      const uploadToken = String(message.params.uploadToken || "");
      const upload = activeUploads.get(uploadToken);
      if (!upload || !upload.fileKey) {
        throw new Error("Upload context is missing");
      }
      const chunkBytes = toUint8Array(message.params.chunkBytes);
      try {
        const encryptedChunk = crypto.encrypt_chunk(
          chunkBytes,
          upload.fileKey,
          aadBytes(message.params.aad),
        );
        try {
          const encryptedHash = crypto.hash_bytes_blake3(encryptedChunk);
          try {
            return {
              encryptedChunk: encryptedChunk.slice(),
              encryptedHash: encodeBase64(encryptedHash),
              encryptedSize: encryptedChunk.length,
            };
          } finally {
            crypto.zeroize(encryptedHash);
          }
        } finally {
          crypto.zeroize(encryptedChunk);
        }
      } finally {
        crypto.zeroize(chunkBytes);
      }
    }
    case "finalizeUpload": {
      const uploadToken = String(message.params.uploadToken || "");
      const upload = activeUploads.get(uploadToken);
      if (!upload || !upload.fileKey) {
        throw new Error("Upload context is missing");
      }
      const manifest = new TextEncoder().encode(
        JSON.stringify(message.params.manifest || {}),
      );
      let encryptedHash = new Uint8Array();
      try {
        const encryptedManifest = crypto.encrypt_chunk(
          manifest,
          upload.fileKey,
          aadBytes(message.params.manifestAad),
        );
        try {
          const partHashes = message.params.partHashes || [];
          let totalLength = 0;
          const decoded = [];
          for (let i = 0; i < partHashes.length; i++) {
            const hashBytes = decodeBase64(partHashes[i]);
            decoded.push(hashBytes);
            totalLength += hashBytes.length;
          }
          const combined = new Uint8Array(totalLength);
          let offset = 0;
          for (let i = 0; i < decoded.length; i++) {
            combined.set(decoded[i], offset);
            offset += decoded[i].length;
            crypto.zeroize(decoded[i]);
          }
          try {
            encryptedHash = crypto.hash_bytes_blake3(combined);
          } finally {
            crypto.zeroize(combined);
          }
          return {
            encryptedManifest: encodeBase64(encryptedManifest),
            encryptedHash: encodeBase64(encryptedHash),
          };
        } finally {
          crypto.zeroize(encryptedManifest);
        }
      } finally {
        crypto.zeroize(manifest);
        crypto.zeroize(encryptedHash);
        crypto.zeroize(upload.fileKey);
        activeUploads.delete(uploadToken);
      }
    }
    case "clearUploadContext": {
      const uploadToken = String(message.params.uploadToken || "");
      const upload = activeUploads.get(uploadToken);
      if (upload && upload.fileKey) {
        crypto.zeroize(upload.fileKey);
        activeUploads.delete(uploadToken);
      }
      return { cleared: true };
    }
    case "openFileContext": {
      const contextID = String(message.params.contextId || "");
      if (!contextID) {
        throw new Error("Missing file context");
      }
      if (!unlockedMasterKey) {
        throw new Error("Vault is locked");
      }
      const encryptedFileKey = decodeBase64(message.params.encryptedFileKey);
      const encryptedMetadata = decodeBase64(message.params.encryptedMetadata);
      const encryptedManifest = decodeBase64(message.params.encryptedManifest);
      try {
        const fileKey = crypto.unwrap_file_key(
          encryptedFileKey,
          unlockedMasterKey,
          aadBytes(message.params.fileKeyAad),
        );
        try {
          const metadataBytes = crypto.decrypt_chunk(
            encryptedMetadata,
            fileKey,
            aadBytes(message.params.metadataAad),
          );
          try {
            const manifestBytes = crypto.decrypt_chunk(
              encryptedManifest,
              fileKey,
              aadBytes(message.params.manifestAad),
            );
            try {
              activeReaders.set(contextID, fileKey.slice());
              return {
                metadata: new TextDecoder().decode(metadataBytes),
                manifest: new TextDecoder().decode(manifestBytes),
              };
            } finally {
              crypto.zeroize(manifestBytes);
            }
          } finally {
            crypto.zeroize(metadataBytes);
          }
        } finally {
          crypto.zeroize(fileKey);
        }
      } finally {
        crypto.zeroize(encryptedFileKey);
        crypto.zeroize(encryptedMetadata);
        crypto.zeroize(encryptedManifest);
      }
    }
    case "closeFileContext": {
      const contextID = String(message.params.contextId || "");
      const fileKey = activeReaders.get(contextID);
      if (fileKey) {
        crypto.zeroize(fileKey);
        activeReaders.delete(contextID);
      }
      return { cleared: true };
    }
    case "openPublicShareContext": {
      const contextID = String(message.params.contextId || "");
      if (!contextID) {
        throw new Error("Missing share context");
      }
      const shareSecret = decodeBase64(message.params.shareSecret);
      const encryptedFileKeyForShare = decodeBase64(message.params.encryptedFileKeyForShare);
      const encryptedMetadata = decodeBase64(message.params.encryptedMetadata);
      const encryptedManifest = decodeBase64(message.params.encryptedManifest);
      try {
        const fileKey = crypto.unwrap_file_key(
          encryptedFileKeyForShare,
          shareSecret,
          aadBytes(message.params.shareFileKeyAad),
        );
        try {
          const metadataBytes = crypto.decrypt_chunk(
            encryptedMetadata,
            fileKey,
            aadBytes(message.params.metadataAad),
          );
          try {
            const manifestBytes = crypto.decrypt_chunk(
              encryptedManifest,
              fileKey,
              aadBytes(message.params.manifestAad),
            );
            try {
              activeShareReaders.set(contextID, fileKey.slice());
              return {
                metadata: new TextDecoder().decode(metadataBytes),
                manifest: new TextDecoder().decode(manifestBytes),
              };
            } finally {
              crypto.zeroize(manifestBytes);
            }
          } finally {
            crypto.zeroize(metadataBytes);
          }
        } finally {
          crypto.zeroize(fileKey);
        }
      } finally {
        crypto.zeroize(shareSecret);
        crypto.zeroize(encryptedFileKeyForShare);
        crypto.zeroize(encryptedMetadata);
        crypto.zeroize(encryptedManifest);
      }
    }
    case "closePublicShareContext": {
      const contextID = String(message.params.contextId || "");
      const fileKey = activeShareReaders.get(contextID);
      if (fileKey) {
        crypto.zeroize(fileKey);
        activeShareReaders.delete(contextID);
      }
      return { cleared: true };
    }
    case "decryptFileChunk": {
      const contextID = String(message.params.contextId || "");
      const fileKey = activeReaders.get(contextID);
      if (!fileKey) {
        throw new Error("File context is missing");
      }
      const encryptedChunk = toUint8Array(message.params.encryptedChunk);
      try {
        const expectedHash = String(message.params.expectedHash || "");
        if (expectedHash) {
          const actualHash = crypto.hash_bytes_blake3(encryptedChunk);
          try {
            if (encodeBase64(actualHash) !== expectedHash) {
              throw new Error("Encrypted chunk hash mismatch");
            }
          } finally {
            crypto.zeroize(actualHash);
          }
        }
        const chunkBytes = crypto.decrypt_chunk(
          encryptedChunk,
          fileKey,
          aadBytes(message.params.aad),
        );
        try {
          return { chunkBytes: chunkBytes.slice() };
        } finally {
          crypto.zeroize(chunkBytes);
        }
      } finally {
        crypto.zeroize(encryptedChunk);
      }
    }
    case "decryptPublicShareChunk": {
      const contextID = String(message.params.contextId || "");
      const fileKey = activeShareReaders.get(contextID);
      if (!fileKey) {
        throw new Error("Share context is missing");
      }
      const encryptedChunk = toUint8Array(message.params.encryptedChunk);
      try {
        const expectedHash = String(message.params.expectedHash || "");
        if (expectedHash) {
          const actualHash = crypto.hash_bytes_blake3(encryptedChunk);
          try {
            const hashEncoding = String(message.params.hashEncoding || "base64").toLowerCase();
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
          encryptedChunk,
          fileKey,
          aadBytes(message.params.aad),
        );
        try {
          return { chunkBytes: chunkBytes.slice() };
        } finally {
          crypto.zeroize(chunkBytes);
        }
      } finally {
        crypto.zeroize(encryptedChunk);
      }
    }
    case "encryptFileMetadataInContext": {
      const contextID = String(message.params.contextId || "");
      const fileKey = activeReaders.get(contextID);
      if (!fileKey) {
        throw new Error("File context is missing");
      }
      const metadata = new TextEncoder().encode(
        JSON.stringify(message.params.metadata || {}),
      );
      try {
        const encrypted = crypto.encrypt_chunk(
          metadata,
          fileKey,
          aadBytes(message.params.aad),
        );
        try {
          return { encryptedMetadata: encodeBase64(encrypted) };
        } finally {
          crypto.zeroize(encrypted);
        }
      } finally {
        crypto.zeroize(metadata);
      }
    }
    case "encryptFileMetadata":
    case "encryptFolderMetadata": {
      const master = activeMasterKey(message.params.masterKey);
      const metadata = new TextEncoder().encode(
        JSON.stringify(message.params.metadata || {}),
      );
      try {
        const encrypted = crypto.encrypt_chunk(
          metadata,
          master.bytes,
          aadBytes(message.params.aad),
        );
        try {
          return { encryptedMetadata: encodeBase64(encrypted) };
        } finally {
          crypto.zeroize(encrypted);
        }
      } finally {
        if (master.temporary) {
          crypto.zeroize(master.bytes);
        }
      }
    }
    case "decryptFileMetadata":
    case "decryptFolderMetadata": {
      const master = activeMasterKey(message.params.masterKey);
      const encryptedMetadata = decodeBase64(message.params.encryptedMetadata);
      try {
        const metadata = crypto.decrypt_chunk(
          encryptedMetadata,
          master.bytes,
          aadBytes(message.params.aad),
        );
        try {
          return { metadata: JSON.parse(new TextDecoder().decode(metadata)) };
        } finally {
          crypto.zeroize(metadata);
        }
      } finally {
        crypto.zeroize(encryptedMetadata);
        if (master.temporary) {
          crypto.zeroize(master.bytes);
        }
      }
    }
    case "encryptFileKey": {
      const master = activeMasterKey(message.params.masterKey);
      const fileKey = decodeBase64(message.params.fileKey);
      try {
        const encryptedFileKey = crypto.wrap_file_key(
          fileKey,
          master.bytes,
          aadBytes(message.params.aad),
        );
        try {
          return { encryptedFileKey: encodeBase64(encryptedFileKey) };
        } finally {
          crypto.zeroize(encryptedFileKey);
        }
      } finally {
        crypto.zeroize(fileKey);
        if (master.temporary) {
          crypto.zeroize(master.bytes);
        }
      }
    }
    case "encryptChunk": {
      const fileKey = decodeBase64(message.params.fileKey);
      const chunkBytes = decodeBase64(message.params.chunkBytes);
      try {
        const encryptedChunk = crypto.encrypt_chunk(
          chunkBytes,
          fileKey,
          aadBytes(message.params.aad),
        );
        try {
          return { encryptedChunk: encodeBase64(encryptedChunk) };
        } finally {
          crypto.zeroize(encryptedChunk);
        }
      } finally {
        crypto.zeroize(fileKey);
        crypto.zeroize(chunkBytes);
      }
    }
    case "decryptChunk": {
      const fileKey = decodeBase64(message.params.fileKey);
      const encryptedChunk = decodeBase64(message.params.encryptedChunk);
      try {
        const chunkBytes = crypto.decrypt_chunk(
          encryptedChunk,
          fileKey,
          aadBytes(message.params.aad),
        );
        try {
          return { chunkBytes: encodeBase64(chunkBytes) };
        } finally {
          crypto.zeroize(chunkBytes);
        }
      } finally {
        crypto.zeroize(fileKey);
        crypto.zeroize(encryptedChunk);
      }
    }
    case "createSearchTokens": {
      if (!unlockedMasterKey) {
        throw new Error("Vault is locked");
      }
      const vaultId = String(message.params.vaultId || "");
      const terms = Array.isArray(message.params.terms) ? message.params.terms : [];
      if (!vaultId || !terms.length) {
        return { tokens: [] };
      }
      const searchKey = await deriveSearchKey(unlockedMasterKey);
      try {
        const hmacKey = await globalThis.crypto.subtle.importKey(
          "raw",
          searchKey,
          { name: "HMAC", hash: "SHA-256" },
          false,
          ["sign"],
        );
        const tokens = [];
        for (let i = 0; i < terms.length; i += 1) {
          const payload = new TextEncoder().encode(vaultId + ":" + String(terms[i] || ""));
          const digest = new Uint8Array(await globalThis.crypto.subtle.sign("HMAC", hmacKey, payload));
          tokens.push(encodeBase64URL(digest));
        }
        return { tokens: tokens };
      } finally {
        searchKey.fill(0);
      }
    }
    case "decryptSearchFile": {
      if (!unlockedMasterKey) {
        throw new Error("Vault is locked");
      }
      const encryptedFileKey = decodeBase64(message.params.encryptedFileKey);
      const encryptedMetadata = decodeBase64(message.params.encryptedMetadata);
      try {
        const fileKey = crypto.unwrap_file_key(
          encryptedFileKey,
          unlockedMasterKey,
          aadBytes(message.params.fileKeyAad),
        );
        try {
          const metadata = crypto.decrypt_chunk(
            encryptedMetadata,
            fileKey,
            aadBytes(message.params.metadataAad),
          );
          try {
            return { metadata: JSON.parse(new TextDecoder().decode(metadata)) };
          } finally {
            crypto.zeroize(metadata);
          }
        } finally {
          crypto.zeroize(fileKey);
        }
      } finally {
        crypto.zeroize(encryptedFileKey);
        crypto.zeroize(encryptedMetadata);
      }
    }
    default:
      throw new Error("Unsupported vault method");
  }
}

self.addEventListener("message", function (event) {
  const payload = event.data || {};
  handleMessage(payload)
    .then(function (result) {
      const transfer = [];
      if (result && result.encryptedChunk instanceof Uint8Array) {
        transfer.push(result.encryptedChunk.buffer);
      }
      if (result && result.chunkBytes instanceof Uint8Array) {
        transfer.push(result.chunkBytes.buffer);
      }
      self.postMessage({
        id: payload.id,
        ok: true,
        result: result,
      }, transfer);
    })
    .catch(function (error) {
      self.postMessage({
        id: payload.id,
        ok: false,
        error: error && error.message ? error.message : "Vault worker failed",
      });
    });
});
