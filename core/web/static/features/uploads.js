import { UploadClient } from "./upload_client.js";

function formatBytes(bytes) {
	if (!bytes) return "0 B";
	const units = ["B", "KB", "MB", "GB", "TB"];
	const i = Math.floor(Math.log(bytes) / Math.log(1024));
	return (bytes / Math.pow(1024, i)).toFixed(i > 0 ? 1 : 0) + " " + units[i];
}

function parseLimit(value, fallback) {
	const parsed = Number.parseInt(String(value || ""), 10);
	return Number.isFinite(parsed) && parsed > 0 ? parsed : fallback;
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
		maxQueueItems: parseLimit(dropzone && dropzone.getAttribute("data-upload-max-queue-items"), 300),
	};
	const client = new UploadClient({ limits: uploadLimits });
	let selectedFiles = [];
	let selectedFileHandles = [];
	let state = { jobs: [], incompleteJobs: [], activeJobId: null };
	let startMode = "new";
	const completedBatches = new Set();

	client.connect();
	client.setLimits(uploadLimits);
	client.onState(function (nextState) {
		state = nextState || state;
		render();
	});
	client.onEvent(function (event) {
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
	client.requestState().then(function (snapshot) {
		state = snapshot || state;
		render();
		if (state.incompleteJobs && state.incompleteJobs.length) {
			showResumeDialog(state.incompleteJobs.length);
		}
	}).catch(function () {});

	if (window.ArkiveVault && typeof window.ArkiveVault.onSessionUnlock === "function") {
		window.ArkiveVault.onSessionUnlock(function (session) {
			client.setVaultSession(session);
		});
	}
	if (window.ArkiveVault && typeof window.ArkiveVault.getSessionUnlock === "function") {
		client.setVaultSession(window.ArkiveVault.getSessionUnlock());
	}

	browseFilesButton && browseFilesButton.addEventListener("click", function () { input.click(); });
	input.addEventListener("change", function () {
		if (!input.files || !input.files.length) return;
		selectedFileHandles = [];
		selectedFiles = Array.from(input.files);
		startMode = "new";
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
			startMode = "new";
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
		if (typeof console !== "undefined" && console.log) {
			console.log("[arkive-uploads] confirm start", { startMode: startMode, selectedFiles: selectedFiles.length, selectedHandles: selectedFileHandles.length });
		}
		hideDialog();
		if (startMode === "resume") {
			resumeIncompleteJobs().catch(function (error) {
				setStatus(error && error.message ? error.message : "Resume failed.");
				if (typeof console !== "undefined" && console.error) console.error(error);
			});
			return;
		}
		if (selectedFiles.length) {
			client.addFiles(selectedFiles.slice()).catch(function (error) {
				setStatus(error && error.message ? error.message : "Upload failed.");
				if (typeof console !== "undefined" && console.error) console.error(error);
			});
			selectedFiles = [];
			selectedFileHandles = [];
			hideSelectedFiles();
		}
	});

	confirmCancel && confirmCancel.addEventListener("click", function () {
		hideDialog();
		if (startMode === "resume") {
			client.cancelAll().catch(function (error) {
				if (typeof console !== "undefined" && console.error) console.error(error);
			});
		}
		selectedFiles = [];
		selectedFileHandles = [];
		hideSelectedFiles();
	});

	abortButton && abortButton.addEventListener("click", function () {
		client.cancelAll().catch(function (error) {
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
		if (action === "cancel") client.cancelJob(jobId).catch(function (error) { if (typeof console !== "undefined" && console.error) console.error(error); });
		if (action === "pause") client.pauseJob(jobId).catch(function (error) { if (typeof console !== "undefined" && console.error) console.error(error); });
		if (action === "resume") client.resumeJob(jobId).catch(function (error) { if (typeof console !== "undefined" && console.error) console.error(error); });
		if (action === "remove") client.removeJob(jobId).catch(function (error) { if (typeof console !== "undefined" && console.error) console.error(error); });
	});

	function setStatus(text) { status.textContent = text; }
	function hideDialog() { confirmBackdrop && confirmBackdrop.classList.add("is-hidden"); }
	function showDialog() { confirmBackdrop && confirmBackdrop.classList.remove("is-hidden"); }
	function showStartDialog(count) {
		startMode = "new";
		const title = document.getElementById("upload-confirm-title");
		const body = document.getElementById("upload-confirm-meta");
		if (title) title.textContent = "Start upload?";
		if (body) body.textContent = count + " file(s) selected.";
		if (confirmStart) confirmStart.textContent = "Start upload";
		if (confirmCancel) confirmCancel.textContent = "Cancel";
		showDialog();
	}
	function showResumeDialog(count) {
		startMode = "resume";
		const title = document.getElementById("upload-confirm-title");
		const body = document.getElementById("upload-confirm-meta");
		if (title) title.textContent = "Resume uploads?";
		if (body) body.textContent = count + " incomplete upload job(s) found. Reselect file to resume or cancel them.";
		if (confirmStart) confirmStart.textContent = "Open files";
		if (confirmCancel) confirmCancel.textContent = "Cancel all";
		showDialog();
	}

	async function resumeIncompleteJobs() {
		if (selectedFileHandles.length) {
			const files = [];
			for (let i = 0; i < selectedFileHandles.length; i++) {
				if (selectedFileHandles[i] && typeof selectedFileHandles[i].getFile === "function") {
					files.push(await selectedFileHandles[i].getFile());
				}
			}
			if (files.length) {
				selectedFiles = files;
				client.resumeWithFiles(files).catch(function (error) {
					setStatus(error && error.message ? error.message : "Resume failed.");
					if (typeof console !== "undefined" && console.error) console.error(error);
				});
				selectedFiles = [];
				selectedFileHandles = [];
				hideSelectedFiles();
			}
			return;
		}
		if (window.showOpenFilePicker) {
			try {
				const handles = await window.showOpenFilePicker({ multiple: true });
				selectedFileHandles = handles || [];
				const files = [];
				for (let i = 0; i < selectedFileHandles.length; i++) {
					if (selectedFileHandles[i] && typeof selectedFileHandles[i].getFile === "function") {
						files.push(await selectedFileHandles[i].getFile());
					}
				}
				if (files.length) {
					client.resumeWithFiles(files).catch(function (error) {
						setStatus(error && error.message ? error.message : "Resume failed.");
						if (typeof console !== "undefined" && console.error) console.error(error);
					});
				}
			} catch (_) {
				return;
			}
		}
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

	function render() {
		queueList.innerHTML = "";
		const jobs = (state.jobs || []).slice().sort(function (a, b) { return String(a.createdAt || "").localeCompare(String(b.createdAt || "")); });
		const hasCancelableJobs = jobs.some(function (job) {
			return job.status === "queued" || job.status === "running" || job.status === "paused";
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
		jobs.forEach(function (job) {
			if (job.status === "running") activeCount++;
			queueList.appendChild(renderJob(job));
		});
		if (queueMeta) queueMeta.textContent = activeCount > 0 ? activeCount + " item" + (activeCount > 1 ? "s" : "") + " active" : jobs.length + " queued";
		const active = jobs.find(function (job) { return job.status === "running"; });
		if (active) {
			setStatus("Uploading " + active.fileName + "...");
		}
	}

	function renderJob(job) {
		const li = document.createElement("li");
		li.className = "queue-item" + (job.status === "completed" ? " is-complete" : job.status === "failed" || job.status === "canceled" ? " is-error" : "");
		li.id = "upload-task-" + job.jobId;
		const top = document.createElement("div");
		top.className = "queue-item-top";
		const fileWrap = document.createElement("div");
		fileWrap.className = "queue-item-file";
		const nameEl = document.createElement("span");
		nameEl.className = "queue-item-name";
		nameEl.textContent = job.fileName;
		const badge = document.createElement("span");
		badge.className = "queue-item-badge is-" + job.status;
		badge.textContent = job.status;
		fileWrap.appendChild(nameEl);
		fileWrap.appendChild(badge);
		const actions = document.createElement("div");
		actions.className = "queue-item-actions";
		if (job.status === "paused") {
			actions.appendChild(actionButton(job.jobId, "resume", "Resume", "queue-item-action", "play"));
			actions.appendChild(actionButton(job.jobId, "cancel", "Cancel", "queue-item-action is-cancel", "trash"));
		} else if (job.status === "queued" || job.status === "running") {
			actions.appendChild(actionButton(job.jobId, "pause", "Pause", "queue-item-action", "pause"));
			actions.appendChild(actionButton(job.jobId, "cancel", "Cancel", "queue-item-action is-cancel", "trash"));
		} else if (job.status === "failed") {
			actions.appendChild(actionButton(job.jobId, "resume", "Retry", "queue-item-action", "refresh-cw"));
			actions.appendChild(actionButton(job.jobId, "remove", "Remove", "queue-item-action is-cancel", "trash"));
		} else if (job.status === "completed") {
			actions.appendChild(actionButton(job.jobId, "remove", "Remove", "queue-item-action is-cancel", "trash"));
		} else if (job.status === "canceled") {
			actions.appendChild(actionButton(job.jobId, "remove", "Remove", "queue-item-action is-cancel", "trash"));
		}
		top.appendChild(fileWrap);
		top.appendChild(actions);
		const track = document.createElement("div");
		track.className = "queue-item-track";
		const fill = document.createElement("span");
		fill.className = "queue-item-fill";
		fill.style.transition = "width 180ms linear";
		fill.style.width = job.fileSize > 0 ? Math.round(((job.completedBytes || 0) / job.fileSize) * 100) + "%" : "0%";
		track.appendChild(fill);
		const meta = document.createElement("div");
		meta.className = "queue-item-meta";
		const progress = document.createElement("span");
		progress.className = "queue-item-progress mono";
		progress.textContent = formatBytes(job.completedBytes || 0) + " / " + formatBytes(job.fileSize);
		const detail = document.createElement("span");
		detail.className = "queue-item-speed mono";
		detail.textContent = job.transferRate > 0 ? formatBytes(job.transferRate) + "/s" : job.status;
		meta.appendChild(progress);
		meta.appendChild(detail);
		li.appendChild(top);
		li.appendChild(track);
		li.appendChild(meta);
		return li;
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
