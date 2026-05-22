(function() {
  const form = document.querySelector(".share-form");
  if (!form) {
    return;
  }

  const originalAction = String(form.getAttribute("action") || window.location.pathname).split("#")[0];

  function applyHashToAction() {
    const hash = String(window.location.hash || "");
    form.setAttribute("action", originalAction + hash);
  }

  applyHashToAction();
  form.addEventListener("submit", applyHashToAction);
})();
