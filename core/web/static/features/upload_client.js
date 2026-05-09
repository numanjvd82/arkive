import { UPLOAD_MESSAGE } from "../lib/upload_protocol.js";

export class UploadClient {
	constructor(options) {
		this.workerURL = options.workerURL || "/static/workers/upload_shared_worker.js";
		this.worker = null;
		this.port = null;
		this.pending = new Map();
		this.requestID = 0;
		this.stateHandlers = [];
		this.eventHandlers = [];
	}

	connect() {
		if (this.port) return this.port;
		if (typeof console !== "undefined" && console.log) {
			console.log("[arkive-uploads] connect worker", this.workerURL);
		}
		const worker = new SharedWorker(this.workerURL, { type: "module", name: "arkive-uploads" });
		this.worker = worker;
		this.port = worker.port;
		this.port.start();
		this.port.addEventListener("message", (event) => this.handleMessage(event.data || {}));
		this.port.addEventListener("messageerror", () => {
			if (typeof console !== "undefined" && console.error) {
				console.error("[arkive-uploads] worker messageerror");
			}
		});
		worker.onerror = (event) => {
			if (typeof console !== "undefined" && console.error) {
				console.error("[arkive-uploads] worker error", event && event.message ? event.message : event);
			}
		};
		this.port.postMessage({ type: UPLOAD_MESSAGE.CONNECT });
		return this.port;
	}

	onState(handler) {
		if (typeof handler === "function") this.stateHandlers.push(handler);
	}

	onEvent(handler) {
		if (typeof handler === "function") this.eventHandlers.push(handler);
	}

	setVaultSession(session) {
		this.connect();
		this.port.postMessage({ type: UPLOAD_MESSAGE.VAULT_SESSION, session: session || null });
	}

	requestState() { return this.request(UPLOAD_MESSAGE.REQUEST_STATE, {}); }
	addFiles(files) { return this.request(UPLOAD_MESSAGE.ADD_FILES, { files: files || [] }); }
	resumeWithFiles(files) { return this.request(UPLOAD_MESSAGE.RESUME_FILES, { files: files || [] }); }
	pauseJob(jobId) { return this.request(UPLOAD_MESSAGE.PAUSE_JOB, { jobId: jobId }); }
	resumeJob(jobId) { return this.request(UPLOAD_MESSAGE.RESUME_JOB, { jobId: jobId }); }
	cancelJob(jobId) { return this.request(UPLOAD_MESSAGE.CANCEL_JOB, { jobId: jobId }); }
	removeJob(jobId) { return this.request(UPLOAD_MESSAGE.REMOVE_JOB, { jobId: jobId }); }
	resumeAll() { return this.request(UPLOAD_MESSAGE.RESUME_ALL, {}); }
	cancelAll() { return this.request(UPLOAD_MESSAGE.CANCEL_ALL, {}); }

	request(type, payload) {
		this.connect();
		return new Promise((resolve, reject) => {
			const id = ++this.requestID;
			this.pending.set(id, { resolve: resolve, reject: reject });
			this.port.postMessage({ id: id, type: type, payload: payload || {} });
		});
	}

	handleMessage(message) {
		if (typeof message.id === "number" && this.pending.has(message.id)) {
			const entry = this.pending.get(message.id);
			this.pending.delete(message.id);
			if (message.ok) entry.resolve(message.result);
			else entry.reject(new Error(message.error || "Upload worker failed"));
			return;
		}
		if (message.type === UPLOAD_MESSAGE.STATE) {
			this.stateHandlers.forEach((handler) => handler(message.state || {}));
			return;
		}
		if (typeof console !== "undefined" && console.log) {
			console.log("[arkive-uploads] event", message);
		}
		this.eventHandlers.forEach((handler) => handler(message));
	}
}
