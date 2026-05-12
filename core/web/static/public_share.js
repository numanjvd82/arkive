(function() {
  const root = document.querySelector("[data-public-share-token]");
  const preview = document.getElementById("public-share-preview");
  const download = document.getElementById("public-share-download");
  const nameEl = document.querySelector("[data-public-share-name='true']");
  const sizeEl = document.querySelector("[data-public-share-size='true']");
  const SMALL_VIDEO_MAX_BYTES = 128 * 1024 * 1024;
  const TEXT_PREVIEW_MAX_BYTES = 2 * 1024 * 1024;
  let currentPreviewURL = "";

  if (!root || !window.ArkiveShareReader) {
    return;
  }

  function revokePreviewURL() {
    if (!currentPreviewURL) {
      return;
    }
    URL.revokeObjectURL(currentPreviewURL);
    currentPreviewURL = "";
  }

  function shareSecretFromHash() {
    const hash = String(window.location.hash || "").replace(/^#/, "");
    if (!hash) {
      return "";
    }
    const params = new URLSearchParams(hash);
    return String(params.get("s") || params.get("share-secret") || params.get("key") || "");
  }

  function setPreview(node) {
    if (!preview) {
      return;
    }
    revokePreviewURL();
    preview.innerHTML = "";
    if (node) {
      preview.appendChild(node);
    }
  }

  function unavailable(message) {
    const node = document.createElement("div");
    node.className = "public-share-empty";
    node.innerHTML = "<span>Preview unavailable</span><p>" + message + "</p>";
    setPreview(node);
  }

  function updateMetadata(metadata, sizeBytes) {
    if (nameEl) {
      nameEl.textContent = String((metadata && metadata.name) || "Encrypted file");
    }
    if (sizeEl && typeof sizeBytes === "number" && sizeBytes > 0) {
      sizeEl.textContent = sizeBytes >= 1024 * 1024
        ? (sizeBytes / 1024 / 1024).toFixed(2) + " MB"
        : Math.max(1, Math.round(sizeBytes / 1024)) + " KB";
    }
  }

  function imagePreview(blob, titleText) {
    const objectURL = URL.createObjectURL(blob);
    const wrap = document.createElement("div");
    wrap.className = "public-share-image-wrap";
    const img = document.createElement("img");
    img.className = "public-share-image";
    img.src = objectURL;
    img.alt = titleText || "Shared image";
    img.setAttribute("data-lightbox-trigger", "true");
    img.setAttribute("data-lightbox-src", objectURL);
    img.setAttribute("data-lightbox-title", titleText || "Shared image");
    wrap.appendChild(img);
    const button = document.createElement("button");
    button.className = "media-fullscreen-button";
    button.type = "button";
    button.setAttribute("aria-label", "Open full screen");
    button.setAttribute("data-lightbox-src", objectURL);
    button.setAttribute("data-lightbox-title", titleText || "Shared image");
    button.innerHTML = '<svg class="media-fullscreen-lucide" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M8 3H5a2 2 0 0 0-2 2v3"/><path d="M16 3h3a2 2 0 0 1 2 2v3"/><path d="M8 21H5a2 2 0 0 1-2-2v-3"/><path d="M16 21h3a2 2 0 0 0 2-2v-3"/></svg>';
    wrap.appendChild(button);
    currentPreviewURL = objectURL;
    setPreview(wrap);
  }

  function videoPreview(blob) {
    const objectURL = URL.createObjectURL(blob);
    const video = document.createElement("video");
    video.className = "public-share-video plyr";
    video.controls = true;
    video.playsInline = true;
    video.setAttribute("data-video-element", "true");
    video.src = objectURL;
    currentPreviewURL = objectURL;
    setPreview(video);
    if (window.ArkiveInitPlyr) {
      window.ArkiveInitPlyr(video);
    }
  }

  function textPreview(text) {
    const node = document.createElement("pre");
    node.className = "public-share-text";
    node.textContent = text;
    setPreview(node);
  }

  const token = String(root.getAttribute("data-public-share-token") || "");
  const shareSecret = shareSecretFromHash();
  if (!shareSecret) {
    unavailable("This link is missing its secret fragment.");
    if (download) {
      download.addEventListener("click", function(event) {
        event.preventDefault();
      });
    }
    return;
  }

  const reader = new window.ArkiveShareReader({
    token: token,
    shareSecret: shareSecret,
  });

  if (download) {
    download.addEventListener("click", function(event) {
      event.preventDefault();
      reader.download().catch(function() {});
    });
  }

  reader.load()
    .then(function() {
      const metadata = reader.getMetadata();
      updateMetadata(metadata, Number(metadata.size || reader.record.plaintextSize || 0));
      const mime = String((metadata && metadata.mime) || "").toLowerCase();
      if (mime.indexOf("image/") === 0) {
        return reader.createBlob().then(function(blob) {
          imagePreview(blob, metadata.name || "Shared image");
        });
      }
      if (mime.indexOf("video/") === 0) {
        if (Number(metadata.size || reader.record.plaintextSize || 0) > SMALL_VIDEO_MAX_BYTES) {
          unavailable("Large encrypted video preview is not enabled yet. Download is available.");
          return;
        }
        return reader.createBlob().then(function(blob) {
          videoPreview(blob);
        });
      }
      if (mime.indexOf("text/") === 0 || mime === "application/json") {
        return reader.textPreview(TEXT_PREVIEW_MAX_BYTES).then(function(text) {
          textPreview(text);
        });
      }
      unavailable("Download the file to view it locally.");
    })
    .catch(function(error) {
      unavailable((error && error.message) || "Failed to load encrypted share.");
    });

  window.addEventListener("pagehide", function() {
    revokePreviewURL();
    reader.dispose();
  });
})();
