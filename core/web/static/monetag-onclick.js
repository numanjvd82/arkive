(() => {
  const target = [document.documentElement, document.body].filter(Boolean).pop();
  if (!target) {
    return;
  }

  const script = document.createElement("script");
  script.dataset.zone = "10431804";
  script.src = "https://al5sm.com/tag.min.js";
  target.appendChild(script);
})();
