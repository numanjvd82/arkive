import { entrySelection } from "./file_selection.js";

function rectsIntersect(a, b) {
  return !(a.right < b.left || a.left > b.right || a.bottom < b.top || a.top > b.bottom);
}

export function initDragSelect() {
  const root = document.querySelector("[data-drag-select-root='grid']");
  const grid = root && root.querySelector ? root.querySelector(".files-grid") : null;
  if (!root || !grid || root.hasAttribute("data-drag-select-bound")) {
    return;
  }
  root.setAttribute("data-drag-select-bound", "true");

  let dragBox = null;
  let startX = 0;
  let startY = 0;
  let active = false;

  function cleanup() {
    active = false;
    if (dragBox && dragBox.parentNode) {
      dragBox.parentNode.removeChild(dragBox);
    }
    dragBox = null;
    document.body.classList.remove("is-drag-selecting");
    window.removeEventListener("pointermove", onPointerMove);
    window.removeEventListener("pointerup", onPointerUp);
  }

  function updateSelection(currentX, currentY) {
    if (!dragBox) {
      return;
    }
    const left = Math.min(startX, currentX);
    const top = Math.min(startY, currentY);
    const width = Math.abs(currentX - startX);
    const height = Math.abs(currentY - startY);
    dragBox.style.left = left + "px";
    dragBox.style.top = top + "px";
    dragBox.style.width = width + "px";
    dragBox.style.height = height + "px";

    const selectionRect = {
      left: left,
      top: top,
      right: left + width,
      bottom: top + height,
    };

    const matches = Array.from(grid.querySelectorAll("[data-selectable-entry]")).filter(function(entry) {
      return rectsIntersect(selectionRect, entry.getBoundingClientRect());
    });
    entrySelection.replaceSelection(matches);
  }

  function onPointerMove(event) {
    if (!active) {
      return;
    }
    updateSelection(event.clientX, event.clientY);
  }

  function onPointerUp() {
    cleanup();
  }

  root.addEventListener("pointerdown", function(event) {
    if (event.button !== 0) {
      return;
    }
    if (event.target.closest("[data-selectable-entry], button, a, input, label")) {
      return;
    }
    active = true;
    startX = event.clientX;
    startY = event.clientY;
    dragBox = document.createElement("div");
    dragBox.className = "files-drag-select-box";
    document.body.appendChild(dragBox);
    document.body.classList.add("is-drag-selecting");
    window.addEventListener("pointermove", onPointerMove);
    window.addEventListener("pointerup", onPointerUp);
    updateSelection(startX, startY);
  });
}
