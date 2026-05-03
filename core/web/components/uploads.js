(function() {
  const input = document.getElementById("upload-file");
  const browseFilesButton = document.getElementById("upload-browse-files");
  const startButton = document.getElementById("upload-start");
  const abortButton = document.getElementById("upload-abort");
  const dropzone = document.getElementById("upload-dropzone");
  const chip = document.getElementById("upload-chip");
  const chipName = document.getElementById("upload-chip-name");
  const chipSize = document.getElementById("upload-chip-size");
  const chipClear = document.getElementById("upload-chip-clear");
  const confirmBackdrop = document.getElementById("upload-confirm-backdrop");
  const confirmMeta = document.getElementById("upload-confirm-meta");
  const confirmStart = document.getElementById("upload-confirm-start");
  const confirmCancel = document.getElementById("upload-confirm-cancel");
  const controls = document.getElementById("upload-controls");
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
  const MAX_CONCURRENCY = 10;

  let primaryTaskId = null;
  let batchCounter = 0;
  let fileCounter = 0;
  let activeTasks = new Map();
  let queuedTasks = [];

  const STORAGE_PREFIX = "upload:";

  function fileSignature(file) {
    return file.name + ":" + file.size + ":" + (file.lastModified || 0);
  }

  function storageKey(sig) {
    return STORAGE_PREFIX + sig;
  }

  function saveState(sig, state) {
    try {
      localStorage.setItem(storageKey(sig), JSON.stringify(state));
    } catch (_) {}
  }

  function loadState(sig) {
    try {
      var raw = localStorage.getItem(storageKey(sig));
      return raw ? JSON.parse(raw) : null;
    } catch (_) {
      return null;
    }
  }

  function clearState(sig) {
    try {
      localStorage.removeItem(storageKey(sig));
    } catch (_) {}
  }

  function clearStateByUploadId(uploadId) {
    try {
      for (var i = 0; i < localStorage.length; i++) {
        var key = localStorage.key(i);
        if (key.indexOf(STORAGE_PREFIX) === 0) {
          try {
            var state = JSON.parse(localStorage.getItem(key));
            if (state && state.uploadId === uploadId) {
              localStorage.removeItem(key);
            }
          } catch (_) {}
        }
      }
    } catch (_) {}
  }

  function updateTransferStats(primary, loaded, total) {
    if (!primary.transferStats) {
      primary.transferStats = { start: Date.now(), lastLoaded: loaded, lastTime: Date.now(), speed: 0 };
      return;
    }
    var stats = primary.transferStats;
    var now = Date.now();
    var elapsed = now - stats.lastTime;
    if (elapsed < 500) {
      return;
    }
    var delta = loaded - stats.lastLoaded;
    stats.speed = delta / (elapsed / 1000);
    stats.lastLoaded = loaded;
    stats.lastTime = now;
  }

  function formatBytes(bytes) {
    if (bytes === 0) {
      return "0 B";
    }
    var units = ["B", "KB", "MB", "GB", "TB"];
    var i = Math.floor(Math.log(bytes) / Math.log(1024));
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
    var secs = remaining / speed;
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
    var detail = formatBytes(loaded) + " / " + formatBytes(total);
    if (transferStats && transferStats.speed > 0) {
      var remaining = total - loaded;
      detail += " - " + formatSpeed(transferStats.speed) + " - ETA " + formatETA(remaining, transferStats.speed);
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
      window.Toast.success(filename + " uploaded.", { title: "Upload complete" });
    }
  }

  function toastUploadError(filename) {
    if (window.Toast) {
      window.Toast.error(filename + " failed to upload.", { title: "Upload failed" });
    }
  }

  function toastUploadInfo(message, title) {
    if (window.Toast) {
      window.Toast.info(message, { title: title || "Upload" });
    }
  }

  function updateTaskUI(task) {
    var el = document.getElementById("upload-task-" + task.id);
    if (!el) {
      return;
    }
    var nameEl = el.querySelector(".queue-item-name");
    var badgeEl = el.querySelector(".queue-item-badge");
    var fillEl = el.querySelector(".queue-item-fill");
    var progressEl = el.querySelector(".queue-item-progress");
    var detailEl = el.querySelector(".queue-item-detail");
    var speedEl = el.querySelector(".queue-item-speed");
    if (nameEl) {
      nameEl.textContent = task.file.name;
    }
    var pct = task.totalBytes > 0 ? Math.round((task.uploadedBytes || 0) / task.totalBytes * 100) : 0;
    if (fillEl) {
      fillEl.style.width = pct + "%";
    }
    if (progressEl) {
      progressEl.textContent = pct + "% • " + formatBytes(task.uploadedBytes || 0) + " / " + formatBytes(task.totalBytes || task.file.size);
    }
    if (detailEl) {
      detailEl.textContent = queueBadgeText(task);
    }
    if (speedEl) {
      if (task.transferStats && task.transferStats.speed > 0 && task.status === "uploading") {
        var remaining = (task.totalBytes || task.file.size) - (task.uploadedBytes || 0);
        speedEl.textContent = formatSpeed(task.transferStats.speed) + " • " + formatETA(remaining, task.transferStats.speed) + " remaining";
      } else {
        speedEl.textContent = queueMetaText(task);
      }
    }
    if (badgeEl) {
      badgeEl.textContent = queueBadgeText(task);
      badgeEl.className = "queue-item-badge is-" + task.status;
    }
    el.className = "queue-item" + (task.status === "complete" ? " is-complete" : "") + (task.status === "error" || task.status === "cancelled" ? " is-error" : "");
  }

  function updateBatchUI(batchId) {
    var items = document.querySelectorAll(".queue-item[data-batch=\"" + batchId + "\"]");
    var complete = 0;
    var total = items.length;
    for (var i = 0; i < items.length; i++) {
      if (items[i].classList.contains("is-complete") || items[i].classList.contains("is-error")) {
        complete++;
      }
    }
    var batchEl = document.getElementById("upload-batch-" + batchId);
    if (batchEl) {
      var progressEl = batchEl.querySelector(".upload-batch-progress");
      if (progressEl) {
        progressEl.textContent = complete + " / " + total;
      }
    }
  }

  function updateQueueMeta() {
    var activeCount = 0;
    var queuedCount = 0;
    activeTasks.forEach(function(task) {
      if (task.status === "uploading") {
        activeCount++;
      }
    });
    for (var i = 0; i < queuedTasks.length; i++) {
      if (queuedTasks[i].status === "queued") {
        queuedCount++;
      }
    }
    if (queueMeta) {
      if (activeCount > 0) {
        queueMeta.textContent = activeCount + " item" + (activeCount > 1 ? "s" : "") + " active";
      } else if (queuedCount > 0) {
        queueMeta.textContent = queuedCount + " queued";
      } else {
        queueMeta.textContent = "0 items active";
      }
    }
    if (queueEmpty) {
      queueEmpty.classList.toggle("is-hidden", activeTasks.size > 0 || queuedTasks.length > 0);
    }
  }

  function removeQueueItem(taskId) {
    var el = document.getElementById("upload-task-" + taskId);
    if (el) {
      el.parentNode.removeChild(el);
    }
  }

  function addQueueItem(task) {
    var li = document.createElement("li");
    li.id = "upload-task-" + task.id;
    li.className = "queue-item";
    if (task.batch) {
      li.setAttribute("data-batch", task.batch);
    }

    var top = document.createElement("div");
    top.className = "queue-item-top";

    var fileWrap = document.createElement("div");
    fileWrap.className = "queue-item-file";

    var icon = document.createElement("span");
    icon.className = "queue-item-icon";
    icon.innerHTML = fileIconSVG(task.file);

    var nameEl = document.createElement("span");
    nameEl.className = "queue-item-name";
    nameEl.textContent = task.file.name;

    var badgeEl = document.createElement("span");
    badgeEl.className = "queue-item-badge is-queued";
    badgeEl.textContent = "Queued";

    fileWrap.appendChild(icon);
    fileWrap.appendChild(nameEl);
    fileWrap.appendChild(badgeEl);

    var actions = document.createElement("div");
    actions.className = "queue-item-actions";

    var cancelBtn = document.createElement("button");
    cancelBtn.className = "queue-item-action is-cancel";
    cancelBtn.type = "button";
    cancelBtn.setAttribute("data-queue-action", "cancel");
    cancelBtn.setAttribute("data-task-id", String(task.id));
    cancelBtn.setAttribute("aria-label", "Cancel upload");
    cancelBtn.innerHTML = closeIconSVG();

    actions.appendChild(cancelBtn);

    top.appendChild(fileWrap);
    top.appendChild(actions);

    var track = document.createElement("div");
    track.className = "queue-item-track";
    var fill = document.createElement("span");
    fill.className = "queue-item-fill";
    track.appendChild(fill);

    var meta = document.createElement("div");
    meta.className = "queue-item-meta";

    var progressEl = document.createElement("span");
    progressEl.className = "queue-item-progress mono";
    progressEl.textContent = "0% • 0 B / " + formatBytes(task.file.size);

    var speedEl = document.createElement("span");
    speedEl.className = "queue-item-speed mono";
    speedEl.textContent = "Waiting to start";

    var detailEl = document.createElement("span");
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
    var header = document.createElement("li");
    header.id = "upload-batch-" + batchId;
    header.className = "queue-item queue-batch-header";
    var nameEl = document.createElement("span");
    nameEl.textContent = "Batch upload";
    var progressEl = document.createElement("span");
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
    var primary = primaryTaskId ? activeTasks.get(primaryTaskId) : null;
    if (!primary) {
      clearMeta();
      setStatus("No uploads in progress.");
      return;
    }
    var file = primary.file;
    var uploadedBytes = primary.uploadedBytes || 0;
    var totalBytes = primary.totalBytes || file.size;
    updateTransferStats(primary, uploadedBytes, totalBytes);
    updateMeta(file.name, uploadedBytes, totalBytes, primary.transferStats);
    setStatus(primary.statusText || ("Uploading " + file.name + "..."));
  }

  function updatePrimaryProgress() {
    var primary = primaryTaskId ? activeTasks.get(primaryTaskId) : null;
    if (!primary) {
      return;
    }
    var uploadedBytes = primary.uploadedBytes || 0;
    var totalBytes = primary.totalBytes || primary.file.size;
    updateTransferStats(primary, uploadedBytes, totalBytes);
    updateMeta(primary.file.name, uploadedBytes, totalBytes, primary.transferStats);
  }

  function nextTask() {
    if (queuedTasks.length === 0) {
      return null;
    }
    var activeCount = 0;
    activeTasks.forEach(function(t) {
      if (t.status === "uploading") {
        activeCount++;
      }
    });
    if (activeCount >= MAX_CONCURRENCY) {
      return null;
    }
    for (var i = 0; i < queuedTasks.length; i++) {
      var task = queuedTasks[i];
      if (task.status === "queued") {
        return task;
      }
    }
    return null;
  }

  function processQueue() {
    var task;
    while ((task = nextTask()) !== null) {
      startTask(task);
    }
  }

  async function api(path, body, method) {
    var headers = { "Content-Type": "application/json" };
    if (method === "GET") {
      headers = {};
    }
    var res = await fetch(path, {
      method: method || "POST",
      headers: headers,
      body: body ? JSON.stringify(body) : undefined
    });
    if (res.status === 204) {
      return null;
    }
    var data = await res.json();
    if (!res.ok) {
      var err = new Error("Request failed");
      err.status = res.status;
      err.data = data;
      throw err;
    }
    return data;
  }

  function startUpload(task) {
    return api("/api/uploads/start", {
      filename: task.file.name,
      size: task.file.size,
      contentType: task.file.type || "application/octet-stream"
    });
  }

  function completeUpload(fileId) {
    return api("/api/uploads/" + encodeURIComponent(fileId) + "/complete", {});
  }

  function cancelUpload(fileId) {
    return api("/api/uploads/" + encodeURIComponent(fileId) + "/cancel", {});
  }

  function uploadSingleWithProgress(task, url, file) {
    return new Promise(function(resolve, reject) {
      var xhr = new XMLHttpRequest();
      task.xhr = xhr;
      xhr.open("PUT", url);
      xhr.upload.onprogress = function(e) {
        if (e.lengthComputable) {
          task.uploadedBytes = e.loaded;
          task.totalBytes = e.total;
          updateTaskProgress(task, e.loaded, e.total);
          updatePrimaryProgress();
        }
      };
      xhr.onload = function() {
        if (xhr.status >= 200 && xhr.status < 300) {
          resolve();
        } else {
          var err = new Error("Upload failed: " + xhr.status);
          err.cancelled = xhr.status === 409;
          reject(err);
        }
      };
      xhr.onerror = function() {
        reject(new Error("Network error"));
      };
      xhr.onabort = function() {
        var err = new Error("Upload cancelled");
        err.cancelled = true;
        reject(err);
      };
      xhr.send(file);
    });
  }

  async function uploadSingle(task, initialResponse) {
    var file = task.file;
    var signature = fileSignature(file);
    var resp = initialResponse;
    try {
      if (!resp) {
        resp = await startUpload(task);
      }
      task.uploadId = resp.uploadId;
      task.fileId = resp.fileId;
      task.statusText = "Uploading " + file.name + "...";
      saveState(signature, {
        uploadId: resp.uploadId,
        fileId: resp.fileId,
        filename: file.name,
        sizeBytes: file.size,
        status: "uploading"
      });

      if (!resp.uploadUrl) {
        throw new Error("Missing upload URL");
      }

      await uploadSingleWithProgress(task, resp.uploadUrl, file);
      await completeUpload(resp.fileId);
      clearState(signature);
      task.status = "complete";
      task.statusText = "Upload complete: " + file.name;
      updateTaskProgress(task, file.size, file.size);
      toastUploadSuccess(file.name);
    } catch (err) {
      if (err && err.cancelled) {
        clearState(signature);
        task.status = "cancelled";
        task.statusText = "Upload cancelled.";
        toastUploadInfo("Upload cancelled.", "Cancelled");
        return;
      }
      await cleanupFailure(task, resp ? resp.fileId : null, signature);
      throw err;
    }
  }

  async function cleanupFailure(task, fileId, signature) {
    if (signature) {
      clearState(signature);
    }
    if (fileId) {
      clearStateByUploadId(fileId);
    }
    if (task.xhr) {
      task.xhr.abort();
    }
    task.status = "error";
    task.statusText = "Upload failed.";
    updateTaskUI(task);
    if (!fileId) {
      return;
    }
    try {
      await cancelUpload(fileId);
    } catch (_) {}
  }

  async function uploadFile(task) {
    if (task.file.size > MAX_FILE_SIZE) {
      task.status = "error";
      task.statusText = "File exceeds the maximum upload size.";
      updateTaskUI(task);
      updateBatchUI(task.batch);
      updateQueueMeta();
      return;
    }
    var resp = await startUpload(task);
    await uploadSingle(task, resp);
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
      .catch(function(err) {
        if (task.status !== "cancelled") {
          toastUploadError(task.file.name);
        }
      })
      .finally(function() {
        task.done = true;
        if (primaryTaskId === task.id) {
          setPrimaryTask(null);
          activeTasks.forEach(function(t) {
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

  function enqueueFile(file) {
    if (queuedTasks.length >= MAX_QUEUE_ITEMS) {
      toastUploadError("Queue limit reached");
      return;
    }
    var task = {
      id: ++fileCounter,
      file: file,
      status: "queued",
      statusText: "Queued: " + file.name,
      uploadedBytes: 0,
      totalBytes: file.size
    };
    queuedTasks.push(task);
    addQueueItem(task);
    updateQueueMeta();
    processQueue();
  }

  function enqueueFiles(files) {
    var batchId = ++batchCounter;
    if (files.length > 1) {
      addBatchHeader(batchId);
    }
    for (var i = 0; i < files.length; i++) {
      enqueueFile(files[i]);
    }
    if (files.length > 1) {
    }
  }

  function handleFiles(files) {
    var fileArray = [];
    for (var i = 0; i < files.length; i++) {
      fileArray.push(files[i]);
    }
    enqueueFiles(fileArray);
  }

  function abortAll() {
    activeTasks.forEach(function(task) {
      if (task.status === "uploading") {
        task.canceled = true;
        if (task.xhr) {
          task.xhr.abort();
        }
        task.status = "cancelled";
        task.statusText = "Upload cancelled.";
        clearStateByUploadId(task.uploadId);
        updateTaskUI(task);
        updateQueueMeta();
      }
    });
    for (var i = 0; i < queuedTasks.length; i++) {
      var task = queuedTasks[i];
      task.status = "cancelled";
      task.statusText = "Cancelled.";
      updateTaskUI(task);
    }
    queuedTasks = [];
    activeTasks.clear();
    primaryTaskId = null;
    clearMeta();
    setStatus("All uploads cancelled.");
    updateQueueMeta();
  }

  function clearCompleted() {
    var els = queueList.querySelectorAll(".is-complete, .is-error");
    for (var i = 0; i < els.length; i++) {
      els[i].parentNode.removeChild(els[i]);
    }
    activeTasks.forEach(function(task, id) {
      if (task.status === "complete" || task.status === "error" || task.status === "cancelled") {
        activeTasks.delete(id);
      }
    });
    updateQueueMeta();
  }

  if (browseFilesButton && input) {
    browseFilesButton.addEventListener("click", function() {
      input.click();
    });
  }

  if (input) {
    input.addEventListener("change", function() {
      if (input.files && input.files.length) {
        handleFiles(input.files);
        input.value = "";
      }
    });
  }

  if (dropzone) {
    dropzone.addEventListener("dragover", function(e) {
      e.preventDefault();
      dropzone.classList.add("is-dragover");
    });
    dropzone.addEventListener("dragleave", function() {
      dropzone.classList.remove("is-dragover");
    });
    dropzone.addEventListener("drop", function(e) {
      e.preventDefault();
      dropzone.classList.remove("is-dragover");
      if (e.dataTransfer.files && e.dataTransfer.files.length) {
        handleFiles(e.dataTransfer.files);
      }
    });
  }

  if (confirmStart && confirmCancel && confirmBackdrop) {
    confirmStart.addEventListener("click", function() {
      confirmBackdrop.classList.add("is-hidden");
      if (startButton) {
        startButton.classList.remove("is-hidden");
      }
    });
    confirmCancel.addEventListener("click", function() {
      confirmBackdrop.classList.add("is-hidden");
    });
  }

  if (abortButton) {
    abortButton.addEventListener("click", abortAll);
  }

  if (queueList) {
    queueList.addEventListener("click", function(event) {
      var target = event.target;
      if (!target || !target.closest) {
        return;
      }
      var button = target.closest("[data-queue-action]");
      if (!button) {
        return;
      }
      var taskId = Number(button.getAttribute("data-task-id"));
      if (!taskId) {
        return;
      }
      var task = activeTasks.get(taskId);
      if (!task) {
        for (var i = 0; i < queuedTasks.length; i++) {
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
        if (task.xhr) {
          task.xhr.abort();
        }
        task.status = "cancelled";
        task.statusText = "Upload cancelled.";
        updateTaskUI(task);
        updateQueueMeta();
        return;
      }
      if (task.status === "queued") {
        task.status = "cancelled";
        task.statusText = "Cancelled.";
        removeQueueItem(taskId);
        queuedTasks = queuedTasks.filter(function(item) { return item.id !== taskId; });
        updateQueueMeta();
        return;
      }
      activeTasks.delete(taskId);
      removeQueueItem(taskId);
      updateQueueMeta();
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

  function svgIcon(path) {
    return '<svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">' + path + '</svg>';
  }

  function closeIconSVG() {
    return svgIcon('<path d="M18 6L6 18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/><path d="M6 6L18 18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>');
  }

  function fileIconSVG(file) {
    var type = (file.type || "").toLowerCase();
    if (type.indexOf("zip") >= 0 || file.name.match(/\.(zip|tar|gz|rar)$/i)) {
      return svgIcon('<path d="M4 7V5C4 3.89543 4.89543 3 6 3H14L20 9V19C20 20.1046 19.1046 21 18 21H6C4.89543 21 4 20.1046 4 19V7Z" stroke="currentColor" stroke-width="2"/><path d="M14 3V9H20" stroke="currentColor" stroke-width="2"/><path d="M10 12H14" stroke="currentColor" stroke-width="2"/><path d="M10 16H14" stroke="currentColor" stroke-width="2"/>');
    }
    if (type.indexOf("pdf") >= 0) {
      return svgIcon('<path d="M4 7V5C4 3.89543 4.89543 3 6 3H14L20 9V19C20 20.1046 19.1046 21 18 21H6C4.89543 21 4 20.1046 4 19V7Z" stroke="currentColor" stroke-width="2"/><path d="M14 3V9H20" stroke="currentColor" stroke-width="2"/><path d="M8 15H16" stroke="currentColor" stroke-width="2"/>');
    }
    if (type.indexOf("text/") === 0 || type.indexOf("json") >= 0 || type.indexOf("xml") >= 0 || file.name.match(/\.(sh|js|ts|json|xml|txt|md)$/i)) {
      return svgIcon('<path d="M4 7V5C4 3.89543 4.89543 3 6 3H14L20 9V19C20 20.1046 19.1046 21 18 21H6C4.89543 21 4 20.1046 4 19V7Z" stroke="currentColor" stroke-width="2"/><path d="M14 3V9H20" stroke="currentColor" stroke-width="2"/><path d="M9 14L7 16L9 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/><path d="M15 14L17 16L15 18" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>');
    }
    return svgIcon('<path d="M4 7V5C4 3.89543 4.89543 3 6 3H14L20 9V19C20 20.1046 19.1046 21 18 21H6C4.89543 21 4 20.1046 4 19V7Z" stroke="currentColor" stroke-width="2"/><path d="M14 3V9H20" stroke="currentColor" stroke-width="2"/>');
  }
})();
