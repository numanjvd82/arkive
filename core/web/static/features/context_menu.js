export function initContextMenu() {
  document.addEventListener("contextmenu", function(event) {
    const entry = event.target && event.target.closest ? event.target.closest("[data-selectable-entry]") : null;
    if (!entry) {
      return;
    }
    if (entry.classList.contains("files-card")) {
      event.preventDefault();
      entry.click();
    }
  });
}
