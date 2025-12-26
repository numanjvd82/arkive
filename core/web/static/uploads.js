(function() {
  var input = document.getElementById("upload-file");
  var startButton = document.getElementById("upload-start");
  var abortButton = document.getElementById("upload-abort");
  var progress = document.getElementById("upload-progress");
  var status = document.getElementById("upload-status");

  if (!input || !startButton || !progress || !status) {
    return;
  }

  var progressBar = progress.querySelector(".progress-bar");
  var active = null;

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

  async function api(path, body) {
    const res = await fetch(path, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body)
    });
    if (!res.ok) {
      var err = new Error("Request failed");
      err.status = res.status;
      throw err;
    }
    if (res.status === 204) {
      return null;
    }
    return res.json();
  }

  function startMultipart(file) {
    return api("/api/uploads/multipart/start", {
      filename: file.name,
      size: file.size,
      contentType: file.type || "application/octet-stream"
    });
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

  function uploadMultipart(file) {
    var signature = fileSignature(file);
    var existing = loadState(signature);

    async function initSession() {
      if (existing && existing.multipartId && existing.chunkSize && existing.totalParts) {
        return Promise.resolve(existing);
      }
      const resp = await startMultipart(file);
      var state = {
        multipartId: resp.multipartId,
        chunkSize: resp.chunkSize,
        totalParts: resp.totalParts,
        uploadedParts: []
      };
      saveState(signature, state);
      return state;
    }

    return initSession().then(function(state) {
      var multipartId = state.multipartId;
      var chunkSize = state.chunkSize;
      var totalParts = state.totalParts;
      var uploadedParts = state.uploadedParts || [];
      var uploadedMap = new Map(uploadedParts.map(function(p) { return [p.partNumber, p.etag]; }));
      var uploadedBytes = uploadedParts.reduce(function(sum, p) { return sum + p.size; }, 0);

      active = { multipartId: multipartId, signature: signature };
      setProgress((uploadedBytes / file.size) * 100);
      setStatus("Uploading " + file.name + "...");

      var sequence = Promise.resolve();

      for (var partNumber = 1; partNumber <= totalParts; partNumber++) {
        if (uploadedMap.has(partNumber)) {
          continue;
        }
        (function(part) {
          sequence = sequence.then(function() {
            var startByte = (part - 1) * chunkSize;
            var endByte = Math.min(startByte + chunkSize, file.size);
            var chunk = file.slice(startByte, endByte);

            return uploadPartWithRetry(multipartId, part, chunk, 5)
              .then(function(result) {
                uploadedMap.set(part, result.etag);
                uploadedBytes += chunk.size;
                uploadedParts.push({ partNumber: part, etag: result.etag, size: chunk.size });
                saveState(signature, {
                  multipartId: multipartId,
                  chunkSize: chunkSize,
                  totalParts: totalParts,
                  uploadedParts: uploadedParts
                });
                setProgress((uploadedBytes / file.size) * 100);
              })
              .catch(function(err) {
                if (err && err.status === 404 && existing) {
                  clearState(signature);
                }
                throw err;
              });
          });
        })(partNumber);
      }

      return sequence
        .then(function() {
          var parts = uploadedParts
            .map(function(p) { return { partNumber: p.partNumber, etag: p.etag }; })
            .sort(function(a, b) { return a.partNumber - b.partNumber; });

          return completeMultipart(multipartId, parts).then(function() {
            clearState(signature);
            setProgress(100);
            setStatus("Upload complete: " + file.name);
            active = null;
          });
        })
        .catch(function(err) {
          setStatus("Upload failed. " + (err && err.message ? err.message : "Try again."));
          throw err;
        });
    });
  }

  startButton.addEventListener("click", function() {
    if (!input.files || !input.files.length) {
      setStatus("Select a file to upload.");
      return;
    }

    startButton.disabled = true;
    abortButton && (abortButton.disabled = false);
    var file = input.files[0];

    uploadMultipart(file)
      .catch(function() {})
      .finally(function() {
        startButton.disabled = false;
        if (abortButton) {
          abortButton.disabled = true;
        }
      });
  });

  if (abortButton) {
    abortButton.addEventListener("click", function() {
      if (!active) {
        setStatus("No active upload to abort.");
        return;
      }
      abortButton.disabled = true;
      abortMultipart(active.multipartId)
        .then(function() {
          clearState(active.signature);
          setStatus("Upload aborted.");
          setProgress(0);
          active = null;
        })
        .catch(function() {
          setStatus("Abort failed. Try again.");
        })
        .finally(function() {
          abortButton.disabled = false;
        });
    });
  }
})();
