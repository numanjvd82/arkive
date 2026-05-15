function ensureLabel(button) {
  if (!button) {
    return null;
  }
  const existing = button.querySelector(".button-label");
  if (existing) {
    return existing;
  }
  for (let i = button.childNodes.length - 1; i >= 0; i--) {
    const node = button.childNodes[i];
    if (!node) {
      continue;
    }
    if (node.nodeType === Node.TEXT_NODE && String(node.textContent || "").trim()) {
      const label = document.createElement("span");
      label.className = "button-label";
      label.textContent = String(node.textContent || "").trim();
      button.replaceChild(label, node);
      return label;
    }
    if (node.nodeType === Node.ELEMENT_NODE && node.tagName === "SPAN" && !node.classList.contains("button-icon") && !node.classList.contains("button-spinner")) {
      node.classList.add("button-label");
      return node;
    }
  }
  return null;
}

function ensureSpinner(button) {
  let spinner = button.querySelector(".button-spinner");
  if (spinner) {
    return spinner;
  }
  spinner = document.createElement("span");
  spinner.className = "button-spinner";
  spinner.setAttribute("aria-hidden", "true");
  const icon = button.querySelector(".button-icon, .button-lucide");
  if (icon && icon.parentNode === button) {
    button.insertBefore(spinner, icon);
    return spinner;
  }
  button.insertBefore(spinner, button.firstChild);
  return spinner;
}

export function setButtonBusy(button, busy, options = {}) {
  if (!button) {
    return;
  }
  const restoreDisabled = options.restoreDisabled !== false;
  const busyText = String(options.busyText || "");
  const label = ensureLabel(button);

  if (busy) {
    button.setAttribute("data-busy-original-disabled", button.disabled ? "true" : "false");
    if (label && busyText && !button.hasAttribute("data-busy-original-label")) {
      button.setAttribute("data-busy-original-label", label.textContent || "");
      label.textContent = busyText;
    }
    ensureSpinner(button);
    button.disabled = true;
    button.classList.add("is-busy");
    button.setAttribute("aria-busy", "true");
    return;
  }

  const spinner = button.querySelector(".button-spinner");
  if (spinner && spinner.parentNode) {
    spinner.parentNode.removeChild(spinner);
  }
  if (label && button.hasAttribute("data-busy-original-label")) {
    label.textContent = button.getAttribute("data-busy-original-label") || "";
    button.removeAttribute("data-busy-original-label");
  }
  if (restoreDisabled) {
    button.disabled = button.getAttribute("data-busy-original-disabled") === "true";
  }
  button.removeAttribute("data-busy-original-disabled");
  button.classList.remove("is-busy");
  button.removeAttribute("aria-busy");
}
