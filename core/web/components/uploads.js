(function() {
  const input = document.getElementById("upload-file");
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
  const progress = document.getElementById("upload-progress");
  const status = document.getElementById("upload-status");

  if (!input || !progress || !status) {
    return;
  }

  const progressBar = progress.querySelector(".progress-bar");
  const progressPercent = progress.querySelector(".progress-percent");
  let active = null;
  let paused = false;
  let resumeWaiters = [];
  let selectedFile = null;
  let transferStats = null;
  const MAX_FILE_SIZE = 1024 * 1024 * 1024;
  const MULTIPART_THRESHOLD = 200 * 1024 * 1024;

  function setStatus(message) {
    status.textContent = message;
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

  function setProgress(percent) {
    const clamped = Math.max(0, Math.min(100, Math.round(percent)));
    if (progressBar) {
      progressBar.style.width = clamped + "%";
      progressBar.setAttribute("aria-valuenow", String(clamped));
    }
    if (progressPercent) {
      progressPercent.textContent = clamped + "%";
    }
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

  function updateMeta(fileName, uploadedBytes, totalBytes) {
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
    updateMeta("", 0, 0);
  }

  function updateTransferStats(uploadedBytes, totalBytes) {
    const now = Date.now();
    if (!transferStats) {
      transferStats = {
        startTime: now,
        lastTime: now,
        lastBytes: uploadedBytes,
        speed: 0,
        eta: 0
      };
      return;
    }
    const deltaTime = (now - transferStats.lastTime) / 1000;
    if (deltaTime <= 0) {
      return;
    }
    const deltaBytes = uploadedBytes - transferStats.lastBytes;
    const instantSpeed = deltaBytes / deltaTime;
    transferStats.speed = instantSpeed > 0 ? instantSpeed : transferStats.speed;
    transferStats.lastTime = now;
    transferStats.lastBytes = uploadedBytes;
    if (transferStats.speed > 0 && totalBytes > 0) {
      transferStats.eta = Math.max(0, (totalBytes - uploadedBytes) / transferStats.speed);
    }
  }

  function setSelectedFile(file, shouldConfirm) {
    selectedFile = file || null;
    if (!chip || !chipName || !chipSize) {
      return;
    }
    if (!selectedFile) {
      chip.classList.add("is-hidden");
      chipName.textContent = "";
      chipSize.textContent = "";
    if (startButton) {
      startButton.disabled = true;
      startButton.textContent = "Select a file to upload";
    }
    return;
  }
    chip.classList.remove("is-hidden");
    chipName.textContent = selectedFile.name;
    chipSize.textContent = formatBytes(selectedFile.size);
    if (startButton) {
      startButton.disabled = false;
      startButton.textContent = "Upload " + selectedFile.name;
    }
    if (shouldConfirm) {
      openConfirmDialog(selectedFile);
    }
  }

  function openConfirmDialog(file) {
    if (!confirmBackdrop || !confirmMeta || !confirmStart) {
      return;
    }
    confirmMeta.textContent = file.name + " • " + formatBytes(file.size);
    if (window.Dialog && window.Dialog.open) {
      window.Dialog.open("upload-confirm-backdrop");
    } else {
      confirmBackdrop.classList.remove("is-hidden");
    }
  }

  function closeConfirmDialog() {
    if (!confirmBackdrop) {
      return;
    }
    if (window.Dialog && window.Dialog.close) {
      window.Dialog.close("upload-confirm-backdrop");
    } else {
      confirmBackdrop.classList.add("is-hidden");
    }
  }

  function updatePauseButtons() {
    if (!pauseButton || !resumeButton || !controls) {
      return;
    }
    if (!active) {
      controls.classList.add("is-hidden");
      return;
    }
    controls.classList.remove("is-hidden");
    if (active.mode !== "multipart") {
      pauseButton.classList.add("is-hidden");
      resumeButton.classList.add("is-hidden");
      return;
    }
    if (paused) {
      pauseButton.classList.add("is-hidden");
      resumeButton.classList.remove("is-hidden");
      return;
    }
    pauseButton.classList.remove("is-hidden");
    resumeButton.classList.add("is-hidden");
  }

  function setPaused(next) {
    paused = next;
    if (!paused) {
      resumeWaiters.forEach(function(resolve) { resolve(); });
      resumeWaiters = [];
    }
    if (active && active.uploadId) {
      saveState("", { uploadId: active.uploadId, status: paused ? "paused" : "uploading" });
    }
    updatePauseButtons();
    updateResumeBanner();
  }

  function waitForResume() {
    if (!paused) {
      return Promise.resolve();
    }
    return new Promise(function(resolve) {
      resumeWaiters.push(resolve);
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

  function startUpload(file) {
    return api("/api/uploads/start", {
      filename: file.name,
      size: file.size,
      contentType: file.type || "application/octet-stream"
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

  async function cleanupFailure(uploadId, signature) {
    clearSelectedResumeUploadId();
    if (signature) {
      clearState(signature);
    }
    if (uploadId) {
      clearStateByUploadId(uploadId);
    }
    if (active && active.cancel) {
      active.cancel();
    }
    active = null;
    setPaused(false);
    setProgress(0);
    clearMeta();
    updatePauseButtons();
    if (!uploadId) {
      return;
    }
    try {
      await cancelUpload(uploadId);
    } catch (_) {}
  }

  function uploadNextPartWithRetry(uploadId, chunkSize, file, uploadedMap, maxRetries) {
    let attempt = 0;

    function tryUpload() {
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
              return { partNumber: partNumber, etag: etag, size: chunk.size };
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

  async function recoverMissingParts(uploadId, chunkSize, file, uploadedMap, signature, fileId, totalParts) {
    while (uploadedMap.size < totalParts) {
      try {
        const result = await uploadNextPartWithRetry(uploadId, chunkSize, file, uploadedMap, 5);
        uploadedMap.set(result.partNumber, result.etag);
        const rebuilt = buildUploadedPartsFromMap(file, chunkSize, uploadedMap);
        updateTransferStats(rebuilt.bytes, file.size);
        updateMeta(file.name, rebuilt.bytes, file.size);
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
        setProgress((rebuilt.bytes / file.size) * 100);
      } catch (err) {
        if (err && err.noNext) {
          break;
        }
        throw err;
      }
    }
  }

  async function uploadMultipart(file) {
    const signature = fileSignature(file);
    let existing = loadState(signature);
    let canceled = false;
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
      const resp = await startUpload(file);
      if (resp.mode !== "multipart") {
        throw new Error("Expected multipart upload");
      }
      const state = {
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
      saveState(signature, state);
      return state;
    }

    try {
      state = await initSession();
      const uploadId = state.uploadId;
      const chunkSize = state.chunkSize;
      const totalParts = state.totalParts;
      const uploadedMap = new Map((state.uploadedParts || []).map(function(p) { return [p.partNumber, p.etag]; }));
      const rebuilt = buildUploadedPartsFromMap(file, chunkSize, uploadedMap);
      updateTransferStats(rebuilt.bytes, file.size);
      updateMeta(file.name, rebuilt.bytes, file.size);
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

      active = {
        mode: "multipart",
        uploadId: uploadId,
        signature: signature,
        fileId: state.fileId,
        cancel: function() { canceled = true; }
      };
      setProgress((rebuilt.bytes / file.size) * 100);
      setStatus("Uploading " + file.name + "...");
      updatePauseButtons();

      while (uploadedMap.size < totalParts) {
        if (canceled) {
          return;
        }
        await waitForResume();
        if (canceled) {
          return;
        }

        try {
          const result = await uploadNextPartWithRetry(uploadId, chunkSize, file, uploadedMap, 5);
          uploadedMap.set(result.partNumber, result.etag);
          const updated = buildUploadedPartsFromMap(file, chunkSize, uploadedMap);
          updateTransferStats(updated.bytes, file.size);
          updateMeta(file.name, updated.bytes, file.size);
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
          setProgress((updated.bytes / file.size) * 100);
        } catch (err) {
          if (err && err.cancelled) {
            clearState(signature);
            setStatus("Upload cancelled.");
            toastUploadInfo("Upload cancelled.", "Cancelled");
            active = null;
            setPaused(false);
            return;
          }
          if (err && err.noNext) {
            break;
          }
          if (err && err.status === 404 && existing) {
            clearState(signature);
          }
          const message = err && err.data && err.data.errors && err.data.errors.size ? err.data.errors.size : null;
          setStatus("Upload failed. " + (message || (err && err.message ? err.message : "Try again.")));
          throw err;
        }
      }

      if (canceled) {
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
        setStatus("Recovering missing parts...");
        await recoverMissingParts(uploadId, chunkSize, file, uploadedMap, signature, state.fileId, totalParts);
        const refreshed = buildUploadedPartsFromMap(file, chunkSize, uploadedMap);
        updateTransferStats(refreshed.bytes, file.size);
        updateMeta(file.name, refreshed.bytes, file.size);
        setProgress((refreshed.bytes / file.size) * 100);
        parts = Array.from(uploadedMap.entries())
          .map(function(entry) { return { partNumber: entry[0], etag: entry[1] }; })
          .sort(function(a, b) { return a.partNumber - b.partNumber; });
        await completeUpload(uploadId, parts);
      }
      clearState(signature);
      setProgress(100);
      setStatus("Upload complete: " + file.name);
      if (metaDetail && metaTooltip) {
        metaDetail.textContent = "Complete";
        metaTooltip.setAttribute("data-tooltip", "100% • Done");
      }
      toastUploadSuccess(file.name);
      active = null;
      setPaused(false);
    } catch (err) {
      const uploadId = state && state.uploadId ? state.uploadId : (existing && existing.uploadId ? existing.uploadId : null);
      await cleanupFailure(uploadId, signature);
      const detail = err && err.data && err.data.errors && err.data.errors.size ? err.data.errors.size : (err && err.message ? err.message : "");
      toastUploadError(detail);
      throw err;
    }
  }

  async function uploadSingle(file) {
    const controller = new AbortController();
    let uploadId = null;
    const signature = fileSignature(file);

    try {
      const resp = await startUpload(file);
      if (resp.mode !== "single") {
        throw new Error("Expected single upload");
      }
      uploadId = resp.uploadId;
      saveState(signature, {
        uploadId: resp.uploadId,
        fileId: resp.fileId,
        mode: "single",
        uploadedParts: [],
        filename: file.name,
        sizeBytes: file.size,
        status: "uploading"
      });
      active = { mode: "single", uploadId: resp.uploadId, fileId: resp.fileId, controller: controller, signature: signature };
      setProgress(0);
      setStatus("Uploading " + file.name + "...");
      updatePauseButtons();

      try {
        let uploadUrl = resp.uploadUrl;
        if (!uploadUrl) {
          try {
            const next = await nextUpload(resp.uploadId, []);
            uploadUrl = next ? next.url : "";
          } catch (err) {
            if (err && err.status === 409) {
              clearState(signature);
              setStatus("Upload cancelled.");
              active = null;
              setPaused(false);
              return;
            }
            throw err;
          }
        }
        if (!uploadUrl) {
          throw new Error("Upload URL missing");
        }
        await uploadSingleWithProgress(uploadUrl, file, controller);
        await completeUpload(resp.uploadId, []);
        clearState(signature);
        setProgress(100);
        setStatus("Upload complete: " + file.name);
        if (metaDetail && metaTooltip) {
          metaDetail.textContent = "Complete";
          metaTooltip.setAttribute("data-tooltip", "100% • Done");
        }
        toastUploadSuccess(file.name);
        active = null;
        setPaused(false);
      } catch (err) {
        if (err && err.name === "AbortError") {
          return;
        }
        const message = err && err.data && err.data.errors && err.data.errors.size ? err.data.errors.size : null;
        await cleanupFailure(resp.uploadId, signature);
        uploadId = null;
        setStatus("Upload failed. " + (message || (err && err.message ? err.message : "Try again.")));
        throw err;
      }
    } catch (err) {
      const message_2 = err && err.data && err.data.errors && err.data.errors.size ? err.data.errors.size : null;
      await cleanupFailure(uploadId, signature);
      const detail = message_2 || (err && err.message ? err.message : "");
      toastUploadError(detail);
      setStatus("Upload failed. " + (message_2 || (err && err.message ? err.message : "Try again.")));
      throw err;
    }
  }

  function uploadSingleWithProgress(url, file, controller) {
    return new Promise(function(resolve, reject) {
      const xhr = new XMLHttpRequest();
      xhr.open("PUT", url, true);
      xhr.upload.onprogress = function(evt) {
        if (!evt.lengthComputable) {
          return;
        }
        updateTransferStats(evt.loaded, file.size);
        updateMeta(file.name, evt.loaded, file.size);
        setProgress((evt.loaded / file.size) * 100);
      };
      xhr.onload = function() {
        if (xhr.status >= 200 && xhr.status < 300) {
          resolve();
          return;
        }
        reject(new Error("Upload failed: " + xhr.status));
      };
      xhr.onerror = function() {
        reject(new Error("Upload failed."));
      };
      xhr.onabort = function() {
        const err = new Error("Upload aborted");
        err.name = "AbortError";
        reject(err);
      };
      if (controller && controller.signal) {
        controller.signal.addEventListener("abort", function() {
          xhr.abort();
        });
      }
      xhr.send(file);
    });
  }

  function uploadFile(file) {
    if (file.size > MULTIPART_THRESHOLD) {
      return uploadMultipart(file);
    }
    clearSelectedResumeUploadId();
    return uploadSingle(file);
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

  function findResumableSession() {
    let candidates = [];
    for (let i = 0; i < localStorage.length; i++) {
      const key = localStorage.key(i);
      if (!key || key.indexOf("upload:") !== 0) {
        continue;
      }
      if (key.indexOf("upload:signature:") === 0 || key.indexOf("upload:file:") === 0 || key === "upload:resume-id") {
        continue;
      }
      const session = getUploadSession(key.replace("upload:", ""));
      if (!session || session.mode !== "multipart" || !session.uploadId || !session.totalParts) {
        continue;
      }
      if (session.status && session.status !== "paused" && session.status !== "uploading") {
        continue;
      }
      const uploadedParts = session.uploadedParts || [];
      if (uploadedParts.length >= session.totalParts) {
        continue;
      }
      candidates.push(session);
    }
    if (!candidates.length) {
      return null;
    }
    candidates.sort(function(a, b) {
      const aTime = a.updatedAt || 0;
      const bTime = b.updatedAt || 0;
      return bTime - aTime;
    });
    return candidates[0];
  }

  function sessionProgress(session) {
    if (!session) {
      return 0;
    }
    const uploadedParts = session.uploadedParts || [];
    if (session.sizeBytes > 0) {
      let bytes = 0;
      uploadedParts.forEach(function(part) {
        if (part && part.size) {
          bytes += part.size;
        }
      });
      return Math.max(0, Math.min(100, Math.round((bytes / session.sizeBytes) * 100)));
    }
    if (session.totalParts > 0) {
      return Math.max(0, Math.min(100, Math.round((uploadedParts.length / session.totalParts) * 100)));
    }
    return 0;
  }

  function updateResumeBanner() {
    const banner = document.getElementById("upload-resume-banner");
    if (!banner) {
      return;
    }
    if (active && !paused) {
      banner.classList.add("is-hidden");
      return;
    }
    const meta = document.getElementById("resume-banner-meta");
    const session = findResumableSession();
    if (!session) {
      banner.classList.add("is-hidden");
      return;
    }
    const percent = sessionProgress(session);
    const filename = session.filename || "Pending upload";
    if (meta) {
      meta.textContent = filename + " • " + percent + "%";
    }
    banner.classList.remove("is-hidden");
    banner.setAttribute("data-upload-id", session.uploadId);
  }

  function resetSelection() {
    setSelectedFile(null);
    updateMeta("", 0, 0);
    if (input) {
      input.value = "";
    }
  }

  function beginUpload(file) {
    if (!file) {
      return;
    }
    if (file.size > MAX_FILE_SIZE) {
      setStatus("File exceeds the 1GB limit.");
      return;
    }
    transferStats = null;
    setStatus("Preparing upload...");
    updateMeta(file.name, 0, file.size);
    if (startButton) {
      startButton.disabled = true;
    }
    abortButton && (abortButton.disabled = false);
    setPaused(false);
    uploadFile(file)
      .catch(function() {})
      .finally(function() {
        if (startButton) {
          startButton.disabled = false;
        }
        if (abortButton) {
          abortButton.disabled = true;
        }
        updatePauseButtons();
        updateResumeBanner();
      });
  }

  if (confirmStart) {
    confirmStart.addEventListener("click", function() {
      if (!selectedFile) {
        closeConfirmDialog();
        return;
      }
      closeConfirmDialog();
      beginUpload(selectedFile);
      resetSelection();
    });
  }

  if (confirmCancel) {
    confirmCancel.addEventListener("click", function() {
      closeConfirmDialog();
    });
  }

  if (dropzone) {
    dropzone.addEventListener("click", function() {
      if (input) {
        input.click();
      }
    });
    dropzone.addEventListener("keydown", function(event) {
      if (event.key === "Enter" || event.key === " ") {
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
      const files = event.dataTransfer ? event.dataTransfer.files : null;
      if (files && files.length) {
        setSelectedFile(files[0], true);
      }
    });
  }

  if (chipClear) {
    chipClear.addEventListener("click", function() {
      resetSelection();
    });
  }

  input.addEventListener("change", function() {
    const resumeUploadId = getSelectedResumeUploadId();
    if (resumeUploadId) {
      if (!input.files || !input.files.length) {
        return;
      }
      const file = input.files[0];
      const session = getUploadSession(resumeUploadId);
      if (session && session.filename && session.sizeBytes) {
        if (session.filename !== file.name || session.sizeBytes !== file.size) {
          setStatus("Selected file does not match the paused upload. Choose the same file to resume.");
          clearSelectedResumeUploadId();
          return;
        }
      }
      clearSelectedResumeUploadId();
      beginUpload(file);
      return;
    }
    if (!input.files || !input.files.length) {
      resetSelection();
      return;
    }
    setSelectedFile(input.files[0], true);
  });

  if (pauseButton) {
    pauseButton.addEventListener("click", function() {
      if (!active || active.mode !== "multipart") {
        setStatus("Pause is only available for multipart uploads.");
        return;
      }
      setPaused(true);
      setStatus("Upload paused.");
    });
  }

  if (resumeButton) {
    resumeButton.addEventListener("click", function() {
      if (!active || active.mode !== "multipart") {
        setStatus("No multipart upload to resume.");
        return;
      }
      setPaused(false);
      setStatus("Resuming upload...");
    });
  }

  if (abortButton) {
    abortButton.addEventListener("click", function() {
      if (!active) {
        setStatus("No active upload to abort.");
        return;
      }
      abortButton.disabled = true;
      if (active.mode === "single") {
        if (active.controller) {
          active.controller.abort();
        }
        cancelUpload(active.uploadId)
          .then(function() {
            setStatus("Upload aborted.");
            toastUploadInfo("Upload aborted.", "Cancelled");
            setProgress(0);
            if (active && active.signature) {
              clearState(active.signature);
            }
            clearMeta();
            active = null;
            setPaused(false);
          })
          .catch(function() {
            setStatus("Abort failed. Try again.");
          })
          .finally(function() {
            abortButton.disabled = false;
            updatePauseButtons();
          });
        return;
      }

      if (active.cancel) {
        active.cancel();
      }
      cancelUpload(active.uploadId)
        .then(function() {
          if (active && active.signature) {
            clearState(active.signature);
          }
          setStatus("Upload aborted.");
          toastUploadInfo("Upload aborted.", "Cancelled");
          setProgress(0);
          clearMeta();
          active = null;
          setPaused(false);
        })
        .catch(function() {
          setStatus("Abort failed. Try again.");
        })
        .finally(function() {
          abortButton.disabled = false;
          updatePauseButtons();
        });
    });
  }

  const resumeBanner = document.getElementById("upload-resume-banner");
  const resumeBannerResume = document.getElementById("resume-banner-resume");
  const resumeBannerCancel = document.getElementById("resume-banner-cancel");
  if (resumeBanner && resumeBannerResume) {
    resumeBannerResume.addEventListener("click", function() {
      const uploadId = resumeBanner.getAttribute("data-upload-id");
      if (!uploadId) {
        return;
      }
      setSelectedResumeUploadId(uploadId);
      input.click();
    });
  }
  if (resumeBanner && resumeBannerCancel) {
    resumeBannerCancel.addEventListener("click", function() {
      const uploadId = resumeBanner.getAttribute("data-upload-id");
      if (!uploadId) {
        return;
      }
      resumeBannerCancel.disabled = true;
      cancelUpload(uploadId)
        .then(function() {
          clearStateByUploadId(uploadId);
          setStatus("Upload cancelled.");
          toastUploadInfo("Upload cancelled.", "Cancelled");
          clearMeta();
          updateResumeBanner();
        })
        .catch(function() {
          setStatus("Cancel failed. Try again.");
        })
        .finally(function() {
          resumeBannerCancel.disabled = false;
        });
    });
  }

  updatePauseButtons();
  updateResumeBanner();
})();
