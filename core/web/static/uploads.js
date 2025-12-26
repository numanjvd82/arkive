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
  }

  function clearState(signature) {
    localStorage.removeItem("mp:" + signature);
  }

  async function api(path, body, method) {
    var res = await fetch(path, {
      method: method || "POST",
      headers: { "Content-Type": "application/json" },
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

  async function uploadMultipart(file) {
    var signature = fileSignature(file);
    var existing = loadState(signature);
    var canceled = false;

    async function initSession() {
      if (existing && existing.multipartId && existing.chunkSize && existing.totalParts) {
        return existing;
      }
      var resp = await startMultipart(file);
      var state = {
        multipartId: resp.multipartId,
        chunkSize: resp.chunkSize,
        totalParts: resp.totalParts,
        uploadedParts: []
      };
      saveState(signature, state);
      return state;
    }

    var state = await initSession();
    var multipartId = state.multipartId;
    var chunkSize = state.chunkSize;
    var totalParts = state.totalParts;
    var uploadedParts = state.uploadedParts || [];
    var uploadedMap = new Map(uploadedParts.map(function(p) { return [p.partNumber, p.etag]; }));
    var uploadedBytes = uploadedParts.reduce(function(sum, p) { return sum + p.size; }, 0);

    active = { type: "multipart", multipartId: multipartId, signature: signature, cancel: function() { canceled = true; } };
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

    await completeMultipart(multipartId, parts);
    clearState(signature);
    setProgress(100);
    setStatus("Upload complete: " + file.name);
    active = null;
    setPaused(false);
  }

  async function uploadSingle(file) {
    var controller = new AbortController();

    try {
      var resp = await startSingle(file);
      active = { type: "single", fileId: resp.fileId, controller: controller };
      setProgress(0);
      setStatus("Uploading " + file.name + "...");
      updatePauseButtons();

      try {
        var res = await fetch(resp.uploadUrl, {
          method: "PUT",
          body: file,
          signal: controller.signal
        });
        if (!res.ok) {
          throw new Error("Upload failed: " + res.status);
        }
        await completeSingle(resp.fileId);
        setProgress(100);
        setStatus("Upload complete: " + file.name);
        active = null;
        setPaused(false);
      } catch (err) {
        if (err && err.name === "AbortError") {
          return;
        }
        try {
          await abortSingle(resp.fileId);
        } catch (_) {}
        var message = err && err.data && err.data.errors && err.data.errors.size ? err.data.errors.size : null;
        setStatus("Upload failed. " + (message || (err && err.message ? err.message : "Try again.")));
        throw err;
      }
    } catch (err) {
      var message_2 = err && err.data && err.data.errors && err.data.errors.size ? err.data.errors.size : null;
      setStatus("Upload failed. " + (message_2 || (err && err.message ? err.message : "Try again.")));
      throw err;
    }
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
          })
          .catch(function() {
            button.disabled = false;
          });
      });
    });
  }

  updatePauseButtons();
})();
