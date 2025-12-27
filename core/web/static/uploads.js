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

  function completeUpload(uploadId, parts) {
    return api("/api/uploads/" + encodeURIComponent(uploadId) + "/complete", {
      parts: parts || []
    });
  }

  function cancelUpload(uploadId) {
    return api("/api/uploads/" + encodeURIComponent(uploadId) + "/cancel", {});
  }

  async function cleanupFailure(uploadId, signature) {
    clearSelectedResumeFileId();
    if (signature) {
      clearState(signature);
    }
    if (uploadId) {
      clearStateForFile(uploadId);
    }
    if (active && active.cancel) {
      active.cancel();
    }
    active = null;
    setPaused(false);
    setProgress(0);
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
        saveState(signature, {
          uploadId: uploadId,
          fileId: fileId,
          mode: "multipart",
          chunkSize: chunkSize,
          totalParts: totalParts,
          uploadedParts: rebuilt.parts
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
      const resumeFileId = getSelectedResumeFileId();
      if (resumeFileId) {
        try {
          const next = await nextUpload(resumeFileId, []);
          if (!next || next.mode !== "multipart") {
            throw new Error("Resume file mismatch");
          }
          const resumedState = {
            uploadId: next.uploadId || resumeFileId,
            fileId: next.fileId,
            chunkSize: next.chunkSize,
            totalParts: next.totalParts,
            uploadedParts: next.uploadedParts || [],
            mode: "multipart"
          };
          saveState(signature, resumedState);
          clearSelectedResumeFileId();
          return resumedState;
        } catch (_) {
          clearSelectedResumeFileId();
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
        mode: "multipart"
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
      saveState(signature, {
        uploadId: uploadId,
        fileId: state.fileId,
        mode: "multipart",
        chunkSize: chunkSize,
        totalParts: totalParts,
        uploadedParts: rebuilt.parts
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
          saveState(signature, {
            uploadId: uploadId,
            fileId: state.fileId,
            mode: "multipart",
            chunkSize: chunkSize,
            totalParts: totalParts,
            uploadedParts: updated.parts
          });
          setProgress((updated.bytes / file.size) * 100);
        } catch (err) {
          if (err && err.cancelled) {
            clearState(signature);
            setStatus("Upload cancelled.");
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
        setProgress((refreshed.bytes / file.size) * 100);
        parts = Array.from(uploadedMap.entries())
          .map(function(entry) { return { partNumber: entry[0], etag: entry[1] }; })
          .sort(function(a, b) { return a.partNumber - b.partNumber; });
        await completeUpload(uploadId, parts);
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
        uploadedParts: []
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
        cancelUpload(active.uploadId)
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
      cancelUpload(active.uploadId)
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
        cancelUpload(fileId)
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
