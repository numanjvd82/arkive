(function() {
  var modes = ["dark", "system", "light"];
  var stored = localStorage.getItem("theme");
  var current = modes.indexOf(stored) !== -1 ? stored : "system";
  var root = document.documentElement;
  var prefersLight = window.matchMedia && window.matchMedia("(prefers-color-scheme: light)");

  function updateToggle(mode) {
    var toggle = document.getElementById("theme-toggle");
    if (!toggle) {
      return;
    }
    var label = toggle.querySelector(".theme-label");
    if (label) {
      label.textContent = mode;
    } else {
      toggle.textContent = mode;
    }
    toggle.setAttribute("aria-label", "Theme: " + mode);
  }

  function resolvedTheme(mode) {
    if (mode !== "system") {
      return mode;
    }
    return prefersLight && prefersLight.matches ? "light" : "dark";
  }

  function setMode(mode) {
    root.setAttribute("data-theme", resolvedTheme(mode));
    root.setAttribute("data-theme-mode", mode);
    localStorage.setItem("theme", mode);
    updateToggle(mode);
  }

  setMode(current);

  if (prefersLight && prefersLight.addEventListener) {
    prefersLight.addEventListener("change", function() {
      if ((root.getAttribute("data-theme-mode") || "system") === "system") {
        setMode("system");
      }
    });
  }

  var toggle = document.getElementById("theme-toggle");
  if (!toggle) {
    return;
  }
  toggle.addEventListener("click", function() {
    var active = root.getAttribute("data-theme-mode") || "system";
    var index = modes.indexOf(active);
    var next = modes[(index + 1) % modes.length];
    setMode(next);
  });
})();

(function() {
  var toggles = document.querySelectorAll(".password-toggle");
  if (!toggles.length) {
    return;
  }

  toggles.forEach(function(toggle) {
    toggle.addEventListener("click", function() {
      var targetId = toggle.getAttribute("data-target");
      if (!targetId) {
        return;
      }
      var input = document.getElementById(targetId);
      if (!input) {
        return;
      }
      var visible = input.getAttribute("type") === "text";
      var nextVisible = !visible;
      input.setAttribute("type", nextVisible ? "text" : "password");
      toggle.setAttribute("aria-pressed", nextVisible ? "true" : "false");
      toggle.setAttribute("data-visible", nextVisible ? "true" : "false");
      toggle.setAttribute("aria-label", nextVisible ? "Hide password" : "Show password");
    });
  });
})();

(function() {
  if (!window.fetch) {
    return;
  }
  var originalFetch = window.fetch;
  var lastToastAt = 0;
  var rateLimitState = window.RateLimit || {};
  rateLimitState.until = rateLimitState.until || 0;
  rateLimitState.isActive = function() {
    return Date.now() < rateLimitState.until;
  };
  rateLimitState.setLimited = function(seconds) {
    var waitSeconds = typeof seconds === "number" && seconds > 0 ? seconds : 60;
    var now = Date.now();
    rateLimitState.until = Math.max(rateLimitState.until, now + waitSeconds * 1000);
  };
  window.RateLimit = rateLimitState;

  window.fetch = async function() {
    const res = await originalFetch.apply(this, arguments);
    if (!res || res.status !== 429) {
      return res;
    }
    var retryHeader = res.headers ? res.headers.get("Retry-After") : null;
    var limitedHeader = res.headers ? res.headers.get("X-Rate-Limited") : null;
    if (!retryHeader && !limitedHeader) {
      return res;
    }
    var retrySeconds = retryHeader ? parseInt(retryHeader, 10) : NaN;
    if (!isFinite(retrySeconds) || retrySeconds <= 0) {
      retrySeconds = 60;
    }
    rateLimitState.setLimited(retrySeconds);
    var now = Date.now();
    if (window.Toast && now - lastToastAt > 30000) {
      window.Toast.warning("Too many requests. Try again in a minute.", { title: "Slow down" });
      lastToastAt = now;
    }
    var err = new Error("Rate limited");
    err.status = 429;
    err.rateLimited = true;
    err.retryAfter = retrySeconds;
    throw err;
  };
})();
