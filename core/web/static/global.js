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
