(function() {
  const playButton = document.querySelector("[data-video-action='play']");
  const videoEl = document.querySelector("[data-video-element='true']");
  const actionsPanel = document.querySelector("[data-media-file-id]");
  const shareButton = document.getElementById("media-share-button");
  const deleteButton = document.getElementById("media-delete-button");

  function bindPlay() {
    if (!playButton || !videoEl) {
      return;
    }
    if (window.Plyr) {
      return;
    }
    playButton.addEventListener("click", function() {
      const src = videoEl.getAttribute("data-video-src");
      if (!src) {
        return;
      }
      if (!videoEl.getAttribute("src")) {
        videoEl.setAttribute("src", src);
      }
      videoEl.play().catch(function() {});
      playButton.disabled = true;
    });
  }

  function copyText(value) {
    if (!value) {
      return Promise.reject(new Error("Missing value"));
    }
    if (navigator.clipboard && navigator.clipboard.writeText) {
      return navigator.clipboard.writeText(value);
    }
    const input = document.createElement("textarea");
    input.value = value;
    input.setAttribute("readonly", "readonly");
    input.style.position = "absolute";
    input.style.left = "-9999px";
    document.body.appendChild(input);
    input.select();
    document.execCommand("copy");
    document.body.removeChild(input);
    return Promise.resolve();
  }

  function fetchExistingShare(fileId) {
    return fetch("/api/files/" + encodeURIComponent(fileId) + "/share", {
      method: "GET",
      headers: { "Content-Type": "application/json" }
    }).then(function(res) {
      if (!res.ok) {
        throw res;
      }
      return res.json();
    });
  }

  function createShare(fileId) {
    return fetch("/api/files/" + encodeURIComponent(fileId) + "/share", {
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

  function bindShare() {
    if (!shareButton || !actionsPanel) {
      return;
    }
    shareButton.addEventListener("click", function() {
      const fileId = actionsPanel.getAttribute("data-media-file-id");
      const filename = actionsPanel.getAttribute("data-media-file-name") || "File";
      if (!fileId) {
        return;
      }
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
              window.Toast.success("Share link copied for " + filename + ".", { title: "Shared" });
            }
          });
        })
        .catch(function(err) {
          if (window.Toast) {
            window.Toast.error((err && err.message) || "Share failed. Try again.");
          }
        })
        .finally(function() {
          shareButton.disabled = false;
        });
    });
  }

  function bindDelete() {
    if (!deleteButton || !actionsPanel) {
      return;
    }
    deleteButton.addEventListener("click", function() {
      const fileId = actionsPanel.getAttribute("data-media-file-id");
      const filename = actionsPanel.getAttribute("data-media-file-name") || "this file";
      if (!fileId) {
        return;
      }
      const confirmed = window.confirm("Delete " + filename + "? This action cannot be undone.");
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
          if (window.Toast) {
            window.Toast.success("File deleted.", { title: "Deleted" });
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

  bindPlay();
  bindShare();
  bindDelete();
})();
