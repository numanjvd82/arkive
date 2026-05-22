let initialized = false;
let layer = null;
let activeTrigger = null;
const MARGIN = 12;
const OFFSET = 14;

function ensureLayer() {
  if (layer) {
    return layer;
  }
  layer = document.createElement("div");
  layer.className = "tooltip-layer";
  layer.setAttribute("role", "tooltip");
  layer.setAttribute("aria-hidden", "true");
  document.body.appendChild(layer);
  return layer;
}

function tooltipText(trigger) {
  return String(trigger.getAttribute("data-tooltip") || "").trim();
}

function hide(trigger) {
  if (trigger && activeTrigger && trigger !== activeTrigger) {
    return;
  }
  if (activeTrigger) {
    activeTrigger.setAttribute("aria-expanded", "false");
  }
  activeTrigger = null;
  if (!layer) {
    return;
  }
  layer.classList.remove("is-visible");
  layer.setAttribute("aria-hidden", "true");
}

function position(trigger) {
  if (!layer || !trigger) {
    return;
  }

  const rect = trigger.getBoundingClientRect();
  const viewportWidth = window.innerWidth || document.documentElement.clientWidth || 0;
  const viewportHeight = window.innerHeight || document.documentElement.clientHeight || 0;
  const bubbleRect = layer.getBoundingClientRect();

  const topSpace = rect.top - MARGIN;
  const bottomSpace = viewportHeight - rect.bottom - MARGIN;
  const placement = topSpace >= bubbleRect.height + OFFSET || topSpace >= bottomSpace ? "top" : "bottom";

  layer.setAttribute("data-placement", placement);

  let left = rect.left + (rect.width / 2) - (bubbleRect.width / 2);
  left = Math.max(MARGIN, Math.min(left, viewportWidth - bubbleRect.width - MARGIN));

  let top = placement === "top"
    ? rect.top - bubbleRect.height - OFFSET
    : rect.bottom + OFFSET;
  top = Math.max(MARGIN, Math.min(top, viewportHeight - bubbleRect.height - MARGIN));

  const arrow = rect.left + (rect.width / 2) - left;
  const arrowLeft = Math.max(18, Math.min(arrow, bubbleRect.width - 18));

  layer.style.left = Math.round(left) + "px";
  layer.style.top = Math.round(top) + "px";
  layer.style.setProperty("--tooltip-arrow-left", Math.round(arrowLeft) + "px");
}

function show(trigger) {
  const text = tooltipText(trigger);
  if (!text) {
    hide();
    return;
  }

  const node = ensureLayer();
  activeTrigger = trigger;
  trigger.setAttribute("aria-expanded", "true");
  node.textContent = text;
  node.setAttribute("aria-hidden", "false");
  node.classList.add("is-visible");
  position(trigger);
}

function bind(trigger) {
  if (!trigger || trigger.__arkiveTooltipBound) {
    return;
  }
  trigger.__arkiveTooltipBound = true;
  trigger.setAttribute("aria-expanded", "false");
  trigger.addEventListener("mouseenter", function() { show(trigger); });
  trigger.addEventListener("mouseleave", function() { hide(trigger); });
  trigger.addEventListener("focus", function() { show(trigger); });
  trigger.addEventListener("blur", function() { hide(trigger); });
  trigger.addEventListener("click", function(event) {
    event.preventDefault();
    event.stopPropagation();
    if (activeTrigger === trigger) {
      hide(trigger);
      return;
    }
    show(trigger);
  });
}

function bindAll() {
  const triggers = document.querySelectorAll(".tooltip-icon[data-tooltip]");
  for (let i = 0; i < triggers.length; i += 1) {
    bind(triggers[i]);
  }
}

export function initTooltips() {
  if (typeof window === "undefined" || typeof document === "undefined") {
    return;
  }
  if (initialized) {
    bindAll();
    return;
  }
  initialized = true;

  document.addEventListener("scroll", function() {
    if (activeTrigger) {
      position(activeTrigger);
    }
  }, true);

  window.addEventListener("resize", function() {
    if (activeTrigger) {
      position(activeTrigger);
    }
  });

  document.addEventListener("keydown", function(event) {
    if (event.key === "Escape") {
      hide();
    }
  });

  document.addEventListener("click", function(event) {
    if (!activeTrigger) {
      return;
    }
    const target = event.target;
    if (target && activeTrigger.contains(target)) {
      return;
    }
    hide();
  });
  bindAll();
}
