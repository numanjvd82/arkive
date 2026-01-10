(() => {
  window.__arkiveMonetagVignetteLoaded = true;

  const target = [document.documentElement, document.body].filter(Boolean).pop();
  if (!target) {
    return;
  }

  const script = document.createElement("script");
  script.onload = () => {
    window.__arkiveMonetagVignetteExternalLoaded = true;
  };
  script.onerror = () => {
    window.__arkiveMonetagVignetteExternalBlocked = true;
  };
  script.dataset.zone = "10431810";
  script.src = "https://gizokraijaw.net/vignette.min.js";
  target.appendChild(script);
})();
