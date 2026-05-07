import { resolveUploadChunkSize } from "./file_chunks.js";
import { MultipartUploadPipeline } from "./upload_pipeline.js";

export function initUploads() {
  if (document.body.hasAttribute("data-uploads-ready")) {
    return;
  }
  document.body.setAttribute("data-uploads-ready", "true");

  const input = document.getElementById("upload-file");
  const browseFilesButton = document.getElementById("upload-browse-files");
  const startButton = document.getElementById("upload-start");
  const abortButton = document.getElementById("upload-abort");
  const dropzone = document.getElementById("upload-dropzone");
  const confirmBackdrop = document.getElementById("upload-confirm-backdrop");
  const confirmStart = document.getElementById("upload-confirm-start");
  const confirmCancel = document.getElementById("upload-confirm-cancel");
  const metaTitle = document.getElementById("upload-meta-title");
  const metaDetail = document.getElementById("upload-meta-detail");
  const metaTooltip = document.getElementById("upload-meta-tooltip");
  const status = document.getElementById("upload-status");
  const queueList = document.getElementById("upload-queue-list");
  const queueMeta = document.getElementById("upload-queue-meta");
  const queueEmpty = document.getElementById("upload-queue-empty");

  if (!input || !status || !queueList) {
    return;
  }

  const MAX_QUEUE_ITEMS = 300;
  const MAX_FILE_SIZE = 10 * 1024 * 1024 * 1024;
  const MAX_CONCURRENCY = 4;
  const AUTO_CLEAR_SUCCESS_MS = 1800;
  const AUTO_CLEAR_CANCELLED_MS = 1200;

  const uploadPipeline = new MultipartUploadPipeline({
    concurrency: 4,
    retries: 3,
    onProgress: function (task, loaded, total) {
      updateTaskProgress(task, loaded, total);
      updatePrimaryProgress();
    },
  });

  let primaryTaskId = null;
  let batchCounter = 0;
  let fileCounter = 0;
  const activeTasks = new Map();
  let queuedTasks = [];

  function updateTransferStats(primary, loaded, total) {
    if (!primary.transferStats) {
      primary.transferStats = {
        start: Date.now(),
        lastLoaded: loaded,
        lastTime: Date.now(),
        speed: 0,
      };
      return;
    }

    const stats = primary.transferStats;
    const now = Date.now();
    const elapsed = now - stats.lastTime;

    if (elapsed < 500) {
      return;
    }

    const delta = loaded - stats.lastLoaded;
    stats.speed = delta / (elapsed / 1000);
    stats.lastLoaded = loaded;
    stats.lastTime = now;
  }

  function formatBytes(bytes) {
    if (bytes === 0) {
      return "0 B";
    }

    const units = ["B", "KB", "MB", "GB", "TB"];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));

    return (bytes / Math.pow(1024, i)).toFixed(i > 0 ? 1 : 0) + " " + units[i];
  }

  function formatSpeed(bytesPerSec) {
    if (bytesPerSec <= 0) {
      return "";
    }

    return formatBytes(bytesPerSec) + "/s";
  }

  function formatETA(remaining, speed) {
    if (speed <= 0 || remaining <= 0) {
      return "";
    }

    const secs = remaining / speed;

    if (secs < 60) {
      return Math.round(secs) + "s";
    }

    if (secs < 3600) {
      return Math.round(secs / 60) + "m";
    }

    return Math.round(secs / 3600) + "h";
  }

  function updateMeta(filename, loaded, total, transferStats) {
    if (!metaTitle || !metaDetail) {
      return;
    }

    metaTitle.textContent = filename;

    let detail = formatBytes(loaded) + " / " + formatBytes(total);

    if (transferStats && transferStats.speed > 0) {
      const remaining = total - loaded;
      detail +=
        " - " +
        formatSpeed(transferStats.speed) +
        " - ETA " +
        formatETA(remaining, transferStats.speed);
    }

    metaDetail.textContent = detail;
  }

  function clearMeta() {
    if (metaTitle) {
      metaTitle.textContent = "";
    }

    if (metaDetail) {
      metaDetail.textContent = "";
    }

    if (metaTooltip) {
      metaTooltip.classList.add("is-hidden");
    }
  }

  function setStatus(text) {
    if (status) {
      status.textContent = text;
    }
  }

  function toastUploadSuccess(filename) {
    if (window.Toast) {
      window.Toast.success(filename + " uploaded.", {
        title: "Upload complete",
      });
    }
  }

  function toastUploadError(filename) {
    if (window.Toast) {
      window.Toast.error(filename + " failed to upload.", {
        title: "Upload failed",
      });
    }
  }

  function toastUploadInfo(message, title) {
    if (window.Toast) {
      window.Toast.info(message, {
        title: title || "Upload",
      });
    }
  }

  function updateTaskUI(task) {
    const el = document.getElementById("upload-task-" + task.id);

    if (!el) {
      return;
    }

    const nameEl = el.querySelector(".queue-item-name");
    const badgeEl = el.querySelector(".queue-item-badge");
    const fillEl = el.querySelector(".queue-item-fill");
    const progressEl = el.querySelector(".queue-item-progress");
    const detailEl = el.querySelector(".queue-item-detail");
    const speedEl = el.querySelector(".queue-item-speed");

    if (nameEl) {
      nameEl.textContent = task.file.name;
    }

    const pct =
      task.totalBytes > 0
        ? Math.round(((task.uploadedBytes || 0) / task.totalBytes) * 100)
        : 0;

    if (fillEl) {
      fillEl.style.width = pct + "%";
    }

    if (progressEl) {
      progressEl.textContent =
        pct +
        "% • " +
        formatBytes(task.uploadedBytes || 0) +
        " / " +
        formatBytes(task.totalBytes || task.file.size);
    }

    if (detailEl) {
      detailEl.textContent = queueDetailText(task);
      detailEl.classList.toggle("is-hidden", !task.errorMessage);
    }

    if (speedEl) {
      if (
        task.transferStats &&
        task.transferStats.speed > 0 &&
        task.status === "uploading"
      ) {
        const remaining =
          (task.totalBytes || task.file.size) - (task.uploadedBytes || 0);

        speedEl.textContent =
          formatSpeed(task.transferStats.speed) +
          " • " +
          formatETA(remaining, task.transferStats.speed) +
          " remaining";
      } else {
        speedEl.textContent = queueMetaText(task);
      }
    }

    if (badgeEl) {
      badgeEl.textContent = queueBadgeText(task);
      badgeEl.className = "queue-item-badge is-" + task.status;
    }

    el.className =
      "queue-item" +
      (task.status === "complete" ? " is-complete" : "") +
      (task.status === "error" || task.status === "cancelled"
        ? " is-error"
        : "");
  }

  function updateBatchUI(batchId) {
    if (!batchId) {
      return;
    }

    const items = document.querySelectorAll(
      '.queue-item[data-batch="' + batchId + '"]',
    );

    let complete = 0;
    const total = items.length;

    for (let i = 0; i < items.length; i++) {
      if (
        items[i].classList.contains("is-complete") ||
        items[i].classList.contains("is-error")
      ) {
        complete++;
      }
    }

    const batchEl = document.getElementById("upload-batch-" + batchId);

    if (batchEl) {
      const progressEl = batchEl.querySelector(".upload-batch-progress");

      if (progressEl) {
        progressEl.textContent = complete + " / " + total;
      }
    }
  }

  function updateQueueMeta() {
    let activeCount = 0;
    let queuedCount = 0;
    let visibleCount = 0;

    activeTasks.forEach(function (task) {
      if (task.status === "uploading") {
        activeCount++;
      }

      if (task.status === "uploading" || task.status === "error") {
        visibleCount++;
      }
    });

    for (let i = 0; i < queuedTasks.length; i++) {
      if (queuedTasks[i].status === "queued") {
        queuedCount++;
        visibleCount++;
      }
    }

    if (queueMeta) {
      if (activeCount > 0) {
        queueMeta.textContent =
          activeCount + " item" + (activeCount > 1 ? "s" : "") + " active";
      } else if (queuedCount > 0) {
        queueMeta.textContent = queuedCount + " queued";
      } else {
        queueMeta.textContent = "0 items active";
      }
    }

    if (queueEmpty) {
      queueEmpty.classList.toggle("is-hidden", visibleCount > 0);
    }
  }

  function removeBatchHeader(batchId) {
    if (!batchId) {
      return;
    }

    if (document.querySelector('.queue-item[data-batch="' + batchId + '"]')) {
      return;
    }

    const header = document.getElementById("upload-batch-" + batchId);

    if (header && header.parentNode) {
      header.parentNode.removeChild(header);
    }
  }

  function removeQueueItem(taskId) {
    const el = document.getElementById("upload-task-" + taskId);

    if (el) {
      const batchId = el.getAttribute("data-batch");
      el.parentNode.removeChild(el);
      removeBatchHeader(batchId);
    }
  }

  function purgeTask(task) {
    if (!task) {
      return;
    }

    activeTasks.delete(task.id);
    queuedTasks = queuedTasks.filter(function (item) {
      return item.id !== task.id;
    });

    if (primaryTaskId === task.id) {
      setPrimaryTask(null);
    }

    removeQueueItem(task.id);
    updateBatchUI(task.batch);
    updateQueueMeta();
  }

  function scheduleTaskRemoval(task, delay) {
    if (!task || task.removeTimer) {
      return;
    }

    task.removeTimer = window.setTimeout(function () {
      task.removeTimer = null;
      purgeTask(task);
    }, delay);
  }

  function uploadErrorMessage(err, fallback) {
    if (err && err.data) {
      if (typeof err.data.error === "string" && err.data.error) {
        return err.data.error;
      }

      if (typeof err.data.message === "string" && err.data.message) {
        return err.data.message;
      }

      if (typeof err.data.detail === "string" && err.data.detail) {
        return err.data.detail;
      }
    }

    if (err && err.message && err.message !== "Request failed") {
      return err.message;
    }

    return fallback || "Upload failed. Try again.";
  }

  function addQueueItem(task) {
    const li = document.createElement("li");
    li.id = "upload-task-" + task.id;
    li.className = "queue-item";

    if (task.batch) {
      li.setAttribute("data-batch", task.batch);
    }

    const top = document.createElement("div");
    top.className = "queue-item-top";

    const fileWrap = document.createElement("div");
    fileWrap.className = "queue-item-file";

    const icon = document.createElement("span");
    icon.className = "queue-item-icon";
    icon.innerHTML = fileIconSVG(task.file);

    const nameEl = document.createElement("span");
    nameEl.className = "queue-item-name";
    nameEl.textContent = task.file.name;

    const badgeEl = document.createElement("span");
    badgeEl.className = "queue-item-badge is-queued";
    badgeEl.textContent = "Queued";

    fileWrap.appendChild(icon);
    fileWrap.appendChild(nameEl);
    fileWrap.appendChild(badgeEl);

    const actions = document.createElement("div");
    actions.className = "queue-item-actions";

    const cancelBtn = document.createElement("button");
    cancelBtn.className = "queue-item-action is-cancel";
    cancelBtn.type = "button";
    cancelBtn.setAttribute("data-queue-action", "cancel");
    cancelBtn.setAttribute("data-task-id", String(task.id));
    cancelBtn.setAttribute("aria-label", "Cancel upload");
    cancelBtn.innerHTML = closeIconSVG();

    actions.appendChild(cancelBtn);

    top.appendChild(fileWrap);
    top.appendChild(actions);

    const track = document.createElement("div");
    track.className = "queue-item-track";

    const fill = document.createElement("span");
    fill.className = "queue-item-fill";

    track.appendChild(fill);

    const meta = document.createElement("div");
    meta.className = "queue-item-meta";

    const progressEl = document.createElement("span");
    progressEl.className = "queue-item-progress mono";
    progressEl.textContent = "0% • 0 B / " + formatBytes(task.file.size);

    const speedEl = document.createElement("span");
    speedEl.className = "queue-item-speed mono";
    speedEl.textContent = "Waiting to start";

    const detailEl = document.createElement("span");
    detailEl.className = "queue-item-detail is-hidden";

    meta.appendChild(progressEl);
    meta.appendChild(speedEl);

    li.appendChild(top);
    li.appendChild(track);
    li.appendChild(meta);
    li.appendChild(detailEl);

    queueList.appendChild(li);
  }

  function addBatchHeader(batchId) {
    const header = document.createElement("li");
    header.id = "upload-batch-" + batchId;
    header.className = "queue-item queue-batch-header";

    const nameEl = document.createElement("span");
    nameEl.textContent = "Batch upload";

    const progressEl = document.createElement("span");
    progressEl.className = "upload-batch-progress";
    progressEl.textContent = "0 / 0";

    header.appendChild(nameEl);
    header.appendChild(progressEl);

    queueList.appendChild(header);
  }

  function updateTaskProgress(task, loaded, total) {
    task.uploadedBytes = loaded;
    task.totalBytes = total;

    updateTransferStats(task, loaded, total);
    updateTaskUI(task);
    updateBatchUI(task.batch);
    updateQueueMeta();
  }

  function setPrimaryTask(task) {
    primaryTaskId = task ? task.id : null;
    updatePrimaryUI();
  }

  function updatePrimaryUI() {
    const primary = primaryTaskId ? activeTasks.get(primaryTaskId) : null;

    if (!primary) {
      clearMeta();
      setStatus("No uploads in progress.");
      return;
    }

    const file = primary.file;
    const uploadedBytes = primary.uploadedBytes || 0;
    const totalBytes = primary.totalBytes || file.size;

    updateTransferStats(primary, uploadedBytes, totalBytes);
    updateMeta(file.name, uploadedBytes, totalBytes, primary.transferStats);
    setStatus(primary.statusText || "Uploading " + file.name + "...");
  }

  function updatePrimaryProgress() {
    const primary = primaryTaskId ? activeTasks.get(primaryTaskId) : null;

    if (!primary) {
      return;
    }

    const uploadedBytes = primary.uploadedBytes || 0;
    const totalBytes = primary.totalBytes || primary.file.size;

    updateTransferStats(primary, uploadedBytes, totalBytes);
    updateMeta(
      primary.file.name,
      uploadedBytes,
      totalBytes,
      primary.transferStats,
    );
  }

  function nextTask() {
    if (queuedTasks.length === 0) {
      return null;
    }

    let activeCount = 0;

    activeTasks.forEach(function (t) {
      if (t.status === "uploading") {
        activeCount++;
      }
    });

    if (activeCount >= MAX_CONCURRENCY) {
      return null;
    }

    for (let i = 0; i < queuedTasks.length; i++) {
      const task = queuedTasks[i];

      if (task.status === "queued") {
        return task;
      }
    }

    return null;
  }

  function processQueue() {
    let task;

    while ((task = nextTask()) !== null) {
      startTask(task);
    }
  }

  async function cleanupFailure(task, err) {
    uploadPipeline.abort(task);

    task.status = "error";
    task.statusText = "Upload failed.";
    task.errorMessage = uploadErrorMessage(
      err,
      "Upload failed. Check the file and try again.",
    );

    updateTaskUI(task);
    updateBatchUI(task.batch);
    updateQueueMeta();

    if (!task.uploadSessionId) {
      return;
    }

    try {
      await uploadPipeline.cancelSession(task.uploadSessionId);
    } catch (_) {}
  }

  async function uploadFile(task) {
    if (task.file.size > MAX_FILE_SIZE) {
      task.status = "error";
      task.statusText = "File exceeds the maximum upload size.";
      task.errorMessage = "This file exceeds the current maximum upload size.";

      updateTaskUI(task);
      updateBatchUI(task.batch);
      updateQueueMeta();

      return;
    }

    try {
      await uploadPipeline.upload(task);

      task.status = "complete";
      task.statusText = "Upload complete: " + task.file.name;
      task.errorMessage = "";

      updateTaskProgress(task, task.file.size, task.file.size);
      toastUploadSuccess(task.file.name);
    } catch (err) {
      task.lastError = err;

      if (task.status !== "error" && task.status !== "cancelled") {
        await cleanupFailure(task, err);
      }

      throw err;
    }
  }

  function startTask(task) {
    task.status = "uploading";
    task.statusText = "Uploading " + task.file.name + "...";

    activeTasks.set(task.id, task);

    if (!primaryTaskId) {
      setPrimaryTask(task);
    }

    updateTaskUI(task);
    updateBatchUI(task.batch);
    updateQueueMeta();

    uploadFile(task)
      .catch(function () {
        if (task.status !== "cancelled") {
          toastUploadError(task.file.name);
        }
      })
      .finally(function () {
        task.done = true;

        if (task.status === "complete") {
          scheduleTaskRemoval(task, AUTO_CLEAR_SUCCESS_MS);
        } else if (task.status === "cancelled") {
          scheduleTaskRemoval(task, AUTO_CLEAR_CANCELLED_MS);
        }

        if (primaryTaskId === task.id) {
          setPrimaryTask(null);

          activeTasks.forEach(function (t) {
            if (!t.done && t.status === "uploading") {
              setPrimaryTask(t);
            }
          });
        }

        processQueue();
        updateBatchUI(task.batch);
        updateQueueMeta();
      });
  }

  function enqueueFile(file, batchId) {
    if (queuedTasks.length >= MAX_QUEUE_ITEMS) {
      toastUploadError("Queue limit reached");
      return;
    }

    const task = {
      id: ++fileCounter,
      file: file,
      batch: batchId || null,
      chunkSize: resolveUploadChunkSize(file.size),
      status: "queued",
      statusText: "Queued: " + file.name,
      uploadedBytes: 0,
      totalBytes: file.size,
    };

    queuedTasks.push(task);
    addQueueItem(task);
    updateQueueMeta();
    processQueue();
  }

  function enqueueFiles(files) {
    const batchId = files.length > 1 ? ++batchCounter : null;

    if (batchId) {
      addBatchHeader(batchId);
    }

    for (let i = 0; i < files.length; i++) {
      enqueueFile(files[i], batchId);
    }

    updateBatchUI(batchId);
  }

  function handleFiles(files) {
    const fileArray = [];

    for (let i = 0; i < files.length; i++) {
      fileArray.push(files[i]);
    }

    enqueueFiles(fileArray);
  }

  function abortAll() {
    activeTasks.forEach(function (task) {
      if (task.status === "uploading") {
        task.canceled = true;

        uploadPipeline.abort(task);

        task.status = "cancelled";
        task.statusText = "Upload cancelled.";
        task.errorMessage = "";

        updateTaskUI(task);
        scheduleTaskRemoval(task, AUTO_CLEAR_CANCELLED_MS);
        updateQueueMeta();
      }
    });

    for (let i = 0; i < queuedTasks.length; i++) {
      const task = queuedTasks[i];

      task.status = "cancelled";
      task.statusText = "Cancelled.";
      task.errorMessage = "";

      updateTaskUI(task);
      scheduleTaskRemoval(task, AUTO_CLEAR_CANCELLED_MS);
    }

    primaryTaskId = null;

    clearMeta();
    setStatus("All uploads cancelled.");
    updateQueueMeta();
  }

  if (browseFilesButton && input) {
    browseFilesButton.addEventListener("click", function () {
      input.click();
    });
  }

  if (input) {
    input.addEventListener("change", function () {
      if (input.files && input.files.length) {
        handleFiles(input.files);
        input.value = "";
      }
    });
  }

  if (dropzone) {
    dropzone.addEventListener("dragover", function (e) {
      e.preventDefault();
      dropzone.classList.add("is-dragover");
    });

    dropzone.addEventListener("dragleave", function () {
      dropzone.classList.remove("is-dragover");
    });

    dropzone.addEventListener("drop", function (e) {
      e.preventDefault();
      dropzone.classList.remove("is-dragover");

      if (e.dataTransfer.files && e.dataTransfer.files.length) {
        handleFiles(e.dataTransfer.files);
      }
    });
  }

  if (confirmStart && confirmCancel && confirmBackdrop) {
    confirmStart.addEventListener("click", function () {
      confirmBackdrop.classList.add("is-hidden");

      if (startButton) {
        startButton.classList.remove("is-hidden");
      }
    });

    confirmCancel.addEventListener("click", function () {
      confirmBackdrop.classList.add("is-hidden");
    });
  }

  if (abortButton) {
    abortButton.addEventListener("click", abortAll);
  }

  if (queueList) {
    queueList.addEventListener("click", function (event) {
      const target = event.target;

      if (!target || !target.closest) {
        return;
      }

      const button = target.closest("[data-queue-action]");

      if (!button) {
        return;
      }

      const taskId = Number(button.getAttribute("data-task-id"));

      if (!taskId) {
        return;
      }

      let task = activeTasks.get(taskId);

      if (!task) {
        for (let i = 0; i < queuedTasks.length; i++) {
          if (queuedTasks[i].id === taskId) {
            task = queuedTasks[i];
            break;
          }
        }
      }

      if (!task) {
        return;
      }

      if (task.status === "uploading") {
        task.canceled = true;

        uploadPipeline.abort(task);

        task.status = "cancelled";
        task.statusText = "Upload cancelled.";
        task.errorMessage = "";

        updateTaskUI(task);
        updateQueueMeta();

        return;
      }

      if (task.status === "queued") {
        task.status = "cancelled";
        task.statusText = "Cancelled.";
        task.errorMessage = "";

        purgeTask(task);

        return;
      }

      purgeTask(task);
    });
  }

  function queueBadgeText(task) {
    switch (task.status) {
      case "uploading":
        return "Encrypting...";
      case "complete":
        return "Secured";
      case "error":
        return "Failed";
      case "cancelled":
        return "Cancelled";
      default:
        return "Queued";
    }
  }

  function queueMetaText(task) {
    switch (task.status) {
      case "complete":
        return "Transfer complete";
      case "error":
        return "Upload failed";
      case "cancelled":
        return "Upload cancelled";
      case "queued":
        return "Waiting to start";
      default:
        return "Preparing transfer";
    }
  }

  function queueDetailText(task) {
    return task.errorMessage || "";
  }

  function svgIcon(path) {
    return (
      '<svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">' +
      path +
      "</svg>"
    );
  }

  function closeIconSVG() {
    return svgIcon(
      '<path d="M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>' +
        '<path d="M6 6L18 18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>',
    );
  }

  function fileIconSVG(file) {
    const type = (file.type || "").toLowerCase();

    if (type.indexOf("zip") >= 0 || file.name.match(/\.(zip|tar|gz|rar)$/i)) {
      return svgIcon(
        '<path d="M4 7V5C4 3.89543 4.89543 3 6 3H14L20 9V19C20 20.1046 19.1046 21 18 21H6C4.89543 21 4 20.1046 4 19V7Z" stroke="currentColor" stroke-width="2"/>' +
          '<path d="M14 3V9H20" stroke="currentColor" stroke-width="2"/>' +
          '<path d="M10 12H14" stroke="currentColor" stroke-width="2"/>' +
          '<path d="M10 16H14" stroke="currentColor" stroke-width="2"/>',
      );
    }

    if (type.indexOf("pdf") >= 0) {
      return svgIcon(
        '<path d="M4 7V5C4 3.89543 4.89543 3 6 3H14L20 9V19C20 20.1046 19.1046 21 18 21H6C4.89543 21 4 20.1046 4 19V7Z" stroke="currentColor" stroke-width="2"/>' +
          '<path d="M14 3V9H20" stroke="currentColor" stroke-width="2"/>' +
          '<path d="M8 15H16" stroke="currentColor" stroke-width="2"/>',
      );
    }

    if (
      type.indexOf("text/") === 0 ||
      type.indexOf("json") >= 0 ||
      type.indexOf("xml") >= 0 ||
      file.name.match(/\.(sh|js|ts|json|xml|txt|md)$/i)
    ) {
      return svgIcon(
        '<path d="M4 7V5C4 3.89543 4.89543 3 6 3H14L20 9V19C20 20.1046 19.1046 21 18 21H6C4.89543 21 4 20.1046 4 19V7Z" stroke="currentColor" stroke-width="2"/>' +
          '<path d="M14 3V9H20" stroke="currentColor" stroke-width="2"/>' +
          '<path d="M9 14L7 16L9 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>' +
          '<path d="M15 14L17 16L15 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>',
      );
    }

    return svgIcon(
      '<path d="M4 7V5C4 3.89543 4.89543 3 6 3H14L20 9V19C20 20.1046 19.1046 21 18 21H6C4.89543 21 4 20.1046 4 19V7Z" stroke="currentColor" stroke-width="2"/>' +
        '<path d="M14 3V9H20" stroke="currentColor" stroke-width="2"/>',
    );
  }
}
