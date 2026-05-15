import { UPLOAD_POLICY } from "../upload/upload_policy.js";
import { STATUS, isTerminal } from "../upload/upload_state.js";
import { startUpload, presignUploadPart, presignUploadParts, presignThumbnailUpload, recordUploadPart, completeUpload, cancelUpload, cancelUploadBestEffort } from "../upload/upload_api.js";
import { setVaultSession, restoreVaultSession, prepareUpload, encryptUploadMetadata, encryptUploadThumbnail, encryptUploadPart, hashUploadPayload, finalizeUpload, clearUploadContext } from "../upload/upload_crypto.js";
import { generateUploadThumbnail } from "../upload/upload_thumbnail.js";

const STOPPED_UPLOAD = new Error("Upload stopped");
const THUMBNAIL_TIMEOUT_MS = 15000;

class PresignCache {
	constructor(options) {
		options = options || {};
		this.sessionId = String(options.sessionId || "");
		this.batchSize = Math.max(1, Number(options.batchSize || 1));
		this.fetcher = typeof options.fetcher === "function" ? options.fetcher : null;
		this.cache = new Map();
		this.inFlight = new Map();
	}

	async get(partNumber) {
		const key = Number(partNumber || 0);
		if (this.cache.has(key)) {
			const url = this.cache.get(key);
			this.cache.delete(key);
			return url;
		}
		if (this.inFlight.has(key)) {
			return this.inFlight.get(key);
		}

		const promise = this.fetchBatch(key)
			.then(() => {
				if (!this.cache.has(key)) {
					throw new Error("Part presign failed");
				}
				const url = this.cache.get(key);
				this.cache.delete(key);
				return url;
			})
			.finally(() => {
				this.inFlight.delete(key);
			});

		this.inFlight.set(key, promise);
		return promise;
	}

	async fetchBatch(partNumber) {
		if (!this.fetcher) {
			throw new Error("Part presign failed");
		}

		const parts = [];
		for (let i = 0; i < this.batchSize; i++) {
			parts.push(partNumber + i);
		}

		const response = await this.fetcher(parts);
		const urls = response && response.urls ? response.urls : {};
		for (const [key, value] of Object.entries(urls)) {
			this.cache.set(Number(key), value);
		}
	}
}

function nowISO() {
	return new Date().toISOString();
}

function uuid() {
	return self.crypto && self.crypto.randomUUID ? self.crypto.randomUUID() : "job-" + Date.now() + "-" + Math.random().toString(36).slice(2);
}

function getUploadPartConcurrency(defaultConcurrency) {
	const base = Math.max(1, Number(defaultConcurrency || 1));
	const nav = typeof navigator !== "undefined" ? navigator : null;
	const ua = nav && nav.userAgent ? nav.userAgent : "";
	const platform = nav && nav.platform ? nav.platform : "";
	const maxTouchPoints = nav && nav.maxTouchPoints ? nav.maxTouchPoints : 0;
	const isIOS = /iPad|iPhone|iPod/.test(ua) || (platform === "MacIntel" && maxTouchPoints > 1);
	const isAndroid = /android/i.test(ua);
	if (!isIOS && !isAndroid) {
		return base;
	}
	return Math.min(base, 2);
}

function joinUint8Arrays(parts) {
	let total = 0;
	for (let i = 0; i < parts.length; i++) {
		total += Number(parts[i] ? parts[i].length : 0);
	}
	const output = new Uint8Array(total);
	let offset = 0;
	for (let i = 0; i < parts.length; i++) {
		const part = parts[i];
		if (!part) {
			continue;
		}
		output.set(part, offset);
		offset += part.length;
	}
	return output;
}

function flattenChunkGroups(groups) {
	const output = [];
	for (let i = 0; i < groups.length; i++) {
		const group = groups[i];
		if (!Array.isArray(group)) {
			continue;
		}
		for (let j = 0; j < group.length; j++) {
			output.push(group[j]);
		}
	}
	return output;
}

