(() => {
  window.__arkiveMonetagPushLoaded = true;

  const script = document.createElement("script");
  script.onload = () => {
    window.__arkiveMonetagPushExternalLoaded = true;
  };
  script.onerror = () => {
    window.__arkiveMonetagPushExternalBlocked = true;
  };
  script.src = "https://3nbf4.com/act/files/tag.min.js?z=10414558";
  script.async = true;
  script.setAttribute("data-cfasync", "false");

  const head = document.head || document.documentElement;
  if (head) {
    head.appendChild(script);
  }
})();
