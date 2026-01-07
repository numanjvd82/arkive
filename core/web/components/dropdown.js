(function() {
  const dropdowns = document.querySelectorAll("[data-dropdown]");
  if (!dropdowns.length) {
    return;
  }

  dropdowns.forEach((dropdown) => {
    const trigger = dropdown.querySelector("[data-dropdown-trigger]");
    const menu = dropdown.querySelector("[data-dropdown-menu]");
    if (!trigger || !menu) {
      return;
    }

    const close = () => {
      dropdown.classList.remove("is-open");
      trigger.setAttribute("aria-expanded", "false");
    };

    trigger.addEventListener("click", (event) => {
      event.stopPropagation();
      const isOpen = dropdown.classList.toggle("is-open");
      trigger.setAttribute("aria-expanded", isOpen ? "true" : "false");
    });

    menu.addEventListener("click", (event) => {
      const target = event.target;
      if (!target || !target.closest) {
        return;
      }
      if (target.closest("a, button")) {
        close();
      }
    });

    document.addEventListener("click", (event) => {
      if (!dropdown.contains(event.target)) {
        close();
      }
    });

    document.addEventListener("keydown", (event) => {
      if (event.key === "Escape") {
        close();
      }
    });
  });
})();