function previewMetadataFromThumbnail(thumbnail) {
	if (!thumbnail) {
		return { thumbnail_file_id: null, has_thumbnail: false };
	}
	return {
		thumbnail_file_id: null,
		has_thumbnail: true,
		thumbnail_mime: thumbnail.mime,
		thumbnail_width: thumbnail.width,
		thumbnail_height: thumbnail.height,
		thumbnail_size: thumbnail.encryptedSize,
		thumbnail_version: 1,
	};
}

function waitForThumbnailResult(promise, timeoutMs) {
	const timeout = Math.max(1, Number(timeoutMs || THUMBNAIL_TIMEOUT_MS));
	return Promise.race([
		Promise.resolve(promise),
		new Promise(function(resolve) {
			setTimeout(function() {
				resolve(null);
			}, timeout);
		}),
	]);
}

export class UploadRunner {
	constructor(options) {
		options = options || {};
		this.limits = options.limits || { maxQueueItems: 300 };
		this.jobs = new Map();
		this.jobCleanupTimers = new Map();
		this.activeJobId = null;
		this.currentControllers = new Set();
		this.stateHandlers = [];
		this.eventHandlers = [];
		this.completedBatches = new Set();
	}

	onState(handler) {
		if (typeof handler === "function") this.stateHandlers.push(handler);
	}

	onEvent(handler) {
		if (typeof handler === "function") this.eventHandlers.push(handler);
	}

	setVaultSession(session) {
		setVaultSession(session || null);
	}

	getState() {
		return {
			jobs: Array.from(this.jobs.values()).map((job) => job.public),
			activeJobId: this.activeJobId,
		};
	}

	emitState() {
		const snapshot = this.getState();
		this.stateHandlers.forEach(function (handler) { handler(snapshot); });
	}

	emitEvent(event) {
		this.eventHandlers.forEach(function (handler) { handler(event); });
	}

	releaseJobFile(job) {
		if (job) {
			job.file = null;
		}
	}

	clearJobCleanup(jobId) {
		const timer = this.jobCleanupTimers.get(jobId);
		if (!timer) {
			return;
		}
		clearTimeout(timer);
		this.jobCleanupTimers.delete(jobId);
	}

	scheduleTerminalCleanup(job) {
		if (!job || !job.public || !job.public.jobId) {
			return;
		}
		const jobId = job.public.jobId;
		this.clearJobCleanup(jobId);
		const runner = this;
		const timer = setTimeout(function() {
			runner.jobCleanupTimers.delete(jobId);
			const current = runner.jobs.get(jobId);
			if (!current || !isTerminal(current.public.status)) {
				return;
			}
			runner.releaseJobFile(current);
			runner.jobs.delete(jobId);
			runner.emitState();
		}, 5000);
		this.jobCleanupTimers.set(jobId, timer);
	}

	hasActiveUploads() {
		for (const job of this.jobs.values()) {
			if (job.public.status === STATUS.QUEUED || job.public.status === STATUS.RUNNING) {
				return true;
			}
		}
		return false;
	}

	async addFiles(files) {
		const nextQueueSize = this.activeQueueCount() + files.length;
		if (this.limits.maxQueueItems > 0 && nextQueueSize > this.limits.maxQueueItems) {
			throw new Error("Upload queue limit reached (" + this.limits.maxQueueItems + " max items)");
		}
		const batchId = uuid();
		for (let i = 0; i < files.length; i++) {
			const file = files[i];
			const jobId = uuid();
			const fileChunkSize = UPLOAD_POLICY.fileChunkSize;
			const uploadPartSize = UPLOAD_POLICY.uploadPartSize;
			this.jobs.set(jobId, {
				file: file,
				public: {
					jobId: jobId,
					batchId: batchId,
					fileId: "",
					vaultId: "",
					sessionId: "",
					fileName: file.name,
					fileSize: file.size,
					fileChunkSize: fileChunkSize,
					totalChunks: Math.max(1, Math.ceil(file.size / fileChunkSize)),
					uploadPartSize: uploadPartSize,
					uploadPartCount: Math.max(1, Math.ceil(file.size / uploadPartSize)),
					completedUploadParts: 0,
					totalParts: Math.max(1, Math.ceil(file.size / uploadPartSize)),
					completedParts: 0,
					completedBytes: 0,
					transferRate: 0,
					startedAt: "",
					status: STATUS.QUEUED,
					createdAt: nowISO(),
					updatedAt: nowISO(),
				},
			});
		}
		this.emitState();
		void this.processQueue();
	}

