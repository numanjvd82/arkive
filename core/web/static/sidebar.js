(function() {
  var body = document.body;
  var toggle = document.getElementById("sidebar-toggle");
  var sidebar = document.getElementById("dashboard-sidebar");
  var scrim = document.querySelector(".sidebar-scrim");
  var closeBtn = document.querySelector(".sidebar-close");

  if (!toggle || !sidebar || !scrim) {
    return;
  }

  function setState(isOpen) {
    body.classList.toggle("sidebar-open", isOpen);
    toggle.setAttribute("aria-expanded", isOpen ? "true" : "false");
    sidebar.setAttribute("aria-hidden", isOpen ? "false" : "true");
  }

  function openSidebar() {
    setState(true);
  }

  function closeSidebar() {
    setState(false);
  }

  toggle.addEventListener("click", function() {
    var isOpen = body.classList.contains("sidebar-open");
    setState(!isOpen);
  });

  scrim.addEventListener("click", closeSidebar);

  if (closeBtn) {
    closeBtn.addEventListener("click", closeSidebar);
  }

  sidebar.addEventListener("click", function(event) {
    var target = event.target;
    if (!target || !target.closest) {
      return;
    }
    var link = target.closest("a");
    if (link) {
      closeSidebar();
    }
  });

  document.addEventListener("keydown", function(event) {
    if (event.key === "Escape") {
      closeSidebar();
    }
  });
})();
