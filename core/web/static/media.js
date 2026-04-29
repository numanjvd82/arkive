(function() {
  const downloadButton = document.querySelector("[data-download-id]");
  const playButton = document.querySelector("[data-video-action='play']");
  const videoEl = document.querySelector("[data-video-element='true']");

  function bindDownload() {
    if (!downloadButton) {
      return;
    }

    downloadButton.addEventListener("click", function() {
      const fileId = downloadButton.getAttribute("data-download-id");
      if (!fileId) {
        return;
      }
      window.open("/api/files/" + encodeURIComponent(fileId) + "/download", "_blank", "noopener");
    });
  }

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

  if (!downloadButton && !playButton) {
    return;
  }

  bindDownload();
  bindPlay();
})();
