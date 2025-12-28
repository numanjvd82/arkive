(function() {
  const downloadButton = document.querySelector("[data-download-id]");
  if (!downloadButton) {
    return;
  }

  downloadButton.addEventListener("click", function() {
    const fileId = downloadButton.getAttribute("data-download-id");
    if (!fileId) {
      return;
    }
    const popup = window.open("", "_blank", "noopener");
    downloadButton.disabled = true;
    fetch("/api/files/" + encodeURIComponent(fileId) + "/download", {
      method: "GET",
      headers: { "Content-Type": "application/json" }
    })
      .then(function(res) {
        if (!res.ok) {
          throw new Error("Download failed");
        }
        return res.json();
      })
      .then(function(payload) {
        if (!payload || !payload.url) {
          throw new Error("Download failed");
        }
        if (popup && !popup.closed) {
          popup.location.href = payload.url;
        } else {
          window.location.href = payload.url;
        }
      })
      .catch(function() {
        window.Toast.error("Download failed. Try again.");
      })
      .finally(function() {
        downloadButton.disabled = false;
      });
  });
})();
