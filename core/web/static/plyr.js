(function() {
  function getPlaybackContext() {
    const nav = typeof navigator !== "undefined" ? navigator : null;
    const ua = nav && nav.userAgent ? nav.userAgent : "";
    const platform = nav && nav.platform ? nav.platform : "";
    const maxTouchPoints = nav && nav.maxTouchPoints ? nav.maxTouchPoints : 0;

    const isIOS =
      /iPad|iPhone|iPod/.test(ua) ||
      (platform === "MacIntel" && maxTouchPoints > 1);

    const isSafari =
      /^((?!chrome|android|crios|fxios|edg|opr).)*safari/i.test(ua);

    const isAndroid = /android/i.test(ua);
    const isMobile = isIOS || isAndroid;

    return {
      isMobile: isMobile,
      isIOSSafari: isIOS && isSafari,
    };
  }

  function enterNativeFullscreen(videoEl, player) {
    if (!videoEl) {
      return;
    }

    const open = function() {
      if (typeof videoEl.webkitEnterFullscreen === "function") {
        videoEl.webkitEnterFullscreen();
        return true;
      }
      if (typeof videoEl.requestFullscreen === "function") {
        videoEl.requestFullscreen().catch(function() {});
        return true;
      }
      if (typeof videoEl.webkitSetPresentationMode === "function") {
        try {
          videoEl.webkitSetPresentationMode("fullscreen");
          return true;
        } catch (_) {}
      }
      if (player && player.fullscreen && typeof player.fullscreen.enter === "function") {
        player.fullscreen.enter();
        return true;
      }
      return false;
    };

    if (videoEl.paused && typeof videoEl.play === "function") {
      videoEl.play().then(function() {
        open();
      }).catch(function() {
        open();
      });
      return;
    }

    open();
  }

  function initPlyr(videoEl) {
    if (!videoEl || !window.Plyr) {
      return null;
    }
    if (videoEl.__arkivePlyr) {
      return videoEl.__arkivePlyr;
    }

    const playback = getPlaybackContext();
    const controls = [
      "play",
      "progress",
      "current-time",
      "duration",
      "mute",
      "volume",
      "settings",
      "pip",
    ];
    if (!playback.isMobile) {
      controls.push("fullscreen");
    }

    const player = new window.Plyr(videoEl, {
      controls: controls,
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
    const controlsEl = player.elements && player.elements.controls ? player.elements.controls : null;
    let cinemaToggle = null;
    if (controlsEl && playback.isMobile) {
      const fullscreenToggle = document.createElement("button");
      fullscreenToggle.type = "button";
      fullscreenToggle.className = "plyr__control";
      fullscreenToggle.setAttribute("aria-label", "Full screen");
      fullscreenToggle.innerHTML =
        '<svg aria-hidden="true" focusable="false" viewBox="0 0 24 24"><path fill="currentColor" d="M7 3H5a2 2 0 0 0-2 2v2h2V5h2V3zm12 0h-2v2h2v2h2V5a2 2 0 0 0-2-2zM3 17v2a2 2 0 0 0 2 2h2v-2H5v-2H3zm16 2h-2v2h2a2 2 0 0 0 2-2v-2h-2v2z"></path></svg>';
      fullscreenToggle.addEventListener("click", function(event) {
        event.preventDefault();
        enterNativeFullscreen(videoEl, player);
      });
      controlsEl.appendChild(fullscreenToggle);
    }
    if (controlsEl) {
      cinemaToggle = document.createElement("button");
      cinemaToggle.type = "button";
      cinemaToggle.className = "plyr__control plyr__control--cinema";
      cinemaToggle.textContent = "Cinema";
      cinemaToggle.setAttribute("aria-label", "Cinema view");
      controlsEl.appendChild(cinemaToggle);
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
