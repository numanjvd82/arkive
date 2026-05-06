export function initToast() {
  if (window.Toast && window.Toast.__arkiveReady) {
    return;
  }

  const toastHost = document.getElementById("toast-host");
  if (!toastHost) {
    return;
  }

  const defaultDuration = 4200;
  const defaultTitles = {
    info: "Heads up",
    success: "Done",
    warning: "Check this",
    error: "Something went wrong",
    processing: "Processing"
  };
  const defaultCodes = {
    success: "200 OK",
    error: "ERR",
    warning: "WARN",
    info: "INFO",
    processing: "PROC"
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
    const code = opts.code || defaultCodes[type] || "";
    const dismissLabel = opts.dismissLabel || "Dismiss";
    const processing = type === "processing";

    const toast = document.createElement("div");
    toast.className = "toast toast--" + type;
    toast.setAttribute("role", type === "error" ? "alert" : "status");

    const meta = document.createElement("div");
    meta.className = "toast-meta";

    const labelEl = document.createElement("span");
    labelEl.className = "toast-label";
    labelEl.textContent = (opts.label || type).toUpperCase();

    meta.appendChild(labelEl);

    if (code) {
      const codeEl = document.createElement("span");
      codeEl.className = "toast-code mono";
      codeEl.textContent = code;
      meta.appendChild(codeEl);
    }

    const body = document.createElement("div");
    body.className = "toast-body";

    const icon = document.createElement("div");
    icon.className = "toast-icon" + (processing ? " is-processing" : "");
    const iconNode = cloneIcon(opts.icon || (processing ? "processing" : type));
    if (iconNode) {
      icon.appendChild(iconNode);
    }

    const copy = document.createElement("div");
    copy.className = "toast-copy";

    if (title) {
      const titleEl = document.createElement("div");
      titleEl.className = "toast-title";
      titleEl.textContent = title;
      copy.appendChild(titleEl);
    }

    if (message) {
      const messageEl = document.createElement("div");
      messageEl.className = "toast-message";
      messageEl.textContent = message;
      copy.appendChild(messageEl);
    }

    body.appendChild(icon);
    body.appendChild(copy);

    toast.appendChild(meta);
    toast.appendChild(body);

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

    const actions = Array.isArray(opts.actions) ? opts.actions.slice() : [];
    const wantsDismiss = opts.dismissible !== false && !processing;
    if (wantsDismiss) {
      actions.unshift({
        label: dismissLabel,
        variant: "secondary",
        onClick: dismissToast
      });
    }

    if (actions.length > 0) {
      const actionsEl = document.createElement("div");
      actionsEl.className = "toast-actions";

      actions.forEach((action) => {
        if (!action || !action.label) {
          return;
        }
        const button = document.createElement("button");
        button.className = "toast-action toast-action--" + (action.variant || "secondary");
        button.type = "button";
        button.textContent = action.label;
        button.addEventListener("click", function() {
          if (typeof action.onClick === "function") {
            action.onClick({ dismiss: dismissToast, toast: toast });
          }
          if (action.dismiss !== false && action.onClick !== dismissToast) {
            dismissToast();
          }
        });
        actionsEl.appendChild(button);
      });

      if (actionsEl.childNodes.length > 0) {
        toast.appendChild(actionsEl);
      }
    }

    if (typeof opts.progress === "number") {
      const progress = document.createElement("div");
      const fill = document.createElement("div");
      const pct = Math.max(0, Math.min(100, opts.progress));
      progress.className = "toast-progress";
      fill.className = "toast-progress-fill";
      fill.style.width = pct + "%";
      progress.appendChild(fill);
      toast.appendChild(progress);
    } else if (duration > 0) {
      const progress = document.createElement("div");
      progress.className = "toast-progress";
      progress.style.setProperty("--toast-duration", duration + "ms");
      const fill = document.createElement("div");
      fill.className = "toast-progress-fill is-timed";
      progress.appendChild(fill);
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
    __arkiveReady: true,
    show: showToast,
    success(message, options) {
      showToast(message, Object.assign({}, options, { type: "success" }));
    },
    error(message, options) {
      if (window.RateLimit && window.RateLimit.isActive && window.RateLimit.isActive() && !(options && options.allowRateLimit)) {
        return;
      }
      showToast(message, Object.assign({}, options, { type: "error" }));
    },
    warning(message, options) {
      showToast(message, Object.assign({}, options, { type: "warning" }));
    },
    info(message, options) {
      showToast(message, Object.assign({}, options, { type: "info" }));
    },
    processing(message, options) {
      showToast(message, Object.assign({}, options, { type: "processing" }));
    }
  };
}
