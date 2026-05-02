(function() {
  const input = document.getElementById("upload-file");
  const folderInput = document.getElementById("upload-folder");
  const browseFilesButton = document.getElementById("upload-browse-files");
  const browseFoldersButton = document.getElementById("upload-browse-folders");
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

  function enableFolderPicker() {
    if (!folderInput) {
      return;
    }
    folderInput.setAttribute("webkitdirectory", "");
    folderInput.setAttribute("directory", "");
    folderInput.setAttribute("mozdirectory", "");
  }

  enableFolderPicker();
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
    var nameEl = el.querySelector(".upload-queue-item-name");
    var progressEl = el.querySelector(".upload-queue-item-progress");
    if (nameEl) {
      nameEl.textContent = task.statusText || task.file.name;
    }
    if (progressEl) {
      var pct = task.totalBytes > 0 ? Math.round((task.uploadedBytes || 0) / task.totalBytes * 100) : 0;
      progressEl.textContent = pct + "%";
    }
    el.className = "upload-queue-item" + (task.status === "complete" ? " is-complete" : "") + (task.status === "error" || task.status === "cancelled" ? " is-error" : "");
  }

  function updateBatchUI(batchId) {
    var items = document.querySelectorAll(".upload-queue-item[data-batch=\"" + batchId + "\"]");
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
    var totalBytes = 0;
    activeTasks.forEach(function(task) {
      if (task.status === "uploading") {
        activeCount++;
        totalBytes += task.file.size;
      }
    });
    if (queueMeta) {
      if (activeCount === 0) {
        queueMeta.textContent = queuedTasks.length + " files queued";
      } else {
        queueMeta.textContent = "Uploading " + activeCount + " file" + (activeCount > 1 ? "s" : "") + " - " + formatBytes(totalBytes);
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
    li.className = "upload-queue-item" + (task.batch ? " batch" : "");
    if (task.batch) {
      li.setAttribute("data-batch", task.batch);
    }
    var nameEl = document.createElement("span");
    nameEl.className = "upload-queue-item-name";
    nameEl.textContent = task.file.name;
    var progressEl = document.createElement("span");
    progressEl.className = "upload-queue-item-progress";
    progressEl.textContent = "0%";
    li.appendChild(nameEl);
    li.appendChild(progressEl);
    queueList.appendChild(li);
  }

  function addBatchHeader(batchId, folderPath) {
    var header = document.createElement("li");
    header.id = "upload-batch-" + batchId;
    header.className = "upload-batch-header";
    var nameEl = document.createElement("span");
    nameEl.textContent = folderPath || "Root";
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
      folderPath: task.folderPath || "",
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

  function enqueueFile(file, folderPath) {
    if (queuedTasks.length >= MAX_QUEUE_ITEMS) {
      toastUploadError("Queue limit reached");
      return;
    }
    var task = {
      id: ++fileCounter,
      file: file,
      folderPath: folderPath || "",
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

  function enqueueFiles(files, basePath) {
    var batchId = ++batchCounter;
    if (files.length > 1) {
      addBatchHeader(batchId, basePath || "");
    }
    for (var i = 0; i < files.length; i++) {
      enqueueFile(files[i], basePath || "");
    }
    if (files.length > 1) {
      var items = queueList.querySelectorAll(".upload-queue-item:last-child");
    }
  }

  function handleFiles(files, folderPath) {
    var fileArray = [];
    for (var i = 0; i < files.length; i++) {
      fileArray.push(files[i]);
    }
    enqueueFiles(fileArray, folderPath);
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

  if (browseFoldersButton && folderInput) {
    browseFoldersButton.addEventListener("click", function() {
      folderInput.click();
    });
  }

  if (input) {
    input.addEventListener("change", function() {
      if (input.files && input.files.length) {
        handleFiles(input.files, "");
        input.value = "";
      }
    });
  }

  if (folderInput) {
    folderInput.addEventListener("change", function() {
      if (folderInput.files && folderInput.files.length) {
        var path = folderInput.files[0].webkitRelativePath || "";
        var root = path.split("/")[0] || "";
        handleFiles(folderInput.files, root);
        folderInput.value = "";
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
      var entries = [];
      if (e.dataTransfer.items) {
        for (var i = 0; i < e.dataTransfer.items.length; i++) {
          var item = e.dataTransfer.items[i];
          if (item && item.webkitGetAsEntry) {
            var entry = item.webkitGetAsEntry();
            if (entry) {
              entries.push(entry);
            }
          }
        }
      }
      if (entries.length) {
        handleEntries(entries, "");
      } else if (e.dataTransfer.files && e.dataTransfer.files.length) {
        handleFiles(e.dataTransfer.files, "");
      }
    });
  }

  async function collectEntries(entry, collected) {
    if (entry.isFile) {
      var file = await new Promise(function(resolve) {
        entry.file(resolve);
      });
      collected.push(file);
    } else if (entry.isDirectory) {
      var reader = entry.createReader();
      var readAll = function() {
        reader.readEntries(async function(results) {
          if (results.length) {
            for (var i = 0; i < results.length; i++) {
              await collectEntries(results[i], collected);
            }
            readAll();
          }
        });
      };
      readAll();
    }
  }

  async function handleEntries(entries, basePath) {
    var collected = [];
    for (var i = 0; i < entries.length; i++) {
      await collectEntries(entries[i], collected);
    }
    if (collected.length) {
      enqueueFiles(collected, basePath);
    }
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
})();
