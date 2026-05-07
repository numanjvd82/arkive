import initArkiveCrypto, * as arkiveCrypto from "../vendor/arkive-crypto/arkive_crypto.js";

let readyPromise = null;
let unlockedMasterKey = null;
let activeUploads = new Map();

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
    crypto.zeroize(fileKey);
  });
  activeUploads.clear();
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
          unlockedMasterKey,
          aadBytes(message.params.metadataAad),
        );
        try {
          const encryptedFileKey = crypto.wrap_file_key(
            fileKey,
            unlockedMasterKey,
            aadBytes(message.params.fileKeyAad),
          );
          try {
            activeUploads.set(uploadToken, fileKey.slice());
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
      const fileKey = activeUploads.get(uploadToken);
      if (!fileKey) {
        throw new Error("Upload context is missing");
      }
      const chunkBytes = toUint8Array(message.params.chunkBytes);
      try {
        const encryptedChunk = crypto.encrypt_chunk(
          chunkBytes,
          fileKey,
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
      const fileKey = activeUploads.get(uploadToken);
      if (fileKey) {
        crypto.zeroize(fileKey);
        activeUploads.delete(uploadToken);
      }
      return { cleared: true };
    }
    case "encryptFileMetadata": {
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
