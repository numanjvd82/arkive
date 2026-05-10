import { UploadRunner } from "./upload_runner.js";

function formatBytes(bytes) {
	if (!bytes) return "0 B";
	const units = ["B", "KB", "MB", "GB", "TB"];
	const i = Math.floor(Math.log(bytes) / Math.log(1024));
	return (bytes / Math.pow(1024, i)).toFixed(i > 0 ? 1 : 0) + " " + units[i];
}

function parseQueueLimit(value, fallback) {
	const parsed = Number.parseInt(String(value || ""), 10);
	if (!Number.isFinite(parsed)) return fallback;
	return parsed <= 0 ? 0 : parsed;
}

export function initUploads() {
	if (document.body.hasAttribute("data-uploads-ready")) return;
	document.body.setAttribute("data-uploads-ready", "true");

	const input = document.getElementById("upload-file");
	const browseFilesButton = document.getElementById("upload-browse-files");
	const abortButton = document.getElementById("upload-abort");
	const dropzone = document.getElementById("upload-dropzone");
	const confirmBackdrop = document.getElementById("upload-confirm-backdrop");
	const confirmStart = document.getElementById("upload-confirm-start");
	const confirmCancel = document.getElementById("upload-confirm-cancel");
	const metaTitle = document.getElementById("upload-meta-title");
	const metaDetail = document.getElementById("upload-meta-detail");
	const status = document.getElementById("upload-status");
	const queueList = document.getElementById("upload-queue-list");
	const queueMeta = document.getElementById("upload-queue-meta");
	const queueEmpty = document.getElementById("upload-queue-empty");
	const uploadControls = document.getElementById("upload-controls");
	const uploadChip = document.getElementById("upload-chip");
	const uploadChipName = document.getElementById("upload-chip-name");
	const uploadChipSize = document.getElementById("upload-chip-size");
	const uploadChipClear = document.getElementById("upload-chip-clear");
	const uploadIconBank = document.querySelector(".upload-icon-bank");

	if (!input || !queueList || !status) return;

	const uploadLimits = {
		maxQueueItems: parseQueueLimit(dropzone && dropzone.getAttribute("data-upload-max-queue-items"), 300),
	};
	const runner = new UploadRunner({ limits: uploadLimits });
	let selectedFiles = [];
	let state = { jobs: [], activeJobId: null };
	const completedBatches = new Set();
	let renderScheduled = false;

	runner.onState(function (nextState) {
		state = nextState || state;
		scheduleRender();
	});
	runner.onEvent(function (event) {
		if (event && event.type === "error") {
			setStatus(event.error || "Upload failed.");
			return;
		}
		if (event && event.type === "batch-complete") {
			if (event.batchId && !completedBatches.has(event.batchId)) {
				completedBatches.add(event.batchId);
				if (window.Toast) {
					window.Toast.success((event.total || 0) + " file(s) uploaded.", { title: "Upload complete" });
				}
			}
		}
	});
	state = runner.getState();
	scheduleRender();

	if (window.ArkiveVault && typeof window.ArkiveVault.onSessionUnlock === "function") {
		window.ArkiveVault.onSessionUnlock(function (session) {
			runner.setVaultSession(session);
		});
	}
	if (window.ArkiveVault && typeof window.ArkiveVault.getSessionUnlock === "function") {
		runner.setVaultSession(window.ArkiveVault.getSessionUnlock());
	}

	browseFilesButton && browseFilesButton.addEventListener("click", function () { input.click(); });
	input.addEventListener("change", function () {
		if (!input.files || !input.files.length) return;
		selectedFiles = Array.from(input.files);
		showSelectedFiles(selectedFiles);
		showStartDialog(selectedFiles.length);
		input.value = "";
	});

	dropzone && dropzone.addEventListener("dragover", function (e) { e.preventDefault(); dropzone.classList.add("is-dragover"); });
	dropzone && dropzone.addEventListener("dragleave", function () { dropzone.classList.remove("is-dragover"); });
	dropzone && dropzone.addEventListener("drop", function (e) {
		e.preventDefault();
		dropzone.classList.remove("is-dragover");
		if (e.dataTransfer.files && e.dataTransfer.files.length) {
			selectedFiles = Array.from(e.dataTransfer.files);
			showSelectedFiles(selectedFiles);
			showStartDialog(selectedFiles.length);
		}
	});

	uploadChipClear && uploadChipClear.addEventListener("click", function () {
		selectedFiles = [];
		hideSelectedFiles();
	});

	confirmStart && confirmStart.addEventListener("click", function () {
		hideDialog();
		if (selectedFiles.length) {
			runner.addFiles(selectedFiles.slice()).catch(function (error) {
				setStatus(error && error.message ? error.message : "Upload failed.");
				if (typeof console !== "undefined" && console.error) console.error(error);
			});
			selectedFiles = [];
			hideSelectedFiles();
		}
	});

	confirmCancel && confirmCancel.addEventListener("click", function () {
		hideDialog();
		selectedFiles = [];
		hideSelectedFiles();
	});

	abortButton && abortButton.addEventListener("click", function () {
		runner.cancelAll().catch(function (error) {
			if (typeof console !== "undefined" && console.error) console.error(error);
		});
	});

	queueList.addEventListener("click", function (event) {
		const target = event.target;
		if (!target || !target.closest) return;
		const button = target.closest("[data-job-action]");
		if (!button) return;
		const jobId = button.getAttribute("data-job-id");
		const action = button.getAttribute("data-job-action");
		if (!jobId || !action) return;
		if (action === "cancel") runner.cancelJob(jobId).catch(function (error) { if (typeof console !== "undefined" && console.error) console.error(error); });
		if (action === "remove") runner.removeJob(jobId);
	});

	window.addEventListener("beforeunload", function (event) {
		if (!runner.hasActiveUploads()) return;
		event.preventDefault();
		event.returnValue = "";
	});

	window.addEventListener("pagehide", function () {
		if (!runner.hasActiveUploads()) return;
		runner.cancelActiveUploadsBestEffort();
	});

	document.addEventListener("click", function (event) {
		const target = event.target;
		if (!target || !target.closest || !runner.hasActiveUploads()) return;
		const link = target.closest("a[href]");
		if (!link) return;
		if (link.hasAttribute("download")) return;
		if (link.target && link.target !== "_self") return;
		if (event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
		const href = link.getAttribute("href") || "";
		if (!href || href.charAt(0) === "#") return;
		let url = null;
		try {
			url = new URL(link.href, window.location.href);
		} catch (_) {
			return;
		}
		if (url.origin !== window.location.origin) return;
		if (url.pathname === window.location.pathname && url.search === window.location.search && url.hash === window.location.hash) return;
		const ok = window.confirm("Upload in progress. Leaving this page will cancel it.");
		if (!ok) {
			event.preventDefault();
			return;
		}
		runner.cancelActiveUploadsBestEffort();
	});

	function setStatus(text) { status.textContent = text; }
	function hideDialog() { confirmBackdrop && confirmBackdrop.classList.add("is-hidden"); }
	function showDialog() { confirmBackdrop && confirmBackdrop.classList.remove("is-hidden"); }
	function showStartDialog(count) {
		const title = document.getElementById("upload-confirm-title");
		const body = document.getElementById("upload-confirm-meta");
		if (title) title.textContent = "Start upload?";
		if (body) body.textContent = count + " file(s) selected.";
		if (confirmStart) confirmStart.textContent = "Start upload";
		if (confirmCancel) confirmCancel.textContent = "Cancel";
		showDialog();
	}
	function showSelectedFiles(files) {
		if (!uploadChip || !uploadChipName || !uploadChipSize) return;
		uploadChipName.textContent = files.length === 1 ? files[0].name : files.length + " files";
		uploadChipSize.textContent = formatBytes(files.reduce(function (sum, file) { return sum + file.size; }, 0));
		uploadChip.classList.remove("is-hidden");
	}
	function hideSelectedFiles() {
		uploadChip && uploadChip.classList.add("is-hidden");
	}

	function scheduleRender() {
		if (renderScheduled) return;
		renderScheduled = true;
		requestAnimationFrame(function () {
			renderScheduled = false;
			render();
		});
	}

	function render() {
		const jobs = (state.jobs || []).slice().sort(function (a, b) { return String(a.createdAt || "").localeCompare(String(b.createdAt || "")); });
		const existing = new Map();
		Array.from(queueList.children).forEach(function (node) {
			if (node && node.id) {
				existing.set(node.id, node);
			}
		});
		const hasCancelableJobs = jobs.some(function (job) {
			return job.status === "queued" || job.status === "running";
		});
		if (uploadControls) {
			uploadControls.classList.toggle("is-hidden", !hasCancelableJobs);
		}
		if (!jobs.length) {
			queueEmpty && queueEmpty.classList.remove("is-hidden");
			queueMeta && (queueMeta.textContent = "0 items active");
			if (!state.activeJobId) setStatus("No uploads in progress.");
			return;
		}
		queueEmpty && queueEmpty.classList.add("is-hidden");
		let activeCount = 0;
		const fragment = document.createDocumentFragment();
		jobs.forEach(function (job) {
			if (job.status === "running") activeCount++;
			const id = "upload-task-" + job.jobId;
			const node = renderJob(job, existing.get(id) || null);
			existing.delete(id);
			fragment.appendChild(node);
		});
		existing.forEach(function (node) {
			if (node && node.parentNode === queueList) {
				node.parentNode.removeChild(node);
			}
		});
		queueList.replaceChildren(fragment);
		if (queueMeta) queueMeta.textContent = activeCount > 0 ? activeCount + " item" + (activeCount > 1 ? "s" : "") + " active" : jobs.length + " queued";
		const active = jobs.find(function (job) { return job.status === "running"; });
		if (active) {
			setStatus("Uploading " + active.fileName + ". Leaving this page will cancel it.");
		} else if (jobs.some(function (job) { return job.status === "queued"; })) {
			setStatus("Uploads queued. Leaving this page will cancel them.");
		}
	}

	function renderJob(job, existing) {
		const li = existing || document.createElement("li");
		if (!existing) {
			li.id = "upload-task-" + job.jobId;
			li.innerHTML = '<div class="queue-item-top"><div class="queue-item-file"><span class="queue-item-name"></span><span class="queue-item-badge"></span></div><div class="queue-item-actions"></div></div><div class="queue-item-track"><span class="queue-item-fill"></span></div><div class="queue-item-meta"><span class="queue-item-progress mono"></span><span class="queue-item-speed mono"></span></div>';
		}
		li.className = "queue-item" + (job.status === "completed" ? " is-complete" : job.status === "failed" || job.status === "canceled" ? " is-error" : "");
		const nameEl = li.querySelector(".queue-item-name");
		const badge = li.querySelector(".queue-item-badge");
		const actions = li.querySelector(".queue-item-actions");
		const fill = li.querySelector(".queue-item-fill");
		const progress = li.querySelector(".queue-item-progress");
		const detail = li.querySelector(".queue-item-speed");
		if (nameEl) nameEl.textContent = job.fileName;
		if (badge) {
			badge.className = "queue-item-badge is-" + job.status;
			badge.textContent = job.status;
		}
		if (actions) {
			const signature = actionSignature(job);
			if (actions.getAttribute("data-signature") !== signature) {
				actions.textContent = "";
				actions.setAttribute("data-signature", signature);
				buildActions(actions, job);
			}
		}
		if (fill) {
			fill.style.transition = "width 180ms linear";
			fill.style.width = job.fileSize > 0 ? Math.round(((job.completedBytes || 0) / job.fileSize) * 100) + "%" : "0%";
		}
		if (progress) {
			progress.textContent = formatBytes(job.completedBytes || 0) + " / " + formatBytes(job.fileSize);
		}
		if (detail) {
			detail.textContent = job.transferRate > 0 ? formatBytes(job.transferRate) + "/s" : job.status;
		}
		return li;
	}

	function actionSignature(job) {
		return String(job.status || "");
	}

	function buildActions(actions, job) {
		if (job.status === "queued" || job.status === "running") {
			actions.appendChild(actionButton(job.jobId, "cancel", "Cancel", "queue-item-action is-cancel", "trash"));
		} else if (job.status === "failed") {
			actions.appendChild(actionButton(job.jobId, "remove", "Remove", "queue-item-action is-cancel", "trash"));
		} else if (job.status === "completed" || job.status === "canceled") {
			actions.appendChild(actionButton(job.jobId, "remove", "Remove", "queue-item-action is-cancel", "trash"));
		}
	}

	function actionButton(jobId, action, label, className, iconName) {
		const button = document.createElement("button");
		button.type = "button";
		button.className = className;
		button.setAttribute("data-job-id", jobId);
		button.setAttribute("data-job-action", action);
		button.setAttribute("aria-label", label);
		button.title = label;
		const icon = uploadIconBank && uploadIconBank.querySelector('[data-upload-icon="' + (iconName || action) + '"] svg');
		if (icon) {
			button.appendChild(icon.cloneNode(true));
		}
		return button;
	}
}
