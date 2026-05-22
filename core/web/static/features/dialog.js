let initialized = false;

function getBackdrop(id) {
  if (!id) {
    return null;
  }
  return document.getElementById(id);
}

function open(id) {
  const backdrop = getBackdrop(id);
  if (!backdrop) {
    return;
  }
  backdrop.classList.remove("is-hidden");
}

function close(id) {
  const backdrop = getBackdrop(id);
  if (!backdrop) {
    return;
  }
  backdrop.classList.add("is-hidden");
}

function closeTopmost() {
  const openBackdrops = Array.from(document.querySelectorAll(".dialog-backdrop:not(.is-hidden)"));
  if (!openBackdrops.length) {
    return;
  }
  openBackdrops[openBackdrops.length - 1].classList.add("is-hidden");
}

export const Dialog = {
  open: open,
  close: close,
};

export function initDialogs() {
  if (initialized) {
    return;
  }
  initialized = true;

  document.addEventListener("click", function(event) {
    const backdrop = event.target && event.target.classList && event.target.classList.contains("dialog-backdrop")
      ? event.target
      : null;
    if (backdrop) {
      backdrop.classList.add("is-hidden");
    }
  });

  document.addEventListener("keydown", function(event) {
    if (event.key === "Escape") {
      closeTopmost();
    }
  });
}