	activeQueueCount() {
		let total = 0;
		for (const job of this.jobs.values()) {
			if (!isTerminal(job.public.status)) {
				total++;
			}
		}
		return total;
	}

	nextQueuedJob() {
		if (this.activeJobId) return null;
		for (const job of this.jobs.values()) {
			if (job.public.status === STATUS.QUEUED) {
				return job;
			}
		}
		return null;
	}

	processQueue() {
		const next = this.nextQueuedJob();
		if (!next) return;
		this.activeJobId = next.public.jobId;
		this.startJob(next).catch((error) => {
			if (error === STOPPED_UPLOAD || (error && error.message === STOPPED_UPLOAD.message)) {
				return;
			}
			return this.failJob(next, error);
		}).finally(() => {
			this.activeJobId = null;
			this.emitState();
			this.processQueue();
		});
	}

	async startJob(job) {
		job.public.status = STATUS.RUNNING;
		job.public.startedAt = job.public.startedAt || nowISO();
		job.public.updatedAt = nowISO();
		this.emitState();

		await restoreVaultSession();

		const started = await this.runWithController(async (signal) => {
				return startUpload({
					originalSize: job.file.size,
					fileChunkSize: job.public.fileChunkSize,
					totalChunks: job.public.totalChunks,
					uploadPartSize: job.public.uploadPartSize,
					uploadPartCount: job.public.uploadPartCount,
					encryptionVersion: 1,
				}, signal);
			});
		if (this.shouldStopJob(job)) {
			throw STOPPED_UPLOAD;
		}

		job.public.fileId = started.fileId;
		job.public.sessionId = started.uploadSessionId;
		job.public.vaultId = started.vaultId;
		job.public.updatedAt = nowISO();
		this.emitState();

		let completed = false;
		try {
			const prepared = await prepareUpload({
				uploadToken: job.public.jobId,
				vaultId: started.vaultId,
				fileId: started.fileId,
				totalParts: job.public.totalChunks,
				fileKeyAad: "arkive:file-key:v1:" + started.vaultId + ":" + started.fileId,
			});
			if (this.shouldStopJob(job)) {
				throw STOPPED_UPLOAD;
			}
			const thumbnailPromise = this.createAndUploadThumbnail(job, started);

			const chunks = await this.uploadPartsPooled(job, started, {
				concurrency: getUploadPartConcurrency(UPLOAD_POLICY.partConcurrency),
				presignBatchSize: UPLOAD_POLICY.presignBatchSize,
			});
			if (this.shouldStopJob(job)) {
				throw STOPPED_UPLOAD;
			}
			const uploadedThumbnail = await waitForThumbnailResult(thumbnailPromise, THUMBNAIL_TIMEOUT_MS);
			if (this.shouldStopJob(job)) {
				throw STOPPED_UPLOAD;
			}
			const preparedMetadata = await encryptUploadMetadata({
				uploadToken: job.public.jobId,
				metadata: {
					schema: "arkive.file.metadata",
					version: 1,
					name: job.file.name,
					mime: job.file.type || "application/octet-stream",
					extension: (job.file.name.split(".").pop() || "").toLowerCase(),
					size: job.file.size,
					created_at_client: nowISO(),
					modified_at_client: job.file.lastModified ? new Date(job.file.lastModified).toISOString() : null,
					preview: previewMetadataFromThumbnail(uploadedThumbnail),
				},
				metadataAad: "arkive:file-metadata:v1:" + started.vaultId + ":" + started.fileId,
			});
			if (this.shouldStopJob(job)) {
				throw STOPPED_UPLOAD;
			}

			const finalized = await finalizeUpload({
				uploadToken: job.public.jobId,
				vaultId: started.vaultId,
				fileId: started.fileId,
				manifest: {
					schema: "arkive.file.manifest",
					version: 1,
						hash_alg: "blake3",
						hash_encoding: "base64",
						file_id: started.fileId,
						name: job.file.name,
						mime: job.file.type || "application/octet-stream",
						extension: (job.file.name.split(".").pop() || "").toLowerCase(),
						size: job.file.size,
						chunk_size: job.public.fileChunkSize,
						chunks: chunks.map((part) => {
							return {
								n: part.chunkNumber,
								plain_size: part.plainSize,
								cipher_size: part.encryptedSize,
								hash: part.encryptedHash,
							};
						}),
					},
					manifestAad: "arkive:file-manifest:v1:" + started.vaultId + ":" + started.fileId,
					chunkHashes: chunks.map((part) => part.encryptedHash),
				});
			if (this.shouldStopJob(job)) {
				throw STOPPED_UPLOAD;
			}

			await this.runWithController(async (signal) => {
				return completeUpload(started.uploadSessionId, {
					encryptedMetadata: preparedMetadata.encryptedMetadata,
					encryptedFileKey: prepared.encryptedFileKey,
					encryptedManifest: finalized.encryptedManifest,
					encryptedHash: finalized.encryptedHash,
					hasThumbnail: !!uploadedThumbnail,
					thumbnailMime: uploadedThumbnail ? uploadedThumbnail.mime : "",
					thumbnailWidth: uploadedThumbnail ? uploadedThumbnail.width : 0,
					thumbnailHeight: uploadedThumbnail ? uploadedThumbnail.height : 0,
				}, signal);
			});
			if (this.shouldStopJob(job)) {
				throw STOPPED_UPLOAD;
			}

			job.public.status = STATUS.COMPLETED;
			job.public.completedBytes = job.file.size;
			job.public.transferRate = 0;
			job.public.updatedAt = nowISO();
			this.releaseJobFile(job);
			this.emitState();
			this.scheduleTerminalCleanup(job);
			completed = true;
			this.notifyBatchComplete(job.public.batchId);
		} finally {
			if (!completed) {
				await clearUploadContext(job.public.jobId).catch(function () {});
			}
		}
	}

