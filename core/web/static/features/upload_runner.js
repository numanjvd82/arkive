import { UPLOAD_POLICY } from "../upload/upload_policy.js";
import { STATUS, isTerminal } from "../upload/upload_state.js";
import { startUpload, presignUploadPart, recordUploadPart, completeUpload, cancelUpload, cancelUploadBestEffort } from "../upload/upload_api.js";
import { setVaultSession, restoreVaultSession, prepareUpload, encryptUploadPart, finalizeUpload, clearUploadContext } from "../upload/upload_crypto.js";

const STOPPED_UPLOAD = new Error("Upload stopped");

function nowISO() {
	return new Date().toISOString();
}

function uuid() {
	return self.crypto && self.crypto.randomUUID ? self.crypto.randomUUID() : "job-" + Date.now() + "-" + Math.random().toString(36).slice(2);
}

export class UploadRunner {
	constructor(options) {
		options = options || {};
		this.limits = options.limits || { maxQueueItems: 300 };
		this.jobs = new Map();
		this.activeJobId = null;
		this.currentController = null;
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
			const chunkSize = UPLOAD_POLICY.defaultPartSize;
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
					chunkSize: chunkSize,
					totalParts: Math.max(1, Math.ceil(file.size / chunkSize)),
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
			this.currentController = null;
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
				partSize: job.public.chunkSize,
				totalParts: job.public.totalParts,
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
			if (this.shouldStopJob(job)) {
				throw STOPPED_UPLOAD;
			}

			const parts = [];
			for (let partNumber = 1; partNumber <= job.public.totalParts; partNumber++) {
				parts.push(await this.uploadPart(job, started, partNumber));
			}
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
					chunk_size: job.public.chunkSize,
					chunks: parts.map((part) => {
						return {
							n: part.partNumber,
							offset: part.offset,
							plain_size: part.plainSize,
							cipher_size: part.encryptedSize,
							hash: part.encryptedHash,
						};
					}),
				},
				manifestAad: "arkive:file-manifest:v1:" + started.vaultId + ":" + started.fileId,
				partHashes: parts.map((part) => part.encryptedHash),
			});
			if (this.shouldStopJob(job)) {
				throw STOPPED_UPLOAD;
			}

			await this.runWithController(async (signal) => {
				return completeUpload(started.uploadSessionId, {
					encryptedMetadata: prepared.encryptedMetadata,
					encryptedFileKey: prepared.encryptedFileKey,
					encryptedManifest: finalized.encryptedManifest,
					encryptedHash: finalized.encryptedHash,
				}, signal);
			});
			if (this.shouldStopJob(job)) {
				throw STOPPED_UPLOAD;
			}

			job.public.status = STATUS.COMPLETED;
			job.public.completedBytes = job.file.size;
			job.public.transferRate = 0;
			job.public.updatedAt = nowISO();
			this.emitState();
			completed = true;
			this.notifyBatchComplete(job.public.batchId);
		} finally {
			if (!completed) {
				await clearUploadContext(job.public.jobId).catch(function () {});
			}
		}
	}

	shouldStopJob(job) {
		return !job || job.public.status === STATUS.CANCELED || this.activeJobId !== job.public.jobId;
	}

	async runWithController(fn) {
		const controller = new AbortController();
		this.currentController = controller;
		try {
			return await fn(controller.signal);
		} finally {
			if (this.currentController === controller) {
				this.currentController = null;
			}
		}
	}

	async uploadPart(job, started, partNumber) {
		if (this.shouldStopJob(job)) {
			throw STOPPED_UPLOAD;
		}

		const start = (partNumber - 1) * job.public.chunkSize;
		const end = Math.min(start + job.public.chunkSize, job.file.size);
		const chunk = new Uint8Array(await job.file.slice(start, end).arrayBuffer());
		const encrypted = await encryptUploadPart({
			uploadToken: job.public.jobId,
			chunkBytes: chunk,
			aad: "arkive:file-chunk:v1:" + started.vaultId + ":" + started.fileId + ":" + partNumber + ":" + job.public.chunkSize + ":" + job.public.totalParts,
		});

		const presigned = await this.runWithController(async (signal) => {
			return presignUploadPart(started.uploadSessionId, partNumber, signal);
		});
		if (this.shouldStopJob(job)) {
			throw STOPPED_UPLOAD;
		}
		const response = await this.runWithController(async (signal) => {
			return fetch(presigned.url, {
				method: "PUT",
				body: encrypted.encryptedChunk,
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
				encryptedHash: encrypted.encryptedHash,
				etag: etag,
			}, signal);
		});
		if (this.shouldStopJob(job)) {
			throw STOPPED_UPLOAD;
		}

		job.public.completedParts += 1;
		job.public.completedBytes += end - start;
		const startedAt = job.public.startedAt ? new Date(job.public.startedAt).getTime() : Date.now();
		const elapsed = Math.max(1, Date.now() - startedAt);
		job.public.transferRate = Math.round((job.public.completedBytes / elapsed) * 1000);
		job.public.updatedAt = nowISO();
		this.emitState();

		return {
			partNumber: partNumber,
			offset: start,
			plainSize: end - start,
			encryptedHash: encrypted.encryptedHash,
			encryptedSize: encrypted.encryptedSize,
		};
	}

	async failJob(job, error) {
		job.public.status = STATUS.FAILED;
		job.public.transferRate = 0;
		job.public.updatedAt = nowISO();
		this.emitState();
		this.emitEvent({ type: "error", jobId: job.public.jobId, error: String(error && error.message ? error.message : error || "Upload failed") });
	}

	async cancelJob(jobId) {
		const job = this.jobs.get(jobId);
		if (!job) return;
		if (isTerminal(job.public.status)) return;
		job.public.status = STATUS.CANCELED;
		job.public.transferRate = 0;
		job.public.updatedAt = nowISO();
		if (this.activeJobId === jobId && this.currentController) {
			this.currentController.abort();
		}
		if (job.public.sessionId) {
			await cancelUpload(job.public.sessionId).catch(function () {});
		}
		this.emitState();
	}

	removeJob(jobId) {
		if (!this.jobs.has(jobId)) return;
		this.jobs.delete(jobId);
		if (this.activeJobId === jobId) {
			this.activeJobId = null;
			if (this.currentController) {
				this.currentController.abort();
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
		if (this.currentController) {
			this.currentController.abort();
		}
		for (const job of this.jobs.values()) {
			if (job.public.status !== STATUS.QUEUED && job.public.status !== STATUS.RUNNING) {
				continue;
			}
			job.public.status = STATUS.CANCELED;
			job.public.transferRate = 0;
			job.public.updatedAt = nowISO();
			if (job.public.sessionId) {
				cancelUploadBestEffort(job.public.sessionId);
			}
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
