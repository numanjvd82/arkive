(function() {
  var stored = localStorage.getItem("theme");
  var current = stored || "dark";
  document.documentElement.setAttribute("data-theme", current);
  var toggle = document.getElementById("theme-toggle");
  if (!toggle) {
    return;
  }
  toggle.setAttribute("aria-pressed", current === "dark" ? "true" : "false");
  toggle.addEventListener("click", function() {
    var active = document.documentElement.getAttribute("data-theme") || "dark";
    var next = active === "dark" ? "light" : "dark";
    document.documentElement.setAttribute("data-theme", next);
    localStorage.setItem("theme", next);
    toggle.setAttribute("aria-pressed", next === "dark" ? "true" : "false");
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
