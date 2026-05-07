(function() {
  function initPlyr(videoEl) {
    if (!videoEl || !window.Plyr) {
      return null;
    }
    if (videoEl.__arkivePlyr) {
      return videoEl.__arkivePlyr;
    }

    const player = new window.Plyr(videoEl, {
      controls: [
        "play",
        "progress",
        "current-time",
        "duration",
        "mute",
        "volume",
        "settings",
        "pip",
        "fullscreen"
      ]
    });

    function ensureSource(src) {
      if (!src) {
        return;
      }
      if (player.source && player.source.sources && player.source.sources.length) {
        return;
      }
      player.source = {
        type: "video",
        sources: [{ src: src }]
      };
    }

    const playButton = document.querySelector("[data-video-action='play']");
    if (playButton && !playButton.__arkivePlyrBound) {
      playButton.__arkivePlyrBound = true;
      playButton.addEventListener("click", function() {
        const src = videoEl.getAttribute("data-video-src");
        ensureSource(src);
        player.play().catch(function() {});
        playButton.disabled = true;
      });
    }

    const mediaView = document.querySelector(".media-view, .share-view");
    const controls = player.elements && player.elements.controls ? player.elements.controls : null;
    let cinemaToggle = null;
    if (controls) {
      cinemaToggle = document.createElement("button");
      cinemaToggle.type = "button";
      cinemaToggle.className = "plyr__control plyr__control--cinema";
      cinemaToggle.textContent = "Cinema";
      cinemaToggle.setAttribute("aria-label", "Cinema view");
      controls.appendChild(cinemaToggle);
    }
    if (cinemaToggle && mediaView) {
      cinemaToggle.addEventListener("click", function() {
        const isCinema = mediaView.classList.toggle("is-cinema");
        cinemaToggle.textContent = isCinema ? "Exit cinema" : "Cinema";
      });
    }

    videoEl.__arkivePlyr = player;
    return player;
  }

  window.ArkiveInitPlyr = initPlyr;

  const videoEl = document.querySelector("[data-video-element='true']");
  if (videoEl) {
    initPlyr(videoEl);
  }
})();