	async createAndUploadThumbnail(job, started) {
		const thumbnailCandidate = await generateUploadThumbnail(job.file);
		if (!thumbnailCandidate) {
			return null;
		}
		try {
			return await this.uploadThumbnail(job, started, thumbnailCandidate);
		} catch (error) {
			if (window.location && window.location.hostname === "localhost" && typeof console !== "undefined" && console.warn) {
				console.warn(error);
			}
			return null;
		}
	}

	async uploadThumbnail(job, started, thumbnail) {
		if (!thumbnail || this.shouldStopJob(job)) {
			return null;
		}
		try {
			const encrypted = await encryptUploadThumbnail({
				uploadToken: job.public.jobId,
				thumbnailBytes: thumbnail.bytes,
				aad: "arkive:file-thumbnail:v1:" + started.vaultId + ":" + started.fileId,
			});
			try {
				const presigned = await this.runWithController(async (signal) => {
					return presignThumbnailUpload(started.uploadSessionId, {
						encryptedSize: encrypted.encryptedSize,
						mime: thumbnail.mime,
						width: thumbnail.width,
						height: thumbnail.height,
					}, signal);
				});
				if (this.shouldStopJob(job)) {
					throw STOPPED_UPLOAD;
				}
				const response = await this.runWithController(async (signal) => {
					return fetch(presigned.url, {
						method: "PUT",
						body: encrypted.encryptedThumbnail,
						signal: signal,
					});
				});
				if (!response.ok) {
					throw new Error("Thumbnail upload failed");
				}
				return {
					mime: thumbnail.mime,
					width: thumbnail.width,
					height: thumbnail.height,
					encryptedSize: encrypted.encryptedSize,
				};
			} finally {
				encrypted.encryptedThumbnail.fill(0);
			}
		} finally {
			if (thumbnail.bytes && typeof thumbnail.bytes.fill === "function") {
				thumbnail.bytes.fill(0);
			}
		}
	}

