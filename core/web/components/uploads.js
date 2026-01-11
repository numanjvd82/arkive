(function() {
  const input = document.getElementById("upload-file");
  const folderInput = document.getElementById("upload-folder");
  const browseFilesButton = document.getElementById("upload-browse-files");
  const browseFoldersButton = document.getElementById("upload-browse-folders");
  const startButton = document.getElementById("upload-start");
  const pauseButton = document.getElementById("upload-pause");
  const resumeButton = document.getElementById("upload-resume");
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
  const isPremium = document.body && document.body.getAttribute("data-user-premium") === "true";
  const MAX_QUEUE_ITEMS = 300;
  const MAX_FILE_SIZE = 10 * 1024 * 1024 * 1024;
  const MULTIPART_THRESHOLD = 200 * 1024 * 1024;
  const LARGE_FILE_THROTTLE_MS = 100000;
  const MAX_CONCURRENCY = isPremium ? 10 : 1;

  let primaryTaskId = null;
  let batchCounter = 0;
  let fileCounter = 0;
  const batches = [];
  const queueFiles = [];
  const activeTasks = new Map();

  function setStatus(message) {
    status.textContent = message;
  }

  function setUploadInputEnabled(enabled) {
    if (input) {
      input.disabled = !enabled;
    }
    if (folderInput) {
      folderInput.disabled = !enabled;
    }
    if (dropzone) {
      dropzone.classList.toggle("is-disabled", !enabled);
      dropzone.setAttribute("aria-disabled", enabled ? "false" : "true");
      dropzone.setAttribute("tabindex", enabled ? "0" : "-1");
    }
  }

  function updateInputState() {
    const remaining = MAX_QUEUE_ITEMS - countQueuedFiles();
    setUploadInputEnabled(remaining > 0);
  }

  function toastUploadSuccess(filename) {
    const message = filename ? filename + " is ready to share." : "Your file is ready to share.";
    window.Toast.success(message, { title: "Uploaded" });
  }

  function toastUploadInfo(message, title) {
    window.Toast.info(message, { title: title || "Upload" });
  }

  function toastUploadError(detail) {
    const message = detail ? "Upload failed. " + detail : "Upload failed. Try again.";
    window.Toast.error(message, { title: "Upload failed" });
  }

  function formatBytes(bytes) {
    if (!bytes || bytes <= 0) {
      return "0 B";
    }
    const units = ["B", "KB", "MB", "GB"];
    let index = 0;
    let value = bytes;
    while (value >= 1024 && index < units.length - 1) {
      value /= 1024;
      index++;
    }
    return value.toFixed(value >= 100 ? 0 : 1) + " " + units[index];
  }

  function formatDuration(seconds) {
    if (!seconds || !isFinite(seconds) || seconds < 0) {
      return "";
    }
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    if (mins <= 0) {
      return secs + "s";
    }
    return mins + "m " + secs + "s";
  }

  function updateMeta(fileName, uploadedBytes, totalBytes, transferStats) {
    if (!metaTitle || !metaDetail || !metaTooltip) {
      return;
    }
    if (!fileName) {
      metaTitle.textContent = "";
      metaDetail.textContent = "";
      metaTooltip.setAttribute("data-tooltip", "");
      metaTooltip.classList.add("is-hidden");
      return;
    }
    const percent = totalBytes > 0 ? Math.round((uploadedBytes / totalBytes) * 100) : 0;
    const speed = transferStats && transferStats.speed ? formatBytes(transferStats.speed) + "/s" : "";
    const eta = transferStats && transferStats.eta ? "~" + formatDuration(transferStats.eta) + " left" : "Calculating ETA";
    metaTitle.textContent = fileName;
    metaDetail.textContent = eta;
    const speedText = speed || "Calculating...";
    metaTooltip.setAttribute("data-tooltip", "Progress: " + percent + "%\nSpeed: " + speedText);
    metaTooltip.classList.remove("is-hidden");
  }

  function clearMeta() {
    updateMeta("", 0, 0, null);
  }

  function updateTransferStats(task, uploadedBytes, totalBytes) {
    const now = Date.now();
    if (!task.transferStats) {
      task.transferStats = {
        startTime: now,
        lastTime: now,
        lastBytes: uploadedBytes,
        speed: 0,
        eta: 0
      };
      return;
    }
    const deltaTime = (now - task.transferStats.lastTime) / 1000;
    if (deltaTime <= 0) {
      return;
    }
    const deltaBytes = uploadedBytes - task.transferStats.lastBytes;
    const instantSpeed = deltaBytes / deltaTime;
    task.transferStats.speed = instantSpeed > 0 ? instantSpeed : task.transferStats.speed;
    task.transferStats.lastTime = now;
    task.transferStats.lastBytes = uploadedBytes;
    if (task.transferStats.speed > 0 && totalBytes > 0) {
      task.transferStats.eta = Math.max(0, (totalBytes - uploadedBytes) / task.transferStats.speed);
    }
  }

  function waitForThrottle(ms) {
    if (!ms || ms <= 0) {
      return Promise.resolve();
    }
    return new Promise(function(resolve) {
      setTimeout(resolve, ms);
    });
  }

  function fileSignature(file) {
    return [file.name, file.size, file.lastModified].join(":");
  }

  function loadState(signature) {
    try {
      if (!signature) {
        return null;
      }
      const uploadId = localStorage.getItem("upload:signature:" + signature);
      if (uploadId) {
        const raw = localStorage.getItem("upload:" + uploadId);
        return raw ? JSON.parse(raw) : null;
      }
      return null;
    } catch (_) {
      return null;
    }
  }

  function saveState(signature, state) {
    if (!state || !state.uploadId) {
      return;
    }
    let existing = null;
    const existingRaw = localStorage.getItem("upload:" + state.uploadId);
    if (existingRaw) {
      try {
        existing = JSON.parse(existingRaw);
      } catch (_) {
        existing = null;
      }
    }
    const storedSignature = signature || (existing ? existing.signature : "");
    const stored = {
      uploadId: state.uploadId,
      fileId: state.fileId || (existing ? existing.fileId : ""),
      mode: state.mode || (existing ? existing.mode : ""),
      chunkSize: state.chunkSize || (existing ? existing.chunkSize : 0),
      totalParts: state.totalParts || (existing ? existing.totalParts : 0),
      uploadedParts: state.uploadedParts || (existing ? existing.uploadedParts : []),
      filename: state.filename || (existing ? existing.filename : ""),
      sizeBytes: state.sizeBytes || (existing ? existing.sizeBytes : 0),
      status: state.status || (existing ? existing.status : ""),
      signature: storedSignature,
      updatedAt: Date.now()
    };
    localStorage.setItem("upload:" + state.uploadId, JSON.stringify(stored));
    if (stored.fileId) {
      localStorage.setItem("upload:file:" + stored.fileId, state.uploadId);
    }
    if (storedSignature) {
      localStorage.setItem("upload:signature:" + storedSignature, state.uploadId);
    }
    updateResumeBanner();
  }

  function clearState(signature) {
    const state = loadState(signature);
    if (!state) {
      if (signature) {
        localStorage.removeItem("upload:signature:" + signature);
      }
      return;
    }
    if (state.fileId) {
      localStorage.removeItem("upload:file:" + state.fileId);
    }
    localStorage.removeItem("upload:" + state.uploadId);
    if (signature) {
      localStorage.removeItem("upload:signature:" + signature);
    }
    updateResumeBanner();
  }

  function clearStateByUploadId(uploadId) {
    if (!uploadId) {
      return;
    }
    const raw = localStorage.getItem("upload:" + uploadId);
    if (raw) {
      try {
        const parsed = JSON.parse(raw);
        if (parsed && parsed.fileId) {
          localStorage.removeItem("upload:file:" + parsed.fileId);
        }
        if (parsed && parsed.signature) {
          localStorage.removeItem("upload:signature:" + parsed.signature);
        }
      } catch (_) {}
    }
    localStorage.removeItem("upload:" + uploadId);
    updateResumeBanner();
  }

  async function api(path, body, method) {
    let headers = { "Content-Type": "application/json" };
    if (method === "GET") {
      headers = {};
    }
    const res = await fetch(path, {
      method: method || "POST",
      headers: headers,
      body: body ? JSON.stringify(body) : undefined
    });
    if (res.status === 204) {
      return null;
    }
    const data = await res.json();
    if (!res.ok) {
      const err = new Error("Request failed");
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

  function nextUpload(uploadId, uploadedParts) {
    return api("/api/uploads/" + encodeURIComponent(uploadId) + "/next", {
      uploadedParts: uploadedParts || []
    });
  }

  function setSelectedResumeUploadId(uploadId) {
    if (!uploadId) {
      localStorage.removeItem("upload:resume-id");
      return;
    }
    localStorage.setItem("upload:resume-id", uploadId);
  }

  function getSelectedResumeUploadId() {
    return localStorage.getItem("upload:resume-id");
  }

  function clearSelectedResumeUploadId() {
    localStorage.removeItem("upload:resume-id");
  }

  function completeUpload(uploadId, parts) {
    return api("/api/uploads/" + encodeURIComponent(uploadId) + "/complete", {
      parts: parts || []
    });
  }

  function cancelUpload(uploadId) {
    return api("/api/uploads/" + encodeURIComponent(uploadId) + "/cancel", {});
  }

  function setPrimaryTask(task) {
    primaryTaskId = task ? task.id : null;
    updatePauseButtons();
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
    setStatus(primary.statusText || ("Uploading " + file.name + "..."));
  }

  function updatePauseButtons() {
    if (!pauseButton || !resumeButton || !controls) {
      return;
    }
    const primary = primaryTaskId ? activeTasks.get(primaryTaskId) : null;
    if (!primary) {
      controls.classList.add("is-hidden");
      return;
    }
    controls.classList.remove("is-hidden");
    if (primary.mode !== "multipart") {
      pauseButton.classList.add("is-hidden");
      resumeButton.classList.add("is-hidden");
      return;
    }
    if (primary.paused) {
      pauseButton.classList.add("is-hidden");
      resumeButton.classList.remove("is-hidden");
      return;
    }
    pauseButton.classList.remove("is-hidden");
    resumeButton.classList.add("is-hidden");
  }

  function waitForResume(task) {
    if (!task.paused) {
      return Promise.resolve();
    }
    return new Promise(function(resolve) {
      task.resumeWaiters.push(resolve);
    });
  }

  function resumeTask(task) {
    task.paused = false;
    task.resumeWaiters.forEach(function(resolve) { resolve(); });
    task.resumeWaiters = [];
    if (task.uploadId) {
      saveState("", { uploadId: task.uploadId, status: "uploading" });
    }
    task.statusText = "Resuming " + task.file.name + "...";
    updateTaskUI(task);
    if (task.id === primaryTaskId) {
      setStatus(task.statusText);
    }
    updatePauseButtons();
  }

  function pauseTask(task) {
    task.paused = true;
    if (task.uploadId) {
      saveState("", { uploadId: task.uploadId, status: "paused" });
    }
    task.statusText = "Upload paused.";
    updateTaskUI(task);
    if (task.id === primaryTaskId) {
      setStatus(task.statusText);
    }
    updatePauseButtons();
  }

  function updateQueueMeta() {
    if (!queueMeta) {
      return;
    }
    const pending = countQueuedFiles();
    queueMeta.textContent = pending + (pending === 1 ? " item" : " items");
    if (queueEmpty) {
      queueEmpty.classList.toggle("is-hidden", queueFiles.length > 0);
    }
  }

  function countQueuedFiles() {
    return queueFiles.filter(function(item) {
      return item.status !== "complete" && item.status !== "cancelled";
    }).length;
  }

  function buildDisplayPath(fileItem) {
    if (!fileItem.folderPath) {
      return fileItem.file.name;
    }
    return fileItem.folderPath + "/" + fileItem.file.name;
  }

  function updateTaskProgress(task, uploadedBytes, totalBytes) {
    const previous = task.uploadedBytes || 0;
    task.uploadedBytes = uploadedBytes;
    task.totalBytes = totalBytes;
    task.batch.uploadedBytes += uploadedBytes - previous;
    updateTransferStats(task, uploadedBytes, totalBytes);
    updateTaskUI(task);
    updateBatchUI(task.batch);
    if (task.id === primaryTaskId) {
      updatePrimaryUI();
    }
  }

  function updateTaskUI(task) {
    if (!task.el) {
      return;
    }
    const percent = task.totalBytes > 0 ? Math.round((task.uploadedBytes / task.totalBytes) * 100) : 0;
    let meta = "Queued • " + formatBytes(task.file.size);
    if (task.status === "uploading") {
      meta = percent + "%" + " • " + formatBytes(task.uploadedBytes) + " / " + formatBytes(task.totalBytes);
    } else if (task.status === "paused") {
      meta = "Paused • " + percent + "%";
    } else if (task.status === "error") {
      meta = "Failed";
    } else if (task.status === "complete") {
      meta = "Complete";
    } else if (task.status === "cancelled") {
      meta = "Cancelled";
    }
    task.metaEl.textContent = meta;
  }

  function updateBatchUI(batch) {
    if (!batch.el) {
      return;
    }
    const percent = batch.totalBytes > 0 ? Math.round((batch.uploadedBytes / batch.totalBytes) * 100) : 0;
    batch.progressEl.style.width = percent + "%";
    batch.statusEl.textContent = batchStatusText(batch);
  }

  function batchStatusText(batch) {
    const totalFiles = batch.files.length;
    const complete = batch.files.filter(function(item) { return item.status === "complete"; }).length;
    const uploading = batch.files.filter(function(item) { return item.status === "uploading"; }).length;
    const paused = batch.files.filter(function(item) { return item.status === "paused"; }).length;
    const failed = batch.files.filter(function(item) { return item.status === "error"; }).length;

    if (failed > 0) {
      return "Action needed";
    }
    if (complete === totalFiles) {
      return "Complete";
    }
    if (paused > 0) {
      return "Paused";
    }
    if (uploading > 0) {
      return "Uploading";
    }
    return "Queued";
  }

  function createBatch(name, files) {
    const batch = {
      id: "batch-" + (++batchCounter),
      name: name,
      files: files,
      totalBytes: files.reduce(function(sum, item) { return sum + item.file.size; }, 0),
      uploadedBytes: 0,
      expanded: true,
      el: null,
      progressEl: null,
      statusEl: null,
      filesWrap: null,
      toggleEl: null
    };

    const batchEl = document.createElement("div");
    batchEl.className = "queue-item";

    const header = document.createElement("div");
    header.className = "queue-item-header";

    const main = document.createElement("div");
    main.className = "queue-item-main";

    const title = document.createElement("span");
    title.className = "queue-item-title";
    title.textContent = name;

    const meta = document.createElement("span");
    meta.className = "queue-item-meta";
    meta.textContent = files.length + (files.length === 1 ? " file" : " files") + " • " + formatBytes(batch.totalBytes);

    main.appendChild(title);
    header.appendChild(main);
    const metaWrap = document.createElement("div");
    metaWrap.className = "queue-item-meta-wrap";
    metaWrap.appendChild(meta);

    let toggle = null;
    if (files.length > 1) {
      toggle = document.createElement("button");
      toggle.className = "queue-toggle";
      toggle.type = "button";
      toggle.setAttribute("aria-expanded", "true");
      toggle.textContent = "Hide files";
      metaWrap.appendChild(toggle);
    }
    header.appendChild(metaWrap);

    const progressWrap = document.createElement("div");
    progressWrap.className = "queue-progress";

    const bar = document.createElement("div");
    bar.className = "queue-bar";

    const barFill = document.createElement("div");
    barFill.className = "queue-bar-fill";
    bar.appendChild(barFill);

    const statusText = document.createElement("span");
    statusText.className = "queue-item-status";
    statusText.textContent = "Queued";

    progressWrap.appendChild(bar);
    progressWrap.appendChild(statusText);

    const filesWrap = document.createElement("div");
    filesWrap.className = "queue-files";

    batch.files.forEach(function(item) {
      const row = document.createElement("div");
      row.className = "queue-file";

      const nameEl = document.createElement("span");
      nameEl.className = "queue-file-name";
      nameEl.textContent = buildDisplayPath(item);

      const metaEl = document.createElement("span");
      metaEl.className = "queue-file-meta";
      metaEl.textContent = formatBytes(item.file.size);

      row.appendChild(nameEl);
      row.appendChild(metaEl);
      filesWrap.appendChild(row);

      item.el = row;
      item.metaEl = metaEl;
    });

    batchEl.appendChild(header);
    batchEl.appendChild(progressWrap);
    batchEl.appendChild(filesWrap);

    function setBatchExpanded(expanded) {
      batch.expanded = expanded;
      filesWrap.classList.toggle("is-hidden", !expanded);
      if (toggle) {
        toggle.textContent = expanded ? "Hide files" : "Show files";
        toggle.setAttribute("aria-expanded", expanded ? "true" : "false");
      }
    }

    if (toggle) {
      toggle.addEventListener("click", function() {
        setBatchExpanded(!batch.expanded);
      });
    }

    batch.el = batchEl;
    batch.progressEl = barFill;
    batch.statusEl = statusText;
    batch.filesWrap = filesWrap;
    batch.toggleEl = toggle;

    return batch;
  }

  function normalizeFolderPathFromRelative(relativePath) {
    if (!relativePath) {
      return "";
    }
    const parts = relativePath.split("/").filter(Boolean);
    if (parts.length <= 1) {
      return "";
    }
    return parts.slice(0, -1).join("/");
  }

  function deriveFolderPath(file) {
    return normalizeFolderPathFromRelative(file.webkitRelativePath || "");
  }

  function normalizeQueueItems(items) {
    return items.map(function(item) {
      const file = item && item.file ? item.file : item;
      const folderPath = item && item.file ? (item.folderPath || "") : deriveFolderPath(file);
      return { file: file, folderPath: folderPath || "" };
    });
  }

  function resolveBatchName(items) {
    if (!items.length) {
      return "Uploads";
    }
    if (items.length === 1 && !items[0].folderPath) {
      return items[0].file.name;
    }
    const roots = items.map(function(item) {
      if (item.folderPath) {
        return item.folderPath.split("/")[0] || "";
      }
      if (item.file && item.file.webkitRelativePath) {
        const parts = item.file.webkitRelativePath.split("/").filter(Boolean);
        return parts.length ? parts[0] : "";
      }
      return "";
    });
    const root = roots[0] || "";
    const sameRoot = root && roots.every(function(item) { return item === root; });
    if (sameRoot) {
      return root;
    }
    return "Selected files";
  }

  function addFilesToQueue(items) {
    if (!items || !items.length) {
      return;
    }
    const newFiles = normalizeQueueItems(Array.from(items));
    const remaining = MAX_QUEUE_ITEMS - countQueuedFiles();
    if (newFiles.length > remaining) {
      setStatus("Queue limit reached. You can queue up to " + MAX_QUEUE_ITEMS + " files.");
      toastUploadInfo("Queue limit reached. Split the upload into smaller batches.", "Queue limit");
      return;
    }

    const batchName = resolveBatchName(newFiles);
    const fileItems = newFiles.map(function(item) {
      return {
        id: "file-" + (++fileCounter),
        file: item.file,
        folderPath: item.folderPath,
        status: "queued",
        uploadedBytes: 0,
        totalBytes: item.file.size,
        resumeWaiters: [],
        paused: false,
        transferStats: null,
        statusText: "Queued",
        batch: null,
        el: null,
        metaEl: null
      };
    });

    const batch = createBatch(batchName, fileItems);
    fileItems.forEach(function(item) {
      item.batch = batch;
      queueFiles.push(item);
    });
    batches.push(batch);
    queueList.appendChild(batch.el);

    updateQueueMeta();
    updateInputState();
    runQueue();
  }

  function readEntryFile(entry) {
    return new Promise(function(resolve, reject) {
      entry.file(function(file) { resolve(file); }, reject);
    });
  }

  function readDirectoryEntries(reader) {
    return new Promise(function(resolve, reject) {
      reader.readEntries(function(entries) { resolve(entries); }, reject);
    });
  }

  async function collectEntries(entry, collected) {
    if (entry.isFile) {
      const file = await readEntryFile(entry);
      const fullPath = entry.fullPath ? entry.fullPath.replace(/^\/+/, "") : file.webkitRelativePath || file.name;
      const folderPath = normalizeFolderPathFromRelative(fullPath);
      collected.push({ file: file, folderPath: folderPath });
      return;
    }
    if (entry.isDirectory) {
      const reader = entry.createReader();
      let entries = await readDirectoryEntries(reader);
      while (entries.length) {
        for (let i = 0; i < entries.length; i++) {
          await collectEntries(entries[i], collected);
        }
        entries = await readDirectoryEntries(reader);
      }
    }
  }

  async function collectDroppedFiles(dataTransfer) {
    if (!dataTransfer) {
      return [];
    }
    const items = dataTransfer.items;
    if (!items || !items.length || !items[0].webkitGetAsEntry) {
      return Array.from(dataTransfer.files || []);
    }
    const collected = [];
    for (let i = 0; i < items.length; i++) {
      const item = items[i];
      if (!item || item.kind !== "file") {
        continue;
      }
      const entry = item.webkitGetAsEntry();
      if (!entry) {
        continue;
      }
      await collectEntries(entry, collected);
    }
    return collected;
  }

  async function uploadNextPartWithRetry(task, uploadId, chunkSize, file, uploadedMap, maxRetries) {
    let attempt = 0;

    function tryUpload() {
      if (task.canceled) {
        throw new Error("Cancelled");
      }
      attempt++;
      return nextUpload(uploadId, Array.from(uploadedMap.keys()))
        .then(function(res) {
          if (!res || !res.nextPart || !res.url) {
            const err = new Error("No next part");
            err.noNext = true;
            if (res && res.uploadedParts) {
              err.uploadedParts = res.uploadedParts;
            }
            throw err;
          }
          if (res.uploadedParts && res.uploadedParts.length) {
            res.uploadedParts.forEach(function(part) {
              if (part && part.partNumber && part.etag) {
                uploadedMap.set(part.partNumber, part.etag);
              }
            });
          }
          const partNumber = res.nextPart;
          const startByte = (partNumber - 1) * chunkSize;
          const endByte = Math.min(startByte + chunkSize, file.size);
          const chunk = file.slice(startByte, endByte);
          return fetch(res.url, { method: "PUT", body: chunk })
            .then(function(uploadRes) {
              if (!uploadRes.ok) {
                throw new Error("Upload failed: " + uploadRes.status);
              }
              const etag = uploadRes.headers.get("ETag");
              if (!etag) {
                throw new Error("Missing ETag");
              }
              return {
                partNumber: partNumber,
                etag: etag,
                size: chunk.size,
                throttleMs: res.throttleMs
              };
            });
        })
        .catch(function(err) {
          if (err && err.status === 409) {
            err.cancelled = true;
            throw err;
          }
          if (err && err.noNext) {
            throw err;
          }
          if (attempt >= maxRetries) {
            throw err;
          }
          return new Promise(function(resolve) {
            setTimeout(resolve, 500 * attempt * attempt);
          }).then(tryUpload);
        });
    }

    return tryUpload();
  }

  function buildUploadedPartsFromMap(file, chunkSize, uploadedMap) {
    let parts = [];
    let bytes = 0;
    const numbers = Array.from(uploadedMap.keys()).sort(function(a, b) { return a - b; });
    for (let i = 0; i < numbers.length; i++) {
      const partNumber = numbers[i];
      const startByte = (partNumber - 1) * chunkSize;
      const endByte = Math.min(startByte + chunkSize, file.size);
      const size = Math.max(0, endByte - startByte);
      parts.push({ partNumber: partNumber, etag: uploadedMap.get(partNumber), size: size });
      bytes += size;
    }
    return { parts: parts, bytes: bytes };
  }

  async function recoverMissingParts(task, uploadId, chunkSize, file, uploadedMap, signature, fileId, totalParts) {
    while (uploadedMap.size < totalParts) {
      if (task.canceled) {
        return;
      }
      try {
        const result = await uploadNextPartWithRetry(task, uploadId, chunkSize, file, uploadedMap, 5);
        uploadedMap.set(result.partNumber, result.etag);
        const rebuilt = buildUploadedPartsFromMap(file, chunkSize, uploadedMap);
        updateTaskProgress(task, rebuilt.bytes, file.size);
        saveState(signature, {
          uploadId: uploadId,
          fileId: fileId,
          mode: "multipart",
          chunkSize: chunkSize,
          totalParts: totalParts,
          uploadedParts: rebuilt.parts,
          filename: file.name,
          sizeBytes: file.size,
          status: "uploading"
        });
        const throttleMs = typeof result.throttleMs === "number"
          ? result.throttleMs
          : (file.size >= MAX_FILE_SIZE ? LARGE_FILE_THROTTLE_MS : 0);
        await waitForThrottle(throttleMs);
      } catch (err) {
        if (err && err.noNext) {
          break;
        }
        throw err;
      }
    }
  }

  async function uploadMultipart(task, initialResponse) {
    const file = task.file;
    const signature = fileSignature(file);
    let existing = loadState(signature);
    let state = null;

    async function initSession() {
      if (existing && existing.mode === "multipart" && existing.uploadId && existing.chunkSize && existing.totalParts) {
        return existing;
      }
      const resumeUploadId = getSelectedResumeUploadId();
      if (resumeUploadId) {
        try {
          const next = await nextUpload(resumeUploadId, []);
          if (!next || next.mode !== "multipart") {
            throw new Error("Resume file mismatch");
          }
          const resumedState = {
            uploadId: next.uploadId || resumeUploadId,
            fileId: next.fileId,
            chunkSize: next.chunkSize,
            totalParts: next.totalParts,
            uploadedParts: next.uploadedParts || [],
            mode: "multipart",
            filename: file.name,
            sizeBytes: file.size,
            status: "uploading"
          };
          saveState(signature, resumedState);
          clearSelectedResumeUploadId();
          return resumedState;
        } catch (_) {
          clearSelectedResumeUploadId();
        }
      }
      if (initialResponse) {
        if (initialResponse.mode !== "multipart") {
          throw new Error("Expected multipart upload");
        }
        const nextState = {
          uploadId: initialResponse.uploadId,
          fileId: initialResponse.fileId,
          chunkSize: initialResponse.chunkSize,
          totalParts: initialResponse.totalParts,
          uploadedParts: [],
          mode: "multipart",
          filename: file.name,
          sizeBytes: file.size,
          status: "uploading"
        };
        saveState(signature, nextState);
        return nextState;
      }
      const resp = await startUpload(task);
      if (resp.mode !== "multipart") {
        throw new Error("Expected multipart upload");
      }
      const nextState = {
        uploadId: resp.uploadId,
        fileId: resp.fileId,
        chunkSize: resp.chunkSize,
        totalParts: resp.totalParts,
        uploadedParts: [],
        mode: "multipart",
        filename: file.name,
        sizeBytes: file.size,
        status: "uploading"
      };
      saveState(signature, nextState);
      return nextState;
    }

    try {
      state = await initSession();
      task.uploadId = state.uploadId;
      task.fileId = state.fileId;
      task.mode = "multipart";
      task.statusText = "Uploading " + file.name + "...";

      const uploadId = state.uploadId;
      const chunkSize = state.chunkSize;
      const totalParts = state.totalParts;
      const uploadedMap = new Map((state.uploadedParts || []).map(function(p) { return [p.partNumber, p.etag]; }));
      const rebuilt = buildUploadedPartsFromMap(file, chunkSize, uploadedMap);
      updateTaskProgress(task, rebuilt.bytes, file.size);

      saveState(signature, {
        uploadId: uploadId,
        fileId: state.fileId,
        mode: "multipart",
        chunkSize: chunkSize,
        totalParts: totalParts,
        uploadedParts: rebuilt.parts,
        filename: file.name,
        sizeBytes: file.size,
        status: "uploading"
      });

      while (uploadedMap.size < totalParts) {
        if (task.canceled) {
          return;
        }
        await waitForResume(task);
        if (task.canceled) {
          return;
        }

        try {
          const result = await uploadNextPartWithRetry(task, uploadId, chunkSize, file, uploadedMap, 5);
          uploadedMap.set(result.partNumber, result.etag);
          const updated = buildUploadedPartsFromMap(file, chunkSize, uploadedMap);
          updateTaskProgress(task, updated.bytes, file.size);
          saveState(signature, {
            uploadId: uploadId,
            fileId: state.fileId,
            mode: "multipart",
            chunkSize: chunkSize,
            totalParts: totalParts,
            uploadedParts: updated.parts,
            filename: file.name,
            sizeBytes: file.size,
            status: "uploading"
          });
          const throttleMs = typeof result.throttleMs === "number"
            ? result.throttleMs
            : (file.size >= MAX_FILE_SIZE ? LARGE_FILE_THROTTLE_MS : 0);
          await waitForThrottle(throttleMs);
        } catch (err) {
          if (err && err.cancelled) {
            clearState(signature);
            task.status = "cancelled";
            task.statusText = "Upload cancelled.";
            toastUploadInfo("Upload cancelled.", "Cancelled");
            return;
          }
          if (err && err.noNext) {
            break;
          }
          if (err && err.status === 404 && existing) {
            clearState(signature);
          }
          const message = err && err.data && err.data.errors && err.data.errors.size ? err.data.errors.size : null;
          task.status = "error";
          task.statusText = "Upload failed.";
          setStatus("Upload failed. " + (message || (err && err.message ? err.message : "Try again.")));
          throw err;
        }
      }

      if (task.canceled) {
        return;
      }

      let parts = Array.from(uploadedMap.entries())
        .map(function(entry) { return { partNumber: entry[0], etag: entry[1] }; })
        .sort(function(a, b) { return a.partNumber - b.partNumber; });

      try {
        await completeUpload(uploadId, parts);
      } catch (err) {
        const missingParts = err && err.status === 409 && err.data && err.data.missingParts ? err.data.missingParts : null;
        if (!missingParts || !missingParts.length) {
          throw err;
        }
        task.statusText = "Recovering missing parts...";
        setStatus(task.statusText);
        await recoverMissingParts(task, uploadId, chunkSize, file, uploadedMap, signature, state.fileId, totalParts);
        const refreshed = buildUploadedPartsFromMap(file, chunkSize, uploadedMap);
        updateTaskProgress(task, refreshed.bytes, file.size);
        parts = Array.from(uploadedMap.entries())
          .map(function(entry) { return { partNumber: entry[0], etag: entry[1] }; })
          .sort(function(a, b) { return a.partNumber - b.partNumber; });
        await completeUpload(uploadId, parts);
      }
      clearState(signature);
      task.status = "complete";
      task.statusText = "Upload complete: " + file.name;
      updateTaskProgress(task, file.size, file.size);
      if (task.id === primaryTaskId && metaDetail && metaTooltip) {
        metaDetail.textContent = "Complete";
        metaTooltip.setAttribute("data-tooltip", "100% • Done");
      }
      toastUploadSuccess(file.name);
    } catch (err) {
      const uploadId = state && state.uploadId ? state.uploadId : (existing && existing.uploadId ? existing.uploadId : null);
      await cleanupFailure(task, uploadId, signature);
      const detail = err && err.data && err.data.errors && err.data.errors.size ? err.data.errors.size : (err && err.message ? err.message : "");
      toastUploadError(detail);
      throw err;
    }
  }

  function uploadSingleWithProgress(task, url, file) {
    return new Promise(function(resolve, reject) {
      const xhr = new XMLHttpRequest();
      task.xhr = xhr;
      xhr.upload.onprogress = function(evt) {
        if (!evt.lengthComputable) {
          return;
        }
        updateTaskProgress(task, evt.loaded, evt.total);
      };
      xhr.onload = function() {
        if (xhr.status >= 200 && xhr.status < 300) {
          resolve();
        } else {
          reject(new Error("Upload failed: " + xhr.status));
        }
      };
      xhr.onerror = function() {
        reject(new Error("Upload failed"));
      };
      xhr.onabort = function() {
        const err = new Error("Upload cancelled");
        err.cancelled = true;
        reject(err);
      };
      xhr.open("PUT", url);
      xhr.send(file);
    });
  }

  async function uploadSingle(task, initialResponse) {
    const file = task.file;
    const signature = fileSignature(file);
    let resp = initialResponse;
    try {
      if (!resp) {
        resp = await startUpload(task);
      }
      if (resp.mode !== "single") {
        throw new Error("Expected single upload");
      }
      task.uploadId = resp.uploadId;
      task.fileId = resp.fileId;
      task.mode = "single";
      task.statusText = "Uploading " + file.name + "...";
      saveState(signature, {
        uploadId: resp.uploadId,
        fileId: resp.fileId,
        mode: "single",
        uploadedParts: [],
        filename: file.name,
        sizeBytes: file.size,
        status: "uploading"
      });

      let uploadUrl = resp.uploadUrl;
      if (!uploadUrl) {
        const next = await nextUpload(resp.uploadId, []);
        uploadUrl = next ? next.url : "";
      }
      if (!uploadUrl) {
        throw new Error("Missing upload URL");
      }

      await uploadSingleWithProgress(task, uploadUrl, file);
      await completeUpload(resp.uploadId, []);
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
      await cleanupFailure(task, resp ? resp.uploadId : null, signature);
      throw err;
    }
  }

  async function cleanupFailure(task, uploadId, signature) {
    clearSelectedResumeUploadId();
    if (signature) {
      clearState(signature);
    }
    if (uploadId) {
      clearStateByUploadId(uploadId);
    }
    if (task.xhr) {
      task.xhr.abort();
    }
    task.status = "error";
    task.statusText = "Upload failed.";
    updateTaskUI(task);
    if (!uploadId) {
      return;
    }
    try {
      await cancelUpload(uploadId);
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
    if (task.file.size >= MULTIPART_THRESHOLD) {
      await uploadMultipart(task);
      return;
    }
    const resp = await startUpload(task);
    if (resp.mode === "multipart") {
      await uploadMultipart(task, resp);
      return;
    }
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
          task.status = "error";
          task.statusText = "Upload failed.";
          const detail = err && err.data && err.data.errors && err.data.errors.size ? err.data.errors.size : (err && err.message ? err.message : "");
          toastUploadError(detail);
        }
      })
      .finally(function() {
        activeTasks.delete(task.id);
        if (primaryTaskId === task.id) {
          setPrimaryTask(null);
          const remaining = Array.from(activeTasks.values())[0];
          if (remaining) {
            setPrimaryTask(remaining);
          }
        }
        updateTaskUI(task);
        updateBatchUI(task.batch);
        updateQueueMeta();
        runQueue();
      });
  }

  function runQueue() {
    while (activeTasks.size < MAX_CONCURRENCY) {
      const next = queueFiles.find(function(item) { return item.status === "queued"; });
      if (!next) {
        break;
      }
      startTask(next);
    }
    updateInputState();
  }

  function resetSelection() {
    if (input) {
      input.value = "";
    }
    if (folderInput) {
      folderInput.value = "";
    }
    if (chip) {
      chip.classList.add("is-hidden");
      if (chipName) {
        chipName.textContent = "";
      }
      if (chipSize) {
        chipSize.textContent = "";
      }
    }
  }

  function updateResumeBanner() {
    const banner = document.getElementById("upload-resume-banner");
    if (!banner) {
      return;
    }
    let session = null;
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (!key || key.indexOf("upload:") !== 0) {
        continue;
      }
      if (key.indexOf("upload:signature:") === 0 || key.indexOf("upload:file:") === 0 || key === "upload:resume-id") {
        continue;
      }
      const sessionCandidate = getUploadSession(key.replace("upload:", ""));
      if (!sessionCandidate || sessionCandidate.mode !== "multipart" || !sessionCandidate.uploadId || !sessionCandidate.totalParts) {
        continue;
      }
      if (sessionCandidate.status && sessionCandidate.status !== "paused" && sessionCandidate.status !== "uploading") {
        continue;
      }
      const uploadedParts = sessionCandidate.uploadedParts || [];
      if (uploadedParts.length >= sessionCandidate.totalParts) {
        continue;
      }
      session = sessionCandidate;
      break;
    }

    if (!session) {
      banner.classList.add("is-hidden");
      return;
    }

    const uploadedParts = session.uploadedParts || [];
    let percent = 0;
    if (session.totalParts > 0) {
      percent = Math.max(0, Math.min(100, Math.round((uploadedParts.length / session.totalParts) * 100)));
    }

    const resumeMeta = document.getElementById("resume-banner-meta");
    if (resumeMeta) {
      const filename = session.filename || "Pending upload";
      resumeMeta.textContent = filename + " • " + percent + "%";
    }
    banner.classList.remove("is-hidden");
    banner.setAttribute("data-upload-id", session.uploadId);
  }

  function getUploadSession(uploadId) {
    if (!uploadId) {
      return null;
    }
    const raw = localStorage.getItem("upload:" + uploadId);
    if (!raw) {
      return null;
    }
    try {
      return JSON.parse(raw);
    } catch (_) {
      return null;
    }
  }

  if (browseFilesButton) {
    browseFilesButton.addEventListener("click", function() {
      if (input && !input.disabled) {
        input.click();
      }
    });
  }

  if (browseFoldersButton) {
    browseFoldersButton.addEventListener("click", function() {
      if (folderInput && !folderInput.disabled) {
        enableFolderPicker();
        folderInput.click();
      }
    });
  }

  if (dropzone) {
    dropzone.addEventListener("click", function() {
      if (input && !input.disabled) {
        input.click();
      }
    });
    dropzone.addEventListener("keydown", function(event) {
      if ((event.key === "Enter" || event.key === " ") && input && !input.disabled) {
        event.preventDefault();
        input.click();
      }
    });
    dropzone.addEventListener("dragover", function(event) {
      event.preventDefault();
      dropzone.classList.add("is-active");
    });
    dropzone.addEventListener("dragleave", function() {
      dropzone.classList.remove("is-active");
    });
    dropzone.addEventListener("drop", function(event) {
      event.preventDefault();
      dropzone.classList.remove("is-active");
      const transfer = event.dataTransfer || null;
      collectDroppedFiles(transfer)
        .then(function(files) {
          if (files && files.length) {
            addFilesToQueue(files);
          }
          resetSelection();
        })
        .catch(function() {
          resetSelection();
        });
    });
  }

  if (chipClear) {
    chipClear.addEventListener("click", function() {
      resetSelection();
    });
  }

  input.addEventListener("change", function() {
    if (!input.files || !input.files.length) {
      resetSelection();
      return;
    }
    const resumeUploadId = getSelectedResumeUploadId();
    if (resumeUploadId) {
      if (input.files.length !== 1) {
        setStatus("Select the paused file to resume.");
        resetSelection();
        return;
      }
      const file = input.files[0];
      const session = getUploadSession(resumeUploadId);
      if (session && session.filename && session.sizeBytes) {
        if (session.filename !== file.name || session.sizeBytes !== file.size) {
          setStatus("Selected file does not match the paused upload. Choose the same file to resume.");
          clearSelectedResumeUploadId();
          resetSelection();
          return;
        }
      }
      clearSelectedResumeUploadId();
      addFilesToQueue([file]);
      resetSelection();
      return;
    }
    addFilesToQueue(input.files);
    resetSelection();
  });

  if (folderInput) {
    folderInput.addEventListener("change", function() {
      if (!folderInput.files || !folderInput.files.length) {
        resetSelection();
        return;
      }
      addFilesToQueue(folderInput.files);
      resetSelection();
    });
  }

  if (pauseButton) {
    pauseButton.addEventListener("click", function() {
      const primary = primaryTaskId ? activeTasks.get(primaryTaskId) : null;
      if (!primary || primary.mode !== "multipart") {
        setStatus("Pause is only available for multipart uploads.");
        return;
      }
      pauseTask(primary);
      setStatus("Upload paused.");
    });
  }

  if (resumeButton) {
    resumeButton.addEventListener("click", function() {
      const primary = primaryTaskId ? activeTasks.get(primaryTaskId) : null;
      if (!primary || primary.mode !== "multipart") {
        setStatus("No multipart upload to resume.");
        return;
      }
      resumeTask(primary);
      setStatus("Resuming upload...");
    });
  }

  if (abortButton) {
    abortButton.addEventListener("click", function() {
      const primary = primaryTaskId ? activeTasks.get(primaryTaskId) : null;
      if (!primary) {
        setStatus("No active upload to abort.");
        return;
      }
      abortButton.disabled = true;
      primary.canceled = true;
      if (primary.xhr) {
        primary.xhr.abort();
      }
      if (primary.uploadId) {
        cancelUpload(primary.uploadId)
          .then(function() {
            primary.status = "cancelled";
            primary.statusText = "Upload cancelled.";
            updateTaskUI(primary);
            updateBatchUI(primary.batch);
            toastUploadInfo("Upload cancelled.", "Cancelled");
          })
          .catch(function() {})
          .finally(function() {
            abortButton.disabled = false;
          });
      }
    });
  }

  const resumeBanner = document.getElementById("upload-resume-banner");
  const resumeBannerResume = document.getElementById("resume-banner-resume");
  const resumeBannerCancel = document.getElementById("resume-banner-cancel");

  if (resumeBannerResume) {
    resumeBannerResume.addEventListener("click", function() {
      if (!resumeBanner) {
        return;
      }
      const uploadId = resumeBanner.getAttribute("data-upload-id");
      if (!uploadId) {
        return;
      }
      setSelectedResumeUploadId(uploadId);
      if (input && !input.disabled) {
        input.click();
      }
    });
  }

  if (resumeBannerCancel) {
    resumeBannerCancel.addEventListener("click", function() {
      if (!resumeBanner) {
        return;
      }
      const uploadId = resumeBanner.getAttribute("data-upload-id");
      if (!uploadId) {
        return;
      }
      cancelUpload(uploadId)
        .then(function() {
          clearStateByUploadId(uploadId);
          setStatus("Upload cancelled.");
          toastUploadInfo("Upload cancelled.", "Cancelled");
          updateResumeBanner();
        })
        .catch(function() {});
    });
  }

  if (confirmStart) {
    confirmStart.addEventListener("click", function() {
      if (confirmBackdrop) {
        confirmBackdrop.classList.add("is-hidden");
      }
    });
  }

  if (confirmCancel) {
    confirmCancel.addEventListener("click", function() {
      if (confirmBackdrop) {
        confirmBackdrop.classList.add("is-hidden");
      }
    });
  }

  updateQueueMeta();
  updateInputState();
  updateResumeBanner();
})();
