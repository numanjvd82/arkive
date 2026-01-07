(function() {
  const body = document.body;
  const toggle = document.getElementById("sidebar-toggle");
  const sidebar = document.getElementById("dashboard-sidebar");
  const scrim = document.querySelector(".sidebar-scrim");
  const closeBtn = document.querySelector(".sidebar-close");
  const links = Array.from(document.querySelectorAll(".sidebar-link"));

  if (!toggle || !sidebar || !scrim) {
    return;
  }

  const normalizePath = function(pathname) {
    if (!pathname) {
      return "/";
    }
    const trimmed = pathname.replace(/\/$/, "");
    return trimmed === "" ? "/" : trimmed;
  };

  const currentPath = normalizePath(window.location.pathname);
  let activeLink = null;
  let activeLength = -1;

  links.forEach(function(link) {
    const linkPath = normalizePath(new URL(link.getAttribute("href"), window.location.origin).pathname);
    const matches = currentPath === linkPath || (linkPath !== "/" && currentPath.indexOf(linkPath + "/") === 0);
    if (matches && linkPath.length > activeLength) {
      activeLink = link;
      activeLength = linkPath.length;
    }
  });

  if (activeLink) {
    links.forEach(function(link) {
      link.classList.remove("is-active");
      link.removeAttribute("aria-current");
    });
    activeLink.classList.add("is-active");
    activeLink.setAttribute("aria-current", "page");
  }

  const setState = function(isOpen) {
    body.classList.toggle("sidebar-open", isOpen);
    toggle.setAttribute("aria-expanded", isOpen ? "true" : "false");
    sidebar.setAttribute("aria-hidden", isOpen ? "false" : "true");
  };

  const closeSidebar = function() {
    setState(false);
  };

  toggle.addEventListener("click", function() {
    const isOpen = body.classList.contains("sidebar-open");
    setState(!isOpen);
  });

  scrim.addEventListener("click", closeSidebar);

  if (closeBtn) {
    closeBtn.addEventListener("click", closeSidebar);
  }

  sidebar.addEventListener("click", function(event) {
    const target = event.target;
    if (!target || !target.closest) {
      return;
    }
    const link = target.closest("a");
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