	shouldStopJob(job) {
		return !job || job.public.status === STATUS.CANCELED || this.activeJobId !== job.public.jobId;
	}

	async runWithController(fn) {
		const controller = new AbortController();
		this.currentControllers.add(controller);
		try {
			return await fn(controller.signal);
		} finally {
			this.currentControllers.delete(controller);
		}
	}

	async uploadPartsPooled(job, started, policy) {
		const settings = policy || {};
		const concurrency = Math.max(1, Number(settings.concurrency || 1));
		const results = new Array(job.public.uploadPartCount);
		let nextPartNumber = 1;
		const runner = this;
		const presigner = new PresignCache({
			sessionId: started.uploadSessionId,
			batchSize: Math.max(1, Number(settings.presignBatchSize || 1)),
			fetcher: async function(partNumbers) {
				const limitedParts = [];
				for (let i = 0; i < partNumbers.length; i++) {
					const partNumber = Number(partNumbers[i] || 0);
						if (partNumber <= 0 || partNumber > job.public.uploadPartCount) {
						continue;
					}
					limitedParts.push(partNumber);
				}
				if (!limitedParts.length) {
					return { urls: {} };
				}
				return runner.runWithController(async function(signal) {
					return presignUploadParts(started.uploadSessionId, limitedParts, signal);
				});
			},
		});

		async function worker() {
			while (true) {
				if (runner.shouldStopJob(job)) {
					throw STOPPED_UPLOAD;
				}
					if (nextPartNumber > job.public.uploadPartCount) {
					return;
				}
				const partNumber = nextPartNumber++;
				const result = await runner.uploadPart(job, started, partNumber, presigner);
				results[partNumber - 1] = result;
			}
		}

		const workers = [];
		for (let i = 0; i < concurrency; i++) {
			workers.push(worker());
		}
		await Promise.all(workers);
		return flattenChunkGroups(results);
	}

	async uploadPart(job, started, partNumber, presigner) {
		if (this.shouldStopJob(job)) {
			throw STOPPED_UPLOAD;
		}

		const start = (partNumber - 1) * job.public.uploadPartSize;
		const end = Math.min(start + job.public.uploadPartSize, job.file.size);
		const firstChunkNumber = Math.floor(start / job.public.fileChunkSize) + 1;
		const encryptedChunks = [];
		const chunkResults = [];
		for (let chunkStart = start, chunkNumber = firstChunkNumber; chunkStart < end; chunkStart += job.public.fileChunkSize, chunkNumber++) {
			const chunkEnd = Math.min(chunkStart + job.public.fileChunkSize, job.file.size, end);
			const chunk = new Uint8Array(await job.file.slice(chunkStart, chunkEnd).arrayBuffer());
			const encrypted = await encryptUploadPart({
				uploadToken: job.public.jobId,
				chunkBytes: chunk,
				aad: "arkive:file-chunk:v1:" + started.vaultId + ":" + started.fileId + ":" + chunkNumber + ":" + job.public.fileChunkSize + ":" + job.public.totalChunks,
			});
			encryptedChunks.push(encrypted.encryptedChunk);
			chunkResults.push({
				chunkNumber: chunkNumber,
				plainSize: chunkEnd - chunkStart,
				encryptedHash: encrypted.encryptedHash,
				encryptedSize: encrypted.encryptedSize,
				});
		}
		const uploadBody = joinUint8Arrays(encryptedChunks);
		const uploadHash = await hashUploadPayload({ bytes: uploadBody });

		const presigned = presigner
			? { url: await presigner.get(partNumber) }
			: await this.runWithController(async (signal) => {
				return presignUploadPart(started.uploadSessionId, partNumber, signal);
			});
		if (this.shouldStopJob(job)) {
			throw STOPPED_UPLOAD;
		}
		const response = await this.runWithController(async (signal) => {
			return fetch(presigned.url, {
				method: "PUT",
				body: uploadBody,
				signal: signal,
			});
		});
		if (this.shouldStopJob(job)) {
			throw STOPPED_UPLOAD;
		}
		if (!response.ok) {
			throw new Error("Part upload failed");
		}
		const etag = response.headers.get("etag") || response.headers.get("ETag") || "";
		await this.runWithController(async (signal) => {
			return recordUploadPart(started.uploadSessionId, {
				partNumber: partNumber,
				encryptedHash: uploadHash,
				etag: etag,
			}, signal);
		});
		uploadBody.fill(0);
		if (this.shouldStopJob(job)) {
			throw STOPPED_UPLOAD;
		}

		job.public.completedUploadParts += 1;
		job.public.completedParts = job.public.completedUploadParts;
		job.public.completedBytes += end - start;
		const startedAt = job.public.startedAt ? new Date(job.public.startedAt).getTime() : Date.now();
		const elapsed = Math.max(1, Date.now() - startedAt);
		job.public.transferRate = Math.round((job.public.completedBytes / elapsed) * 1000);
		job.public.updatedAt = nowISO();
		this.emitState();

		return chunkResults;
	}

