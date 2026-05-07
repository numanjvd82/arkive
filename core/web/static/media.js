(function() {
  const actionsPanel = document.querySelector("[data-media-file-id]");
  const stage = document.querySelector("[data-media-stage='true']");
  const status = document.querySelector("[data-media-status='true'] span");
  const title = document.querySelector("[data-media-title='true']");
  const typeChip = document.querySelector("[data-media-chip-type='true']");
  const hashValue = document.querySelector("[data-media-hash='true']");
  const hashNote = document.querySelector("[data-media-hash-note='true']");
  const mimeField = document.querySelector("[data-media-field='media-mime']");
  const sizeField = document.querySelector("[data-media-field='media-size']");
  const dimensionsField = document.querySelector("[data-media-field='media-dimensions']");
  const downloadButton = document.getElementById("media-download-button");
  const shareButton = document.getElementById("media-share-button");
  const deleteButton = document.getElementById("media-delete-button");

  if (!actionsPanel || !stage || !window.ArkiveFileReader) {
    return;
  }

  const fileId = actionsPanel.getAttribute("data-media-file-id");
  const reader = new window.ArkiveFileReader({ fileId: fileId });
  const SMALL_VIDEO_MAX_BYTES = 128 * 1024 * 1024;
  const TEXT_PREVIEW_MAX_BYTES = 2 * 1024 * 1024;

  function setStatus(message) {
    if (status) {
      status.textContent = message || "";
      status.parentElement.classList.toggle("is-hidden", !message);
    }
  }

  function setStage(node) {
    stage.innerHTML = "";
    if (node) {
      stage.appendChild(node);
    }
  }

  function formatBytes(bytes) {
    const value = Number(bytes || 0);
    if (value <= 0) {
      return "0 B";
    }
    const units = ["B", "KB", "MB", "GB", "TB"];
    const index = Math.min(
      Math.floor(Math.log(value) / Math.log(1024)),
      units.length - 1,
    );
    const sized = value / Math.pow(1024, index);
    return sized.toFixed(index === 0 ? 0 : 1) + " " + units[index];
  }

  function base64ToHex(value) {
    const normalized = String(value || "").trim();
    if (!normalized) {
      return "";
    }
    try {
      const binary = atob(normalized);
      let hex = "";
      for (let i = 0; i < binary.length; i++) {
        hex += binary.charCodeAt(i).toString(16).padStart(2, "0");
      }
      return hex;
    } catch (_) {
      return normalized;
    }
  }

  function copyText(value) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      return navigator.clipboard.writeText(value);
    }
    return Promise.reject(new Error("Clipboard unavailable"));
  }

  function previewUnavailable(message) {
    const shell = document.createElement("div");
    shell.className = "media-placeholder";
    const label = document.createElement("span");
    label.textContent = "Preview unavailable";
    const note = document.createElement("p");
    note.textContent = message || "Download the file to inspect it locally.";
    shell.appendChild(label);
    shell.appendChild(note);
    setStage(shell);
  }

  function setDimensions(width, height) {
    if (!dimensionsField) {
      return;
    }
    dimensionsField.textContent =
      width > 0 && height > 0 ? width + "×" + height : "Not available";
  }

  function imagePreview(blob, alt, titleText) {
    const img = document.createElement("img");
    img.className = "media-image";
    img.alt = alt || "Image preview";
    img.src = URL.createObjectURL(blob);
    img.setAttribute("data-lightbox-trigger", "true");
    img.setAttribute("data-lightbox-src", img.src);
    img.setAttribute("data-lightbox-title", titleText || alt || "Image preview");
    img.addEventListener("load", function() {
      setDimensions(img.naturalWidth || 0, img.naturalHeight || 0);
    });
    const wrap = document.createElement("div");
    wrap.className = "media-image-wrap";
    wrap.appendChild(img);
    const button = document.createElement("button");
    button.className = "media-fullscreen-button";
    button.type = "button";
    button.setAttribute("aria-label", "Open full screen");
    button.setAttribute("data-lightbox-src", img.src);
    button.setAttribute("data-lightbox-title", titleText || alt || "Image preview");
    button.innerHTML =
      '<svg class="media-fullscreen-lucide" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M8 3H5a2 2 0 0 0-2 2v3"/><path d="M16 3h3a2 2 0 0 1 2 2v3"/><path d="M8 21H5a2 2 0 0 1-2-2v-3"/><path d="M16 21h3a2 2 0 0 0 2-2v-3"/></svg>';
    wrap.appendChild(button);
    setStage(wrap);
    stage.firstElementChild.classList.add("is-image");
  }

  function videoPreview(blob) {
    const video = document.createElement("video");
    video.className = "media-video plyr";
    video.controls = true;
    video.playsInline = true;
    video.setAttribute("data-video-element", "true");
    video.src = URL.createObjectURL(blob);
    video.addEventListener("loadedmetadata", function() {
      setDimensions(video.videoWidth || 0, video.videoHeight || 0);
    });
    setStage(video);
    if (window.ArkiveInitPlyr) {
      window.ArkiveInitPlyr(video);
    }
  }

  function textPreview(text) {
    const pre = document.createElement("pre");
    pre.className = "media-text-preview";
    pre.textContent = text;
    setStage(pre);
  }

  function updateMetadata(metadata, manifest, record) {
    if (title) {
      title.textContent = metadata.name || "Encrypted file";
    }
    if (typeChip) {
      typeChip.textContent = String(metadata.mime || "unknown").toUpperCase();
    }
    if (mimeField) {
      mimeField.textContent = metadata.mime || "Unknown";
    }
    if (sizeField) {
      sizeField.textContent = formatBytes(metadata.size || record.plaintextSize);
    }
    const preview = metadata.preview || {};
    setDimensions(preview.width || 0, preview.height || 0);
    if (hashValue) {
      hashValue.textContent = base64ToHex(record.encryptedHash) || "Unavailable";
    }
    if (hashNote) {
      hashNote.textContent = "BLAKE3 over encrypted object bytes.";
    }
    actionsPanel.setAttribute("data-media-file-name", metadata.name || "File");
  }

  function isTextMime(mime) {
    const value = String(mime || "").toLowerCase();
    return (
      value.startsWith("text/") ||
      value.includes("json") ||
      value.includes("javascript") ||
      value.includes("xml")
    );
  }

  async function renderPreview() {
    await reader.load();
    const metadata = reader.getMetadata();
    const manifest = reader.getManifest();
    const record = reader.record;
    const mime = String(metadata.mime || "").toLowerCase();

    updateMetadata(metadata, manifest, record);

    if (mime.startsWith("image/")) {
      imagePreview(await reader.createBlob(), metadata.name, metadata.name);
      setStatus("");
      return;
    }
    if (mime === "application/pdf") {
      previewUnavailable("PDF preview is not enabled yet. Download is available now.");
      setStatus("PDF preview unavailable.");
      return;
    }
    if (isTextMime(mime)) {
      setStatus("Decrypting text preview.");
      textPreview(await reader.textPreview(TEXT_PREVIEW_MAX_BYTES));
      setStatus("");
      return;
    }
    if (mime.startsWith("video/")) {
      if (Number(metadata.size || record.plaintextSize || 0) > SMALL_VIDEO_MAX_BYTES) {
        previewUnavailable("Encrypted large-video range playback is next step. Download is available now.");
        setStatus("Large encrypted video. Download recommended.");
        return;
      }
      videoPreview(await reader.createBlob());
      setStatus("");
      return;
    }

    previewUnavailable("This file type does not have an in-browser preview yet.");
    setStatus("Preview unavailable.");
  }

  function fetchExistingShare(fileID) {
    return fetch("/api/files/" + encodeURIComponent(fileID) + "/share", {
      method: "GET",
      headers: { "Content-Type": "application/json" }
    }).then(function(res) {
      if (!res.ok) {
        throw res;
      }
      return res.json();
    });
  }

  function createShare(fileID) {
    return fetch("/api/files/" + encodeURIComponent(fileID) + "/share", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({})
    }).then(function(res) {
      if (!res.ok) {
        return res.json().then(function(data) {
          throw new Error((data && data.error) || "Share failed");
        });
      }
      return res.json();
    });
  }

  if (downloadButton) {
    downloadButton.addEventListener("click", async function(event) {
      event.preventDefault();
      downloadButton.disabled = true;
      try {
        await reader.download();
      } catch (error) {
        if (window.Toast) {
          window.Toast.error((error && error.message) || "Download failed.");
        }
      } finally {
        downloadButton.disabled = false;
      }
    });
  }

  if (shareButton) {
    shareButton.addEventListener("click", function() {
      shareButton.disabled = true;
      fetchExistingShare(fileId)
        .catch(function(res) {
          if (res && res.status === 404) {
            return createShare(fileId);
          }
          throw new Error("Share failed");
        })
        .then(function(data) {
          const token = data && data.token;
          if (!token) {
            throw new Error("Share failed");
          }
          const url = window.location.origin + "/s/" + token;
          return copyText(url).then(function() {
            if (window.Toast) {
              window.Toast.success("Share link copied.", { title: "Shared" });
            }
          });
        })
        .catch(function(error) {
          if (window.Toast) {
            window.Toast.error((error && error.message) || "Share failed.");
          }
        })
        .finally(function() {
          shareButton.disabled = false;
        });
    });
  }

  if (deleteButton) {
    deleteButton.addEventListener("click", function() {
      const confirmed = window.confirm("Delete this file? This action cannot be undone.");
      if (!confirmed) {
        return;
      }
      deleteButton.disabled = true;
      fetch("/api/files/" + encodeURIComponent(fileId), {
        method: "DELETE",
        headers: { "Content-Type": "application/json" }
      })
        .then(function(res) {
          if (!res.ok) {
            throw new Error("Delete failed");
          }
          window.location.href = "/files";
        })
        .catch(function() {
          deleteButton.disabled = false;
          if (window.Toast) {
            window.Toast.error("Delete failed. Try again.");
          }
        });
    });
  }

  renderPreview().catch(function(error) {
    previewUnavailable("Download is available while preview pipeline finishes loading.");
    setStatus((error && error.message) || "Preview failed.");
    if (window.Toast) {
      window.Toast.error((error && error.message) || "Preview failed.");
    }
  });

  window.addEventListener("pagehide", function() {
    reader.dispose();
  });
})();
