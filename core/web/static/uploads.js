(function() {
  const input = document.getElementById("upload-file");
  const startButton = document.getElementById("upload-start");
  const pauseButton = document.getElementById("upload-pause");
  const resumeButton = document.getElementById("upload-resume");
  const abortButton = document.getElementById("upload-abort");
  const progress = document.getElementById("upload-progress");
  const status = document.getElementById("upload-status");

  if (!input || !startButton || !progress || !status) {
    return;
  }

  const progressBar = progress.querySelector(".progress-bar");
  const progressPercent = progress.querySelector(".progress-percent");
  let active = null;
  let paused = false;
  let resumeWaiters = [];
  const MAX_FILE_SIZE = 1024 * 1024 * 1024;
  const MULTIPART_THRESHOLD = 200 * 1024 * 1024;

  function setStatus(message) {
    status.textContent = message;
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

  function updatePauseButtons() {
    if (!pauseButton || !resumeButton) {
      return;
    }
    if (!active || active.mode !== "multipart") {
      pauseButton.disabled = true;
      resumeButton.disabled = true;
      return;
    }
    if (paused) {
      pauseButton.disabled = true;
      resumeButton.disabled = false;
      return;
    }
    pauseButton.disabled = false;
    resumeButton.disabled = true;
  }

  function setPaused(next) {
    paused = next;
    if (!paused) {
      resumeWaiters.forEach(function(resolve) { resolve(); });
      resumeWaiters = [];
    }
    updatePauseButtons();
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
    const stored = {
      uploadId: state.uploadId,
      multipartId: state.multipartId,
      fileId: state.fileId,
      mode: state.mode,
      chunkSize: state.chunkSize,
      totalParts: state.totalParts,
      uploadedParts: state.uploadedParts || [],
      signature: signature
    };
    localStorage.setItem("upload:" + state.uploadId, JSON.stringify(stored));
    if (state.fileId) {
      localStorage.setItem("upload:file:" + state.fileId, state.uploadId);
    }
    if (signature) {
      localStorage.setItem("upload:signature:" + signature, state.uploadId);
    }
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
  }

  function clearStateForFile(fileId) {
    if (!fileId) {
      return;
    }
    const uploadId = localStorage.getItem("upload:file:" + fileId);
    if (uploadId) {
      const raw = localStorage.getItem("upload:" + uploadId);
      let signature = null;
      if (raw) {
        try {
          signature = JSON.parse(raw).signature;
        } catch (_) {
          signature = null;
        }
      }
      localStorage.removeItem("upload:" + uploadId);
      localStorage.removeItem("upload:file:" + fileId);
      if (signature) {
        localStorage.removeItem("upload:signature:" + signature);
      }
      return;
    }
    localStorage.removeItem("upload:file:" + fileId);
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

  function startMultipart(file) {
    return api("/api/uploads/multipart/start", {
      filename: file.name,
      size: file.size,
      contentType: file.type || "application/octet-stream"
    });
  }

  function startSingle(file) {
    return api("/api/uploads/single/start", {
      filename: file.name,
      size: file.size,
      contentType: file.type || "application/octet-stream"
    });
  }

  function completeSingle(fileId) {
    return api("/api/uploads/single/complete", { fileId: fileId });
  }

  function abortSingle(fileId) {
    return api("/api/uploads/single/abort", { fileId: fileId });
  }

  function partURL(multipartId, partNumber) {
    return api("/api/uploads/multipart/part-url", {
      multipartId: multipartId,
      partNumber: partNumber
    });
  }

  function resumeMultipart(fileId) {
    return api("/api/uploads/multipart/resume?fileId=" + encodeURIComponent(fileId), null, "GET");
  }

  function setSelectedResumeFileId(fileId) {
    if (!fileId) {
      localStorage.removeItem("upload:resume-file");
      return;
    }
    localStorage.setItem("upload:resume-file", fileId);
    setStatus("Selected a pending upload. Now choose the same file to resume.");
  }

  function getSelectedResumeFileId() {
    return localStorage.getItem("upload:resume-file");
  }

  function clearSelectedResumeFileId() {
    localStorage.removeItem("upload:resume-file");
  }

  function completeMultipart(multipartId, parts) {
    return api("/api/uploads/multipart/complete", {
      multipartId: multipartId,
      parts: parts
    });
  }

  function abortMultipart(multipartId) {
    return api("/api/uploads/multipart/abort", {
      multipartId: multipartId
    });
  }

  function abortByFileId(fileId) {
    return api("/api/uploads/abort", { fileId: fileId });
  }

  async function cleanupFailure(fileId, signature) {
    clearSelectedResumeFileId();
    if (signature) {
      clearState(signature);
    }
    if (fileId) {
      clearStateForFile(fileId);
    }
    if (active && active.cancel) {
      active.cancel();
    }
    active = null;
    setPaused(false);
    setProgress(0);
    updatePauseButtons();
    if (!fileId) {
      return;
    }
    try {
      await abortByFileId(fileId);
    } catch (_) {}
  }

  function uploadPartWithRetry(multipartId, partNumber, chunk, maxRetries) {
    let attempt = 0;

    function tryUpload() {
      attempt++;
      return partURL(multipartId, partNumber)
        .then(function(res) {
          return fetch(res.url, { method: "PUT", body: chunk });
        })
        .then(function(res) {
          if (!res.ok) {
            throw new Error("Upload failed: " + res.status);
          }
          const etag = res.headers.get("ETag");
          if (!etag) {
            throw new Error("Missing ETag");
          }
          return { etag: etag };
        })
        .catch(function(err) {
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

  async function uploadPartsByNumber(multipartId, chunkSize, file, parts, uploadedMap, uploadedBytes, signature, fileId, totalParts, maxRetries) {
    const queue = parts.slice();
    queue.sort(function(a, b) { return a - b; });

    for (let i = 0; i < queue.length; i++) {
      const partNumber = queue[i];
      if (uploadedMap.has(partNumber)) {
        continue;
      }
      const startByte = (partNumber - 1) * chunkSize;
      const endByte = Math.min(startByte + chunkSize, file.size);
      const chunk = file.slice(startByte, endByte);
      const result = await uploadPartWithRetry(multipartId, partNumber, chunk, maxRetries);
      uploadedMap.set(partNumber, result.etag);
      uploadedBytes += chunk.size;
    }
    const rebuilt = buildUploadedPartsFromMap(file, chunkSize, uploadedMap);
    saveState(signature, {
      uploadId: multipartId,
      multipartId: multipartId,
      fileId: fileId,
      mode: "multipart",
      chunkSize: chunkSize,
      totalParts: totalParts,
      uploadedParts: rebuilt.parts
    });
    return { uploadedParts: rebuilt.parts, uploadedBytes: rebuilt.bytes };
  }

  async function uploadMultipart(file) {
    const signature = fileSignature(file);
    let existing = loadState(signature);
    let canceled = false;
    let state = null;

    async function initSession() {
      if (existing && existing.mode === "multipart" && (existing.multipartId || existing.uploadId) && existing.chunkSize && existing.totalParts) {
        const existingUploadId = existing.multipartId || existing.uploadId;
        if (existing.fileId) {
          try {
            const resumedExisting = await resumeMultipart(existing.fileId);
            if (resumedExisting.filename && resumedExisting.sizeBytes) {
              if (resumedExisting.filename !== file.name || resumedExisting.sizeBytes !== file.size) {
                clearState(signature);
                existing = null;
                setStatus("Could not resume this upload. Starting a new upload.");
                throw new Error("Resume file mismatch");
              }
            }
            const resumedState = {
              uploadId: resumedExisting.multipartId,
              fileId: resumedExisting.fileId,
              multipartId: resumedExisting.multipartId,
              chunkSize: resumedExisting.chunkSize,
              totalParts: resumedExisting.totalParts,
              uploadedParts: resumedExisting.uploadedParts || [],
              mode: "multipart"
            };
            saveState(signature, resumedState);
            return resumedState;
          } catch (_) {
            clearState(signature);
            existing = null;
            setStatus("Could not resume this upload. Starting a new upload.");
          }
        }
        if (existing) {
          existing.uploadId = existingUploadId;
          existing.multipartId = existingUploadId;
          existing.mode = "multipart";
          return existing;
        }
      }
      const resumeFileId = getSelectedResumeFileId();
      if (resumeFileId) {
        try {
          const resumed = await resumeMultipart(resumeFileId);
          if (resumed.filename && resumed.sizeBytes) {
            if (resumed.filename !== file.name || resumed.sizeBytes !== file.size) {
              clearSelectedResumeFileId();
              setStatus("Selected upload does not match this file. Choose the same file used to start the upload.");
              throw new Error("Resume file mismatch");
            }
          }
          const resumedState_1 = {
            uploadId: resumed.multipartId,
            fileId: resumed.fileId,
            multipartId: resumed.multipartId,
            chunkSize: resumed.chunkSize,
            totalParts: resumed.totalParts,
            uploadedParts: resumed.uploadedParts || [],
            mode: "multipart"
          };
          saveState(signature, resumedState_1);
          clearSelectedResumeFileId();
          return resumedState_1;
        } catch (_) {
          clearSelectedResumeFileId();
        }
      }
      const resp = await startMultipart(file);
      const state = {
        uploadId: resp.multipartId,
        fileId: resp.fileId,
        multipartId: resp.multipartId,
        chunkSize: resp.chunkSize,
        totalParts: resp.totalParts,
        uploadedParts: [],
        mode: "multipart"
      };
      saveState(signature, state);
      return state;
    }

    try {
      state = await initSession();
      const multipartId = state.multipartId;
      const chunkSize = state.chunkSize;
      const totalParts = state.totalParts;
      const uploadedMap = new Map((state.uploadedParts || []).map(function(p) { return [p.partNumber, p.etag]; }));
      const rebuilt = buildUploadedPartsFromMap(file, chunkSize, uploadedMap);
      let uploadedParts = rebuilt.parts;
      let uploadedBytes = rebuilt.bytes;
      saveState(signature, {
        uploadId: multipartId,
        multipartId: multipartId,
        fileId: state.fileId,
        mode: "multipart",
        chunkSize: chunkSize,
        totalParts: totalParts,
        uploadedParts: uploadedParts
      });

      active = {
        mode: "multipart",
        uploadId: multipartId,
        multipartId: multipartId,
        signature: signature,
        fileId: state.fileId,
        cancel: function() { canceled = true; }
      };
      setProgress((uploadedBytes / file.size) * 100);
      setStatus("Uploading " + file.name + "...");
      updatePauseButtons();

      for (let partNumber = 1; partNumber <= totalParts; partNumber++) {
        if (canceled) {
          return;
        }
        if (uploadedMap.has(partNumber)) {
          continue;
        }
        await waitForResume();
        if (canceled) {
          return;
        }

        const startByte = (partNumber - 1) * chunkSize;
        const endByte = Math.min(startByte + chunkSize, file.size);
        const chunk = file.slice(startByte, endByte);

        try {
          const result = await uploadPartWithRetry(multipartId, partNumber, chunk, 5);
          uploadedMap.set(partNumber, result.etag);
          uploadedBytes += chunk.size;
          uploadedParts.push({ partNumber: partNumber, etag: result.etag, size: chunk.size });
          saveState(signature, {
            uploadId: multipartId,
            multipartId: multipartId,
            fileId: state.fileId,
            mode: "multipart",
            chunkSize: chunkSize,
            totalParts: totalParts,
            uploadedParts: uploadedParts
          });
          setProgress((uploadedBytes / file.size) * 100);
        } catch (err) {
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

      let parts = uploadedParts
        .map(function(p) { return { partNumber: p.partNumber, etag: p.etag }; })
        .sort(function(a, b) { return a.partNumber - b.partNumber; });

      try {
        await completeMultipart(multipartId, parts);
      } catch (err) {
        const missingParts = err && err.status === 409 && err.data && err.data.missingParts ? err.data.missingParts : null;
        if (!missingParts || !missingParts.length) {
          throw err;
        }
        setStatus("Recovering missing parts...");
        const recovered = await uploadPartsByNumber(multipartId, chunkSize, file, missingParts, uploadedMap, uploadedBytes, signature, state.fileId, totalParts, 5);
        uploadedParts = recovered.uploadedParts;
        uploadedBytes = recovered.uploadedBytes;
        setProgress((uploadedBytes / file.size) * 100);
        parts = Array.from(uploadedMap.entries())
          .map(function(entry) { return { partNumber: entry[0], etag: entry[1] }; })
          .sort(function(a, b) { return a.partNumber - b.partNumber; });
        await completeMultipart(multipartId, parts);
      }
      clearState(signature);
      setProgress(100);
      setStatus("Upload complete: " + file.name);
      active = null;
      setPaused(false);
    } catch (err) {
      const fileId = state && state.fileId ? state.fileId : (existing && existing.fileId ? existing.fileId : null);
      await cleanupFailure(fileId, signature);
      throw err;
    }
  }

  async function uploadSingle(file) {
    const controller = new AbortController();
    let fileId = null;
    const signature = fileSignature(file);

    try {
      const resp = await startSingle(file);
      fileId = resp.fileId;
      saveState(signature, {
        uploadId: resp.fileId,
        fileId: resp.fileId,
        mode: "single",
        uploadedParts: []
      });
      active = { mode: "single", uploadId: resp.fileId, fileId: resp.fileId, controller: controller, signature: signature };
      setProgress(0);
      setStatus("Uploading " + file.name + "...");
      updatePauseButtons();

      try {
        await uploadSingleWithProgress(resp.uploadUrl, file, controller);
        await completeSingle(resp.fileId);
        clearState(signature);
        setProgress(100);
        setStatus("Upload complete: " + file.name);
        active = null;
        setPaused(false);
      } catch (err) {
        if (err && err.name === "AbortError") {
          return;
        }
        const message = err && err.data && err.data.errors && err.data.errors.size ? err.data.errors.size : null;
        await cleanupFailure(resp.fileId, signature);
        fileId = null;
        setStatus("Upload failed. " + (message || (err && err.message ? err.message : "Try again.")));
        throw err;
      }
    } catch (err) {
      const message_2 = err && err.data && err.data.errors && err.data.errors.size ? err.data.errors.size : null;
      await cleanupFailure(fileId, signature);
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
    clearSelectedResumeFileId();
    return uploadSingle(file);
  }

  startButton.addEventListener("click", function() {
    if (!input.files || !input.files.length) {
      setStatus("Select a file to upload.");
      return;
    }

    startButton.disabled = true;
    abortButton && (abortButton.disabled = false);
    setPaused(false);
    const file = input.files[0];

    if (file.size > MAX_FILE_SIZE) {
      setStatus("File exceeds the 1GB limit.");
      startButton.disabled = false;
      abortButton && (abortButton.disabled = true);
      updatePauseButtons();
      return;
    }
    uploadFile(file)
      .catch(function() {})
      .finally(function() {
        startButton.disabled = false;
        if (abortButton) {
          abortButton.disabled = true;
        }
        updatePauseButtons();
      });
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
        abortSingle(active.fileId)
          .then(function() {
            setStatus("Upload aborted.");
            setProgress(0);
            if (active && active.signature) {
              clearState(active.signature);
            }
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
      abortMultipart(active.multipartId)
        .then(function() {
          if (active && active.signature) {
            clearState(active.signature);
          }
          setStatus("Upload aborted.");
          setProgress(0);
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

  const fileAbortButtons = document.querySelectorAll(".file-abort");
  if (fileAbortButtons.length) {
    fileAbortButtons.forEach(function(button) {
      button.addEventListener("click", function() {
        const fileId = button.getAttribute("data-file-id");
        if (!fileId) {
          return;
        }
        button.disabled = true;
        abortByFileId(fileId)
          .then(function() {
            const row = button.closest(".files-row");
            if (row) {
              row.remove();
            }
            clearStateForFile(fileId);
          })
          .catch(function() {
            button.disabled = false;
            setStatus("Abort failed. Try again.");
          });
      });
    });
  }

  const fileRows = document.querySelectorAll(".files-row");
  if (fileRows.length) {
    fileRows.forEach(function(row) {
      const fileId = row.getAttribute("data-file-id");
      const resume = row.querySelector(".files-resume");
      if (!resume || !fileId) {
        return;
      }
      const uploadId = localStorage.getItem("upload:file:" + fileId);
      if (uploadId) {
        const raw = localStorage.getItem("upload:" + uploadId);
        if (raw) {
          try {
            const parsed = JSON.parse(raw);
            if (parsed && parsed.mode === "multipart") {
              resume.textContent = "Resume ready on this device.";
              return;
            }
          } catch (_) {}
        }
      }
    });
  }

  const resumeButtons = document.querySelectorAll("[data-resume-id]");
  if (resumeButtons.length) {
    resumeButtons.forEach(function(button) {
      button.addEventListener("click", function() {
        const fileId = button.getAttribute("data-resume-id");
        if (!fileId) {
          return;
        }
        setSelectedResumeFileId(fileId);
      });
    });
  }

  updatePauseButtons();
})();
