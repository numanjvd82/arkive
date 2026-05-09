const DB_NAME = "arkive_uploads";
const DB_VERSION = 1;

function openDB() {
	return new Promise(function (resolve, reject) {
		const request = indexedDB.open(DB_NAME, DB_VERSION);
		request.onupgradeneeded = function () {
			const db = request.result;
			if (!db.objectStoreNames.contains("upload_jobs")) {
				db.createObjectStore("upload_jobs", { keyPath: "jobId" });
			}
			if (!db.objectStoreNames.contains("upload_parts")) {
				const parts = db.createObjectStore("upload_parts", { keyPath: ["jobId", "partNumber"] });
				parts.createIndex("jobId", "jobId", { unique: false });
			}
			if (!db.objectStoreNames.contains("upload_queue")) {
				db.createObjectStore("upload_queue", { keyPath: "jobId" });
			}
		};
		request.onsuccess = function () { resolve(request.result); };
		request.onerror = function () { reject(request.error || new Error("Failed to open upload DB")); };
	});
}

function withStore(mode, storeName, fn) {
	return openDB().then(function (db) {
		return new Promise(function (resolve, reject) {
			const tx = db.transaction(storeName, mode);
			const store = tx.objectStore(storeName);
			let settled = false;
			let result;
			tx.oncomplete = function () {
				settled = true;
				resolve(result);
			};
			tx.onerror = function () {
				if (settled) return;
				reject(tx.error || new Error("IndexedDB transaction failed"));
			};
			tx.onabort = function () {
				if (settled) return;
				reject(tx.error || new Error("IndexedDB transaction aborted"));
			};
			Promise.resolve(fn(store, tx)).then(function (value) {
				result = value;
			}).catch(function (error) {
				settled = true;
				reject(error);
				try { tx.abort(); } catch (_) {}
			});
		});
	});
}

function getAll(store) {
	return new Promise(function (resolve, reject) {
		const request = store.getAll();
		request.onsuccess = function () { resolve(request.result || []); };
		request.onerror = function () { reject(request.error || new Error("IndexedDB read failed")); };
	});
}

function put(store, value) {
	return new Promise(function (resolve, reject) {
		const request = store.put(value);
		request.onsuccess = function () { resolve(value); };
		request.onerror = function () { reject(request.error || new Error("IndexedDB write failed")); };
	});
}

function del(store, key) {
	return new Promise(function (resolve, reject) {
		const request = store.delete(key);
		request.onsuccess = function () { resolve(); };
		request.onerror = function () { reject(request.error || new Error("IndexedDB delete failed")); };
	});
}

export function getJobs() {
	return withStore("readonly", "upload_jobs", function (store) { return getAll(store); });
}

export function putJob(job) {
	return withStore("readwrite", "upload_jobs", function (store) { return put(store, job); });
}

export function deleteJob(jobId) {
	return Promise.all([
		withStore("readwrite", "upload_jobs", function (store) { return del(store, jobId); }),
		withStore("readwrite", "upload_queue", function (store) { return del(store, jobId); }),
	]).then(function () {
		return withStore("readwrite", "upload_parts", function (store) {
			return new Promise(function (resolve, reject) {
				const request = store.openCursor();
				request.onsuccess = function () {
					const cursor = request.result;
					if (!cursor) {
						resolve();
						return;
					}
					if (cursor.value && cursor.value.jobId === jobId) {
						cursor.delete();
					}
					cursor.continue();
				};
				request.onerror = function () { reject(request.error || new Error("IndexedDB cursor failed")); };
			});
		});
	});
}

export function putPart(part) {
	return withStore("readwrite", "upload_parts", function (store) { return put(store, part); });
}

export function getParts(jobId) {
	return withStore("readonly", "upload_parts", function (store) {
		return new Promise(function (resolve, reject) {
			const request = store.index("jobId").getAll(jobId);
			request.onsuccess = function () {
				resolve((request.result || []).sort(function (a, b) { return a.partNumber - b.partNumber; }));
			};
			request.onerror = function () { reject(request.error || new Error("IndexedDB read failed")); };
		});
	});
}

export function putQueueItem(item) {
	return withStore("readwrite", "upload_queue", function (store) { return put(store, item); });
}
