import initArkiveCrypto, * as arkiveCrypto from "../vendor/arkive-crypto/arkive_crypto.js";
import { UPLOAD_POLICY, getPartConcurrency } from "../lib/upload_policy.js";
import { UPLOAD_MESSAGE } from "../lib/upload_protocol.js";
import { deleteJob, getJobs, getParts, putJob, putPart } from "../lib/upload_db.js";

const ports = new Set();
const jobs = new Map();
let activeJobId = null;
let vaultSession = null;
let cryptoReady = null;
let unlockedMasterKey = null;
const activeUploads = new Map();
const activeControllers = new Map();
const activeCancelCalls = new Map();
const STOPPED_UPLOAD = new Error("Upload stopped");

function logUpload() {
	if (typeof console !== "undefined" && console.debug) {
		console.debug.apply(console, ["[arkive-uploads]"].concat(Array.prototype.slice.call(arguments)));
	}
}

function logUploadError() {
	if (typeof console !== "undefined" && console.error) {
		console.error.apply(console, ["[arkive-uploads]"].concat(Array.prototype.slice.call(arguments)));
	}
}

function nowISO() {
	return new Date().toISOString();
}

function uuid() {
	return self.crypto && self.crypto.randomUUID ? self.crypto.randomUUID() : "job-" + Date.now() + "-" + Math.random().toString(36).slice(2);
}

function fileIdentity(file) {
	return [String(file.name || ""), String(file.size || 0), String(file.lastModified || 0)].join("|");
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
	if (value instanceof Uint8Array) return value;
	if (value instanceof ArrayBuffer) return new Uint8Array(value);
	if (ArrayBuffer.isView(value)) {
		return new Uint8Array(value.buffer, value.byteOffset, value.byteLength);
	}
	return decodeBase64(value);
}

function aadBytes(value) {
	if (!value) return new Uint8Array();
	if (value instanceof Uint8Array) return value;
	return new TextEncoder().encode(String(value));
}

function ensureCrypto() {
	if (cryptoReady) return cryptoReady;
	cryptoReady = initArkiveCrypto({ module_or_path: "/static/vendor/arkive-crypto/arkive_crypto_bg.wasm" }).then(function () {
		return arkiveCrypto;
	});
	return cryptoReady;
}

function lockVault(crypto) {
	if (unlockedMasterKey) {
		crypto.zeroize(unlockedMasterKey);
		unlockedMasterKey = null;
	}
	activeUploads.forEach(function (fileKey) {
		if (fileKey && fileKey.fileKey) {
			crypto.zeroize(fileKey.fileKey);
		}
	});
	activeUploads.clear();
}

function activeMasterKey(supplied) {
	if (supplied) {
		return { bytes: decodeBase64(supplied), temporary: true };
	}
	if (!unlockedMasterKey) {
		throw new Error("Vault is locked");
	}
	return { bytes: unlockedMasterKey, temporary: false };
}

function clearJobControllers(jobId) {
	for (const [key, controller] of activeControllers.entries()) {
		if (key.indexOf(jobId + ":") === 0) {
			controller.abort();
			activeControllers.delete(key);
		}
	}
}

logUpload("worker loaded");

function broadcast(message) {
	ports.forEach(function (port) {
		port.postMessage(message);
	});
}

function summarizeJobs() {
	return {
		jobs: Array.from(jobs.values()).map(function (job) { return job.public; }),
		activeJobId: activeJobId,
		incompleteJobs: Array.from(jobs.values()).filter(function (job) {
			return job.public.status !== "completed" && job.public.status !== "canceled";
		}).map(function (job) { return job.public; }),
	};
}

function broadcastState() {
	broadcast({ type: UPLOAD_MESSAGE.STATE, state: summarizeJobs() });
}

