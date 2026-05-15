const DB_NAME = "arkive-thumbnail-cache";
const STORE_NAME = "thumbnails";
const DB_VERSION = 1;
const MAX_CACHE_BYTES = 200 * 1024 * 1024;
const MAX_ENTRY_AGE_MS = 60 * 24 * 60 * 60 * 1000;

let openPromise = null;

function cacheKey(fileId, thumbnailVersion, thumbnailSize) {
  return [
    String(fileId || ""),
    String(Number(thumbnailVersion || 0)),
    String(Number(thumbnailSize || 0)),
  ].join(":");
}

function openDatabase() {
  if (openPromise) {
    return openPromise;
  }
  openPromise = new Promise(function(resolve, reject) {
    if (typeof indexedDB === "undefined") {
      resolve(null);
      return;
    }
    const request = indexedDB.open(DB_NAME, DB_VERSION);
    request.onupgradeneeded = function(event) {
      const db = event.target.result;
      const store = db.createObjectStore(STORE_NAME, { keyPath: "cacheKey" });
      store.createIndex("lastAccessedAt", "lastAccessedAt", { unique: false });
    };
    request.onsuccess = function() {
      resolve(request.result);
    };
    request.onerror = function() {
      reject(request.error || new Error("Thumbnail cache unavailable"));
    };
  }).catch(function() {
    return null;
  });
  return openPromise;
}

function requestResult(request) {
  return new Promise(function(resolve, reject) {
    request.onsuccess = function() {
      resolve(request.result);
    };
    request.onerror = function() {
      reject(request.error || new Error("Thumbnail cache request failed"));
    };
  });
}

function transactionDone(tx) {
  return new Promise(function(resolve, reject) {
    tx.oncomplete = function() {
      resolve();
    };
    tx.onabort = function() {
      reject(tx.error || new Error("Thumbnail cache transaction aborted"));
    };
    tx.onerror = function() {
      reject(tx.error || new Error("Thumbnail cache transaction failed"));
    };
  });
}

async function evictExpired(db, now) {
  const tx = db.transaction(STORE_NAME, "readwrite");
  const store = tx.objectStore(STORE_NAME);
  const records = await requestResult(store.getAll());
  for (let i = 0; i < records.length; i++) {
    const record = records[i];
    if (!record || now - Number(record.lastAccessedAt || 0) <= MAX_ENTRY_AGE_MS) {
      continue;
    }
    store.delete(record.cacheKey);
  }
  await transactionDone(tx);
}

async function evictToFit(db) {
  const tx = db.transaction(STORE_NAME, "readwrite");
  const store = tx.objectStore(STORE_NAME);
  const records = await requestResult(store.getAll());
  let totalBytes = 0;
  for (let i = 0; i < records.length; i++) {
    totalBytes += Number((records[i] && records[i].byteLength) || 0);
  }
  if (totalBytes <= MAX_CACHE_BYTES) {
    await transactionDone(tx);
    return;
  }
  records.sort(function(a, b) {
    return Number((a && a.lastAccessedAt) || 0) - Number((b && b.lastAccessedAt) || 0);
  });
  for (let i = 0; i < records.length && totalBytes > MAX_CACHE_BYTES; i++) {
    const record = records[i];
    if (!record) {
      continue;
    }
    store.delete(record.cacheKey);
    totalBytes -= Number(record.byteLength || 0);
  }
  await transactionDone(tx);
}

export const thumbnailCache = {
  async get(fileId, thumbnailVersion, thumbnailSize) {
    const db = await openDatabase();
    if (!db) {
      return null;
    }
    const key = cacheKey(fileId, thumbnailVersion, thumbnailSize);
    try {
      const tx = db.transaction(STORE_NAME, "readwrite");
      const store = tx.objectStore(STORE_NAME);
      const record = await requestResult(store.get(key));
      if (!record) {
        await transactionDone(tx);
        return null;
      }
      const now = Date.now();
      if (now - Number(record.lastAccessedAt || 0) > MAX_ENTRY_AGE_MS) {
        store.delete(key);
        await transactionDone(tx);
        return null;
      }
      record.lastAccessedAt = now;
      store.put(record);
      await transactionDone(tx);
      return record.bytes instanceof Uint8Array ? record.bytes : new Uint8Array(record.bytes || []);
    } catch (_) {
      return null;
    }
  },

  async put(fileId, thumbnailVersion, thumbnailSize, encryptedBytes) {
    const db = await openDatabase();
    if (!db || !(encryptedBytes instanceof Uint8Array) || !encryptedBytes.length) {
      return;
    }
    const now = Date.now();
    const key = cacheKey(fileId, thumbnailVersion, thumbnailSize);
    try {
      const tx = db.transaction(STORE_NAME, "readwrite");
      tx.objectStore(STORE_NAME).put({
        cacheKey: key,
        fileId: String(fileId || ""),
        thumbnailVersion: Number(thumbnailVersion || 0),
        thumbnailSize: Number(thumbnailSize || 0),
        byteLength: encryptedBytes.length,
        lastAccessedAt: now,
        bytes: encryptedBytes.slice(),
      });
      await transactionDone(tx);
      await evictExpired(db, now);
      await evictToFit(db);
    } catch (_) {
    }
  },
};