	async failJob(job, error) {
		if (!job || job.public.status === STATUS.CANCELED) {
			return;
		}
		if (job.public.sessionId) {
			await cancelUpload(job.public.sessionId).catch(function () {});
		}
		this.releaseJobFile(job);
		job.public.status = STATUS.FAILED;
		job.public.transferRate = 0;
		job.public.updatedAt = nowISO();
		this.emitState();
		this.scheduleTerminalCleanup(job);
		this.emitEvent({ type: "error", jobId: job.public.jobId, error: String(error && error.message ? error.message : error || "Upload failed") });
	}

	async cancelJob(jobId) {
		const job = this.jobs.get(jobId);
		if (!job) return;
		if (isTerminal(job.public.status)) return;
		job.public.status = STATUS.CANCELED;
		this.releaseJobFile(job);
		job.public.transferRate = 0;
		job.public.updatedAt = nowISO();
		if (this.activeJobId === jobId) {
			for (const controller of this.currentControllers) {
				controller.abort();
			}
		}
		if (job.public.sessionId) {
			await cancelUpload(job.public.sessionId).catch(function () {});
		}
		this.clearJobCleanup(jobId);
		this.jobs.delete(jobId);
		this.emitState();
	}

	removeJob(jobId) {
		if (!this.jobs.has(jobId)) return;
		const job = this.jobs.get(jobId);
		this.clearJobCleanup(jobId);
		this.releaseJobFile(job);
		this.jobs.delete(jobId);
		if (this.activeJobId === jobId) {
			this.activeJobId = null;
			for (const controller of this.currentControllers) {
				controller.abort();
			}
		}
		this.emitState();
	}

	async cancelAll() {
		const ids = Array.from(this.jobs.keys());
		for (let i = 0; i < ids.length; i++) {
			const job = this.jobs.get(ids[i]);
			if (!job || isTerminal(job.public.status)) continue;
			await this.cancelJob(ids[i]);
		}
	}

	cancelActiveUploadsBestEffort() {
		for (const controller of this.currentControllers) {
			controller.abort();
		}
		for (const [jobId, job] of this.jobs.entries()) {
			if (job.public.status !== STATUS.QUEUED && job.public.status !== STATUS.RUNNING) {
				continue;
			}
			job.public.status = STATUS.CANCELED;
			this.releaseJobFile(job);
			job.public.transferRate = 0;
			job.public.updatedAt = nowISO();
			if (job.public.sessionId) {
				cancelUploadBestEffort(job.public.sessionId);
			}
			this.clearJobCleanup(jobId);
			this.jobs.delete(jobId);
		}
		this.emitState();
	}

	notifyBatchComplete(batchId) {
		if (!batchId || this.completedBatches.has(batchId)) {
			return;
		}
		let total = 0;
		let done = 0;
		for (const job of this.jobs.values()) {
			if (job.public.batchId !== batchId) continue;
			total++;
			if (job.public.status === STATUS.COMPLETED) done++;
		}
		if (total > 0 && done >= total) {
			this.completedBatches.add(batchId);
			this.emitEvent({ type: "batch-complete", batchId: batchId, total: total });
		}
	}
}
