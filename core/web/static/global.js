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