async function cryptoCall(method, params, transfer) {
	const crypto = await ensureCrypto();
	switch (method) {
		case "restoreSessionUnlock": {
			const sessionUnwrapKey = toUint8Array(params.sessionUnwrapKey);
			const wrappedMasterKey = decodeBase64(params.wrappedMasterKey);
			try {
				const masterKey = crypto.unwrap_master_key(wrappedMasterKey, sessionUnwrapKey, aadBytes(params.aad));
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
		case "prepareUpload": {
			const uploadToken = String(params.uploadToken || "");
			if (!uploadToken) throw new Error("Missing upload token");
			if (!unlockedMasterKey) throw new Error("Vault is locked");
			const metadata = new TextEncoder().encode(JSON.stringify(params.metadata || {}));
			const fileKey = crypto.generate_file_key();
			try {
				const encryptedMetadata = crypto.encrypt_chunk(metadata, fileKey, aadBytes(params.metadataAad));
				try {
					const encryptedFileKey = crypto.wrap_file_key(fileKey, unlockedMasterKey, aadBytes(params.fileKeyAad));
					try {
						activeUploads.set(uploadToken, { fileKey: fileKey.slice() });
						return { encryptedMetadata: encodeBase64(encryptedMetadata), encryptedFileKey: encodeBase64(encryptedFileKey), totalParts: Number(params.totalParts || 0), encryptionVersion: 1 };
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
			const uploadToken = String(params.uploadToken || "");
			const upload = activeUploads.get(uploadToken);
			if (!upload || !upload.fileKey) throw new Error("Upload context is missing");
			const chunkBytes = toUint8Array(params.chunkBytes);
			try {
				const encryptedChunk = crypto.encrypt_chunk(chunkBytes, upload.fileKey, aadBytes(params.aad));
				try {
					const encryptedHash = crypto.hash_bytes_blake3(encryptedChunk);
					try {
						return { encryptedChunk: encryptedChunk.slice(), encryptedHash: encodeBase64(encryptedHash), encryptedSize: encryptedChunk.length };
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
			const uploadToken = String(params.uploadToken || "");
			const upload = activeUploads.get(uploadToken);
			if (!upload || !upload.fileKey) throw new Error("Upload context is missing");
			const manifest = new TextEncoder().encode(JSON.stringify(params.manifest || {}));
			let encryptedHash = new Uint8Array();
			try {
				const encryptedManifest = crypto.encrypt_chunk(manifest, upload.fileKey, aadBytes(params.manifestAad));
				try {
					const partHashes = params.partHashes || [];
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
					return { encryptedManifest: encodeBase64(encryptedManifest), encryptedHash: encodeBase64(encryptedHash) };
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
			const uploadToken = String(params.uploadToken || "");
			const upload = activeUploads.get(uploadToken);
			if (upload && upload.fileKey) {
				crypto.zeroize(upload.fileKey);
				activeUploads.delete(uploadToken);
			}
			return { cleared: true };
		}
		case "lockVault": {
			lockVault(crypto);
			return { unlocked: false };
		}
		case "generateFileKey": {
			const fileKey = crypto.generate_file_key();
			try {
				return { fileKey: encodeBase64(fileKey) };
			} finally {
				crypto.zeroize(fileKey);
			}
		}
		default:
			throw new Error("Unsupported vault method");
	}
}

async function restoreVault() {
	if (!vaultSession) return;
	await cryptoCall("restoreSessionUnlock", {
		sessionUnwrapKey: vaultSession.sessionUnwrapKey,
		wrappedMasterKey: vaultSession.wrappedMasterKey,
		aad: "arkive:session-master-key:v1",
	});
}

async function addFiles(files) {
	logUpload("addFiles", { count: files.length });
	for (let i = 0; i < files.length; i++) {
		const file = files[i];
		const jobId = uuid();
		const chunkSize = UPLOAD_POLICY.defaultPartSize;
		const job = {
			public: {
				jobId: jobId,
				fileId: "",
				vaultId: "",
				sessionId: "",
				fileName: file.name,
				fileSize: file.size,
				chunkSize: chunkSize,
				totalParts: Math.max(1, Math.ceil(file.size / chunkSize)),
				completedParts: 0,
				status: "queued",
				createdAt: nowISO(),
				updatedAt: nowISO(),
			},
			file: file,
			partConcurrency: getPartConcurrency(file.size),
			state: "queued",
		};
		jobs.set(jobId, job);
		await putJob(job.public);
	}
	broadcastState();
	processQueue();
}

async function resumeFiles(files) {
	logUpload("resumeFiles", { count: files.length });
	const requested = new Map();
	for (let i = 0; i < files.length; i++) {
		requested.set(fileIdentity(files[i]), files[i]);
	}
	const matched = [];
	for (const job of jobs.values()) {
		if (job.public.status === "completed" || job.public.status === "canceled") {
			continue;
		}
		const key = fileIdentity(job.file || { name: job.public.fileName, size: job.public.fileSize, lastModified: job.public.lastModified || 0 });
		const file = requested.get(key);
		if (!file) {
			continue;
		}
		job.file = file;
		job.public.status = job.public.completedParts > 0 ? "paused" : "queued";
		job.public.updatedAt = nowISO();
		const parts = await getParts(job.public.jobId);
		job.public.completedParts = parts.length;
		await putJob(job.public);
		logUpload("resume match", { jobId: job.public.jobId, fileName: job.public.fileName, completedParts: job.public.completedParts });
		matched.push(job.public.fileName);
	}
	broadcastState();
	processQueue();
	return { matched: matched };
}

async function startJob(job) {
	logUpload("startJob", { jobId: job.public.jobId, fileName: job.public.fileName, completedParts: job.public.completedParts, totalParts: job.public.totalParts });
	job.state = "running";
	job.public.status = "running";
	job.public.updatedAt = nowISO();
	await putJob(job.public);
	broadcastState();

	let started = null;
	let completed = false;
	try {
		await restoreVault();

		if (job.public.sessionId && job.public.fileId && job.public.vaultId) {
			logUpload("reuse session", { jobId: job.public.jobId, sessionId: job.public.sessionId });
			started = {
				uploadSessionId: job.public.sessionId,
				fileId: job.public.fileId,
				vaultId: job.public.vaultId,
			};
		} else {
			logUpload("create session", { jobId: job.public.jobId, size: job.file.size, partSize: job.public.chunkSize, totalParts: job.public.totalParts });
			started = await fetch("/api/uploads/start", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({
					originalSize: job.file.size,
					partSize: job.public.chunkSize,
					totalParts: job.public.totalParts,
					encryptionVersion: 1,
				}),
			}).then(async function (res) {
				const data = await res.json();
				if (!res.ok) throw new Error((data && data.error) || "Upload start failed");
				return data;
			});
			logUpload("session created", { jobId: job.public.jobId, sessionId: started.uploadSessionId, fileId: started.fileId });
			job.public.fileId = started.fileId;
			job.public.sessionId = started.uploadSessionId;
			job.public.vaultId = started.vaultId;
			job.public.updatedAt = nowISO();
			await putJob(job.public);
		}

		const prepare = await cryptoCall("prepareUpload", {
			uploadToken: job.public.jobId,
			vaultId: started.vaultId,
			fileId: started.fileId,
			metadata: {
				schema: "arkive.file.metadata",
				version: 1,
				name: job.file.name,
				mime: job.file.type || "application/octet-stream",
				extension: (job.file.name.split(".").pop() || "").toLowerCase(),
				size: job.file.size,
				created_at_client: nowISO(),
				modified_at_client: job.file.lastModified ? new Date(job.file.lastModified).toISOString() : null,
				preview: { thumbnail_file_id: null, has_thumbnail: false },
			},
			totalParts: job.public.totalParts,
			metadataAad: "arkive:file-metadata:v1:" + started.vaultId + ":" + started.fileId,
			fileKeyAad: "arkive:file-key:v1:" + started.vaultId + ":" + started.fileId,
		});

		let nextPart = 1;
		let active = 0;
		let error = null;
		const existingParts = await getParts(job.public.jobId);
		const doneParts = new Set(existingParts.map(function (part) { return part.partNumber; }));
		job.public.completedParts = doneParts.size;
		await putJob(job.public).catch(function () {});
		logUpload("part resume state", { jobId: job.public.jobId, doneParts: doneParts.size });
		await new Promise(function (resolve, reject) {
			function launch() {
				if (error) {
					reject(error);
					return;
				}
				if (nextPart > job.public.totalParts && active === 0) {
					resolve();
					return;
				}
				while (active < job.partConcurrency && nextPart <= job.public.totalParts) {
					const partNumber = nextPart++;
					if (doneParts.has(partNumber)) {
						logUpload("skip completed part", { jobId: job.public.jobId, partNumber: partNumber });
						continue;
					}
					active++;
					logUpload("upload part", { jobId: job.public.jobId, partNumber: partNumber });
					uploadPart(job, started, prepare, partNumber).then(function () {
						active--;
						launch();
					}).catch(function (err) {
						active--;
						logUploadError("part failed", { jobId: job.public.jobId, partNumber: partNumber, error: String(err && err.message ? err.message : err || "upload part failed") });
						if (err === STOPPED_UPLOAD || (err && (err.name === "AbortError" || err.message === STOPPED_UPLOAD.message)) || job.public.status === "paused" || job.public.status === "canceled") {
							error = STOPPED_UPLOAD;
						} else {
							error = err;
						}
						launch();
					});
				}
			}
			launch();
		});
		if (error === STOPPED_UPLOAD) {
			return;
		}

		const completedParts = await getParts(job.public.jobId);
		logUpload("finalize job", { jobId: job.public.jobId, uploadedParts: completedParts.length });
		const manifest = {
			schema: "arkive.file.manifest",
			version: 1,
			file_id: started.fileId,
			name: job.file.name,
			mime: job.file.type || "application/octet-stream",
			extension: (job.file.name.split(".").pop() || "").toLowerCase(),
			size: job.file.size,
			chunk_size: job.public.chunkSize,
			chunks: completedParts.map(function (part) {
				return { n: part.partNumber, offset: part.offset, plain_size: part.size, cipher_size: part.encryptedSize, hash: part.encryptedHash };
			}),
		};
		const finalized = await cryptoCall("finalizeUpload", {
			uploadToken: job.public.jobId,
			vaultId: started.vaultId,
			fileId: started.fileId,
			manifest: manifest,
			manifestAad: "arkive:file-manifest:v1:" + started.vaultId + ":" + started.fileId,
			partHashes: completedParts.map(function (part) { return part.encryptedHash; }),
		});
		await fetch("/api/uploads/" + encodeURIComponent(started.uploadSessionId) + "/complete", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({
				encryptedMetadata: prepare.encryptedMetadata,
				encryptedFileKey: prepare.encryptedFileKey,
				encryptedManifest: finalized.encryptedManifest,
				encryptedHash: finalized.encryptedHash,
			}),
		});
		logUpload("job completed", { jobId: job.public.jobId });
		job.public.status = "completed";
		job.public.updatedAt = nowISO();
		await putJob(job.public);
		await deleteJob(job.public.jobId);
		jobs.delete(job.public.jobId);
		broadcast({ type: UPLOAD_MESSAGE.JOB_REMOVED, jobId: job.public.jobId });
		broadcastState();
		completed = true;
	} finally {
		if (!completed) {
			logUpload("clear upload context after failure", { jobId: job.public.jobId });
			await cryptoCall("clearUploadContext", { uploadToken: job.public.jobId }).catch(function () {});
		}
	}
}

async function uploadPart(job, started, prepare, partNumber) {
	const start = (partNumber - 1) * job.public.chunkSize;
	const end = Math.min(start + job.public.chunkSize, job.file.size);
	const chunk = new Uint8Array(await job.file.slice(start, end).arrayBuffer());
	const encrypted = await cryptoCall("encryptUploadPart", {
		uploadToken: job.public.jobId,
		chunkBytes: chunk,
		aad: "arkive:file-chunk:v1:" + started.vaultId + ":" + started.fileId + ":" + partNumber + ":" + job.public.chunkSize + ":" + job.public.totalParts,
	}, [chunk.buffer]);
	const presigned = await fetch("/api/uploads/" + encodeURIComponent(started.uploadSessionId) + "/parts/" + encodeURIComponent(String(partNumber)) + "/presign", { method: "POST" }).then(async function (res) {
		const data = await res.json();
		if (!res.ok) throw new Error((data && data.error) || "Part presign failed");
		return data;
	});
	const controller = new AbortController();
	activeControllers.set(job.public.jobId + ":" + partNumber, controller);
	try {
		logUpload("presign part", { jobId: job.public.jobId, partNumber: partNumber });
		const response = await fetch(presigned.url, { method: "PUT", body: encrypted.encryptedChunk, signal: controller.signal });
		if (controller.signal.aborted) throw STOPPED_UPLOAD;
		if (!response.ok) throw new Error("Part upload failed");
		const etag = response.headers.get("etag") || response.headers.get("ETag") || "";
		logUpload("record part", { jobId: job.public.jobId, partNumber: partNumber, etag: etag });
		await fetch("/api/uploads/" + encodeURIComponent(started.uploadSessionId) + "/parts", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ partNumber: partNumber, encryptedSize: encrypted.encryptedSize, encryptedHash: encrypted.encryptedHash, etag: etag }),
		});
		await putPart({ jobId: job.public.jobId, partNumber: partNumber, offset: start, size: end - start, etag: etag, encryptedSize: encrypted.encryptedSize, encryptedHash: encrypted.encryptedHash, status: "done" });
		job.public.completedParts += 1;
		job.public.updatedAt = nowISO();
		await putJob(job.public);
		broadcastState();
	} finally {
		activeControllers.delete(job.public.jobId + ":" + partNumber);
	}
}

function selectNextJob() {
	if (activeJobId) return null;
	for (const job of jobs.values()) {
		if (job.public.status === "queued" || job.public.status === "paused") return job;
	}
	return null;
}

function processQueue() {
	const next = selectNextJob();
	if (!next) return;
	activeJobId = next.public.jobId;
	startJob(next).catch(function (error) {
		if (error === STOPPED_UPLOAD || (error && (error.name === "AbortError" || error.message === STOPPED_UPLOAD.message))) {
			return;
		}
		logUploadError("job failed", { jobId: next.public.jobId, error: String(error && error.message ? error.message : error || "Upload failed") });
		next.public.status = "failed";
		next.public.updatedAt = nowISO();
		putJob(next.public).catch(function () {});
		broadcast({ type: UPLOAD_MESSAGE.ERROR, jobId: next.public.jobId, error: String(error && error.message ? error.message : error || "Upload failed") });
	}).finally(function () {
		activeJobId = null;
		broadcastState();
		processQueue();
	});
}

async function hydrate() {
	const stored = await getJobs();
	for (let i = 0; i < stored.length; i++) {
		const job = stored[i];
		if (!job || !job.jobId) continue;
		if (job.status === "running") job.status = "paused";
		jobs.set(job.jobId, { public: job, file: null, partConcurrency: getPartConcurrency(job.fileSize) });
	}
	broadcastState();
}

async function cancelJob(jobId) {
	const job = jobs.get(jobId);
	if (!job) return;
	clearJobControllers(jobId);
	logUpload("cancel job", { jobId: jobId });
	job.public.status = "canceled";
	job.public.updatedAt = nowISO();
	if (!activeCancelCalls.has(jobId)) {
		const cancelCall = fetch("/api/uploads/" + encodeURIComponent(job.public.sessionId) + "/cancel", { method: "POST" }).catch(function () {});
		activeCancelCalls.set(jobId, cancelCall);
		await cancelCall;
		activeCancelCalls.delete(jobId);
	}
	await deleteJob(jobId).catch(function () {});
	jobs.delete(jobId);
	if (activeJobId === jobId) activeJobId = null;
	broadcastState();
}

async function removeJob(jobId) {
	const job = jobs.get(jobId);
	if (!job) {
		return;
	}
	clearJobControllers(jobId);
	logUpload("remove job", { jobId: jobId });
	await deleteJob(jobId).catch(function () {});
	jobs.delete(jobId);
	if (activeJobId === jobId) {
		activeJobId = null;
	}
	broadcastState();
}

async function pauseJob(jobId) {
	const job = jobs.get(jobId);
	if (!job) return;
	clearJobControllers(jobId);
	logUpload("pause job", { jobId: jobId });
	job.public.status = "paused";
	job.public.updatedAt = nowISO();
	await putJob(job.public).catch(function () {});
	broadcastState();
}

async function resumeJob(jobId) {
	const job = jobs.get(jobId);
	if (!job) return;
	job.public.status = "queued";
	job.public.updatedAt = nowISO();
	logUpload("resume job", { jobId: jobId });
	await putJob(job.public).catch(function () {});
	broadcastState();
	processQueue();
}

async function cancelAll() {
	await Promise.all(Array.from(jobs.keys()).map(function (jobId) { return cancelJob(jobId); }));
}

async function handleMessage(port, message) {
	try {
		logUpload("handleMessage", { type: message.type, id: message.id });
		switch (message.type) {
			case UPLOAD_MESSAGE.CONNECT:
				await hydrate();
				port.postMessage({ type: UPLOAD_MESSAGE.STATE, state: summarizeJobs() });
				break;
			case UPLOAD_MESSAGE.REQUEST_STATE:
				port.postMessage({ id: message.id, ok: true, result: summarizeJobs() });
				break;
			case UPLOAD_MESSAGE.VAULT_SESSION:
				vaultSession = message.session || null;
				await restoreVault();
				port.postMessage({ id: message.id, ok: true, result: { ok: true } });
				break;
			case UPLOAD_MESSAGE.ADD_FILES:
				await addFiles(message.payload.files || []);
				port.postMessage({ id: message.id, ok: true, result: { ok: true } });
				break;
			case UPLOAD_MESSAGE.RESUME_FILES:
				await resumeFiles(message.payload.files || []);
				port.postMessage({ id: message.id, ok: true, result: { ok: true } });
				break;
			case UPLOAD_MESSAGE.CANCEL_JOB:
				await cancelJob(String(message.payload.jobId || ""));
				port.postMessage({ id: message.id, ok: true, result: { ok: true } });
				break;
			case UPLOAD_MESSAGE.PAUSE_JOB:
				await pauseJob(String(message.payload.jobId || ""));
				port.postMessage({ id: message.id, ok: true, result: { ok: true } });
				break;
			case UPLOAD_MESSAGE.RESUME_JOB:
				await resumeJob(String(message.payload.jobId || ""));
				port.postMessage({ id: message.id, ok: true, result: { ok: true } });
				break;
			case UPLOAD_MESSAGE.RESUME_ALL:
				Array.from(jobs.keys()).forEach(function (jobId) { resumeJob(jobId); });
				port.postMessage({ id: message.id, ok: true, result: { ok: true } });
				break;
			case UPLOAD_MESSAGE.CANCEL_ALL:
				await cancelAll();
				port.postMessage({ id: message.id, ok: true, result: { ok: true } });
				break;
			case UPLOAD_MESSAGE.REMOVE_JOB:
				await removeJob(String(message.payload.jobId || ""));
				port.postMessage({ id: message.id, ok: true, result: { ok: true } });
				break;
			default:
				throw new Error("Unknown upload message: " + message.type);
		}
	} catch (error) {
		logUploadError("worker message failed", { type: message.type, error: String(error && error.message ? error.message : error || "Upload worker failed") });
		if (typeof message.id === "number") {
			port.postMessage({ id: message.id, ok: false, error: String(error && error.message ? error.message : error || "Upload worker failed") });
		}
		broadcast({ type: UPLOAD_MESSAGE.ERROR, error: String(error && error.message ? error.message : error || "Upload worker failed") });
	}
}

self.onconnect = function (event) {
	const port = event.ports[0];
	ports.add(port);
	port.start();
	port.onmessage = function (e) {
		handleMessage(port, e.data || {});
	};
	port.postMessage({ type: UPLOAD_MESSAGE.STATE, state: summarizeJobs() });
};
