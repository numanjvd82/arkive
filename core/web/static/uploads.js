(function() {
  var input = document.getElementById("upload-file");
  var startButton = document.getElementById("upload-start");
  var pauseButton = document.getElementById("upload-pause");
  var resumeButton = document.getElementById("upload-resume");
  var abortButton = document.getElementById("upload-abort");
  var progress = document.getElementById("upload-progress");
  var status = document.getElementById("upload-status");

  if (!input || !startButton || !progress || !status) {
    return;
  }

  var progressBar = progress.querySelector(".progress-bar");
  var progressPercent = progress.querySelector(".progress-percent");
  var active = null;
  var paused = false;
  var resumeWaiters = [];
  var MAX_FILE_SIZE = 1024 * 1024 * 1024;
  var MULTIPART_THRESHOLD = 200 * 1024 * 1024;

  function setStatus(message) {
    status.textContent = message;
  }

  function setProgress(percent) {
    var clamped = Math.max(0, Math.min(100, Math.round(percent)));
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
    if (!active || active.type !== "multipart") {
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
      var raw = localStorage.getItem("mp:" + signature);
      return raw ? JSON.parse(raw) : null;
    } catch (_) {
      return null;
    }
  }

  function saveState(signature, state) {
    localStorage.setItem("mp:" + signature, JSON.stringify(state));
    if (state && state.fileId) {
      localStorage.setItem("mp:file:" + state.fileId, signature);
    }
  }

  function clearState(signature) {
    var state = loadState(signature);
    if (state && state.fileId) {
      localStorage.removeItem("mp:file:" + state.fileId);
    }
    localStorage.removeItem("mp:" + signature);
  }

  function clearStateForFile(fileId) {
    if (!fileId) {
      return;
    }
    var signature = localStorage.getItem("mp:file:" + fileId);
    if (signature) {
      clearState(signature);
    } else {
      localStorage.removeItem("mp:file:" + fileId);
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
      localStorage.removeItem("mp:resume-file");
      return;
    }
    localStorage.setItem("mp:resume-file", fileId);
    setStatus("Selected a pending upload. Now choose the same file to resume.");
  }

  function getSelectedResumeFileId() {
    return localStorage.getItem("mp:resume-file");
  }

  function clearSelectedResumeFileId() {
    localStorage.removeItem("mp:resume-file");
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
    var attempt = 0;

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
          var etag = res.headers.get("ETag");
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
    var parts = [];
    var bytes = 0;
    var numbers = Array.from(uploadedMap.keys()).sort(function(a, b) { return a - b; });
    for (var i = 0; i < numbers.length; i++) {
      var partNumber = numbers[i];
      var startByte = (partNumber - 1) * chunkSize;
      var endByte = Math.min(startByte + chunkSize, file.size);
      var size = Math.max(0, endByte - startByte);
      parts.push({ partNumber: partNumber, etag: uploadedMap.get(partNumber), size: size });
      bytes += size;
    }
    return { parts: parts, bytes: bytes };
  }

  async function uploadPartsByNumber(multipartId, chunkSize, file, parts, uploadedMap, uploadedBytes, signature, fileId, totalParts, maxRetries) {
    var queue = parts.slice();
    queue.sort(function(a, b) { return a - b; });

    for (var i = 0; i < queue.length; i++) {
      var partNumber = queue[i];
      if (uploadedMap.has(partNumber)) {
        continue;
      }
      var startByte = (partNumber - 1) * chunkSize;
      var endByte = Math.min(startByte + chunkSize, file.size);
      var chunk = file.slice(startByte, endByte);
      var result = await uploadPartWithRetry(multipartId, partNumber, chunk, maxRetries);
      uploadedMap.set(partNumber, result.etag);
      uploadedBytes += chunk.size;
    }
    var rebuilt = buildUploadedPartsFromMap(file, chunkSize, uploadedMap);
    saveState(signature, {
      fileId: fileId,
      multipartId: multipartId,
      chunkSize: chunkSize,
      totalParts: totalParts,
      uploadedParts: rebuilt.parts
    });
    return { uploadedParts: rebuilt.parts, uploadedBytes: rebuilt.bytes };
  }

  async function uploadMultipart(file) {
    var signature = fileSignature(file);
    var existing = loadState(signature);
    var canceled = false;
    var state = null;

    async function initSession() {
      if (existing && existing.multipartId && existing.chunkSize && existing.totalParts) {
        if (existing.fileId) {
          try {
            var resumedExisting = await resumeMultipart(existing.fileId);
            if (resumedExisting.filename && resumedExisting.sizeBytes) {
              if (resumedExisting.filename !== file.name || resumedExisting.sizeBytes !== file.size) {
                clearState(signature);
                existing = null;
                setStatus("Could not resume this upload. Starting a new upload.");
                throw new Error("Resume file mismatch");
              }
            }
            var resumedState = {
              fileId: resumedExisting.fileId,
              multipartId: resumedExisting.multipartId,
              chunkSize: resumedExisting.chunkSize,
              totalParts: resumedExisting.totalParts,
              uploadedParts: resumedExisting.uploadedParts || []
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
          return existing;
        }
      }
      var resumeFileId = getSelectedResumeFileId();
      if (resumeFileId) {
        try {
          var resumed = await resumeMultipart(resumeFileId);
          if (resumed.filename && resumed.sizeBytes) {
            if (resumed.filename !== file.name || resumed.sizeBytes !== file.size) {
              clearSelectedResumeFileId();
              setStatus("Selected upload does not match this file. Choose the same file used to start the upload.");
              throw new Error("Resume file mismatch");
            }
          }
          var resumedState_1 = {
            fileId: resumed.fileId,
            multipartId: resumed.multipartId,
            chunkSize: resumed.chunkSize,
            totalParts: resumed.totalParts,
            uploadedParts: resumed.uploadedParts || []
          };
          saveState(signature, resumedState_1);
          clearSelectedResumeFileId();
          return resumedState_1;
        } catch (_) {
          clearSelectedResumeFileId();
        }
      }
      var resp = await startMultipart(file);
      var state = {
        fileId: resp.fileId,
        multipartId: resp.multipartId,
        chunkSize: resp.chunkSize,
        totalParts: resp.totalParts,
        uploadedParts: []
      };
      saveState(signature, state);
      return state;
    }

    try {
      state = await initSession();
      var multipartId = state.multipartId;
      var chunkSize = state.chunkSize;
      var totalParts = state.totalParts;
      var uploadedMap = new Map((state.uploadedParts || []).map(function(p) { return [p.partNumber, p.etag]; }));
      var rebuilt = buildUploadedPartsFromMap(file, chunkSize, uploadedMap);
      var uploadedParts = rebuilt.parts;
      var uploadedBytes = rebuilt.bytes;
      saveState(signature, {
        fileId: state.fileId,
        multipartId: multipartId,
        chunkSize: chunkSize,
        totalParts: totalParts,
        uploadedParts: uploadedParts
      });

      active = {
        type: "multipart",
        multipartId: multipartId,
        signature: signature,
        fileId: state.fileId,
        cancel: function() { canceled = true; }
      };
      setProgress((uploadedBytes / file.size) * 100);
      setStatus("Uploading " + file.name + "...");
      updatePauseButtons();

      for (var partNumber = 1; partNumber <= totalParts; partNumber++) {
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

        var startByte = (partNumber - 1) * chunkSize;
        var endByte = Math.min(startByte + chunkSize, file.size);
        var chunk = file.slice(startByte, endByte);

        try {
          var result = await uploadPartWithRetry(multipartId, partNumber, chunk, 5);
          uploadedMap.set(partNumber, result.etag);
          uploadedBytes += chunk.size;
          uploadedParts.push({ partNumber: partNumber, etag: result.etag, size: chunk.size });
          saveState(signature, {
            fileId: state.fileId,
            multipartId: multipartId,
            chunkSize: chunkSize,
            totalParts: totalParts,
            uploadedParts: uploadedParts
          });
          setProgress((uploadedBytes / file.size) * 100);
        } catch (err) {
          if (err && err.status === 404 && existing) {
            clearState(signature);
          }
          var message = err && err.data && err.data.errors && err.data.errors.size ? err.data.errors.size : null;
          setStatus("Upload failed. " + (message || (err && err.message ? err.message : "Try again.")));
          throw err;
        }
      }

      if (canceled) {
        return;
      }

      var parts = uploadedParts
        .map(function(p) { return { partNumber: p.partNumber, etag: p.etag }; })
        .sort(function(a, b) { return a.partNumber - b.partNumber; });

      try {
        await completeMultipart(multipartId, parts);
      } catch (err) {
        var missingParts = err && err.status === 409 && err.data && err.data.missingParts ? err.data.missingParts : null;
        if (!missingParts || !missingParts.length) {
          throw err;
        }
        setStatus("Recovering missing parts...");
        var recovered = await uploadPartsByNumber(multipartId, chunkSize, file, missingParts, uploadedMap, uploadedBytes, signature, state.fileId, totalParts, 5);
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
      var fileId = state && state.fileId ? state.fileId : (existing && existing.fileId ? existing.fileId : null);
      await cleanupFailure(fileId, signature);
      throw err;
    }
  }

  async function uploadSingle(file) {
    var controller = new AbortController();
    var fileId = null;

    try {
      var resp = await startSingle(file);
      fileId = resp.fileId;
      active = { type: "single", fileId: resp.fileId, controller: controller };
      setProgress(0);
      setStatus("Uploading " + file.name + "...");
      updatePauseButtons();

      try {
        await uploadSingleWithProgress(resp.uploadUrl, file, controller);
        await completeSingle(resp.fileId);
        setProgress(100);
        setStatus("Upload complete: " + file.name);
        active = null;
        setPaused(false);
      } catch (err) {
        if (err && err.name === "AbortError") {
          return;
        }
        var message = err && err.data && err.data.errors && err.data.errors.size ? err.data.errors.size : null;
        await cleanupFailure(resp.fileId, "");
        fileId = null;
        setStatus("Upload failed. " + (message || (err && err.message ? err.message : "Try again.")));
        throw err;
      }
    } catch (err) {
      var message_2 = err && err.data && err.data.errors && err.data.errors.size ? err.data.errors.size : null;
      await cleanupFailure(fileId, "");
      setStatus("Upload failed. " + (message_2 || (err && err.message ? err.message : "Try again.")));
      throw err;
    }
  }

  function uploadSingleWithProgress(url, file, controller) {
    return new Promise(function(resolve, reject) {
      var xhr = new XMLHttpRequest();
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
        var err = new Error("Upload aborted");
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

  startButton.addEventListener("click", function() {
    if (!input.files || !input.files.length) {
      setStatus("Select a file to upload.");
      return;
    }

    startButton.disabled = true;
    abortButton && (abortButton.disabled = false);
    setPaused(false);
    var file = input.files[0];

    if (file.size > MAX_FILE_SIZE) {
      setStatus("File exceeds the 1GB limit.");
      startButton.disabled = false;
      abortButton && (abortButton.disabled = true);
      updatePauseButtons();
      return;
    }
    var uploader = file.size > MULTIPART_THRESHOLD ? uploadMultipart : uploadSingle;
    if (uploader === uploadSingle) {
      clearSelectedResumeFileId();
    }

    uploader(file)
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
      if (!active || active.type !== "multipart") {
        setStatus("Pause is only available for multipart uploads.");
        return;
      }
      setPaused(true);
      setStatus("Upload paused.");
    });
  }

  if (resumeButton) {
    resumeButton.addEventListener("click", function() {
      if (!active || active.type !== "multipart") {
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
      if (active.type === "single") {
        if (active.controller) {
          active.controller.abort();
        }
        abortSingle(active.fileId)
          .then(function() {
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
        return;
      }

      if (active.cancel) {
        active.cancel();
      }
      abortMultipart(active.multipartId)
        .then(function() {
          clearState(active.signature);
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

  var fileAbortButtons = document.querySelectorAll(".file-abort");
  if (fileAbortButtons.length) {
    fileAbortButtons.forEach(function(button) {
      button.addEventListener("click", function() {
        var fileId = button.getAttribute("data-file-id");
        if (!fileId) {
          return;
        }
        button.disabled = true;
        abortByFileId(fileId)
          .then(function() {
            var row = button.closest(".files-row");
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

  var fileRows = document.querySelectorAll(".files-row");
  if (fileRows.length) {
    fileRows.forEach(function(row) {
      var fileId = row.getAttribute("data-file-id");
      var resume = row.querySelector(".files-resume");
      if (!resume || !fileId) {
        return;
      }
      if (localStorage.getItem("mp:file:" + fileId)) {
        resume.textContent = "Resume ready on this device.";
      }
    });
  }

  var resumeButtons = document.querySelectorAll("[data-resume-id]");
  if (resumeButtons.length) {
    resumeButtons.forEach(function(button) {
      button.addEventListener("click", function() {
        var fileId = button.getAttribute("data-resume-id");
        if (!fileId) {
          return;
        }
        setSelectedResumeFileId(fileId);
      });
    });
  }

  updatePauseButtons();
})();
