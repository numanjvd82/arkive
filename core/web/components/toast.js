const toastHost = document.getElementById("toast-host");
if (!toastHost) {
  return;
}

const defaultDuration = 4200;
const defaultTitles = {
  info: "Heads up",
  success: "Done",
  warning: "Check this",
  error: "Something went wrong"
};
const iconBank = {};

document.querySelectorAll("[data-toast-icon]").forEach((node) => {
  const key = node.getAttribute("data-toast-icon");
  const svg = node.querySelector("svg");
  if (key && svg) {
    iconBank[key] = svg;
  }
});

function cloneIcon(name) {
  const svg = iconBank[name];
  if (!svg) {
    return null;
  }
  return svg.cloneNode(true);
}

function buildToast(message, options) {
  const opts = options || {};
  const type = opts.type || "info";
  const title = opts.title || defaultTitles[type] || "Notice";
  const duration = typeof opts.duration === "number" ? opts.duration : defaultDuration;

  const toast = document.createElement("div");
  toast.className = "toast toast--" + type;
  toast.setAttribute("role", type === "error" ? "alert" : "status");

  const header = document.createElement("div");
  header.className = "toast-header";

  const icon = document.createElement("div");
  icon.className = "toast-icon";
  const iconNode = cloneIcon(opts.icon || type);
  if (iconNode) {
    icon.appendChild(iconNode);
  }

  const titleEl = document.createElement("div");
  titleEl.className = "toast-title";
  titleEl.textContent = title;

  const messageEl = document.createElement("div");
  messageEl.className = "toast-message";
  messageEl.textContent = message;

  const close = document.createElement("button");
  close.className = "toast-close";
  close.type = "button";
  close.setAttribute("aria-label", "Close");
  const closeIcon = cloneIcon("close");
  if (closeIcon) {
    close.appendChild(closeIcon);
  } else {
    close.textContent = "\u00d7";
  }

  header.appendChild(icon);
  header.appendChild(titleEl);

  toast.appendChild(header);
  toast.appendChild(messageEl);
  toast.appendChild(close);

  let removed = false;
  function removeToast() {
    if (removed) {
      return;
    }
    removed = true;
    if (toast && toast.parentNode) {
      toast.parentNode.removeChild(toast);
    }
  }

  function dismissToast() {
    if (toast.classList.contains("is-leaving")) {
      return;
    }
    toast.classList.add("is-leaving");
    setTimeout(removeToast, 220);
  }

  close.addEventListener("click", dismissToast);

  if (duration > 0) {
    const progress = document.createElement("div");
    progress.className = "toast-progress";
    progress.style.setProperty("--toast-duration", duration + "ms");
    toast.appendChild(progress);
    setTimeout(dismissToast, duration);
  }

  return toast;
}

function showToast(message, options) {
  if (!message) {
    return;
  }
  const toast = buildToast(message, options);
  toastHost.prepend(toast);
}

window.Toast = {
  show: showToast,
  success(message, options) {
    showToast(message, Object.assign({}, options, { type: "success" }));
  },
  error(message, options) {
    showToast(message, Object.assign({}, options, { type: "error" }));
  },
  warning(message, options) {
    showToast(message, Object.assign({}, options, { type: "warning" }));
  },
  info(message, options) {
    showToast(message, Object.assign({}, options, { type: "info" }));
  }
};
