(function () {
  const backdrop = document.getElementById("lightbox-backdrop");
  const imageEl = document.getElementById("lightbox-image");
  const titleEl = document.getElementById("lightbox-title");
  const closeButton = document.querySelector(".lightbox-close");
  const zoomInButton = document.querySelector("[data-lightbox-action='zoom-in']");
  const zoomOutButton = document.querySelector("[data-lightbox-action='zoom-out']");
  const stage = document.querySelector(".lightbox-stage");
  const actionButtons = document.querySelectorAll("[data-lightbox-action]");
  const triggerImages = document.querySelectorAll("[data-lightbox-src]");

  if (!backdrop || !imageEl || !stage || !triggerImages.length) return;

  let zoom = 1;
  let offsetX = 0;
  let offsetY = 0;
  let raf = null;
  const minZoom = 0.5;
  const maxZoom = 3;

  // Interaction state
  let interacting = false;
  let settleTimer = null;

  // For pinch/pan
  const pointers = new Map();

  // Full-res swap
  let fullSrc = "";
  let previewSrc = "";

  function isOpen() {
    return !backdrop.classList.contains("is-hidden");
  }

  function clamp(v, min, max) {
    return Math.max(min, Math.min(max, v));
  }

  function applyTransform() {
    imageEl.style.transform = `translate3d(${offsetX}px, ${offsetY}px, 0) scale(${zoom})`;
  }

  function updateZoomButtons() {
    if (zoomInButton) {
      zoomInButton.disabled = zoom >= maxZoom;
      zoomInButton.classList.toggle("is-disabled", zoom >= maxZoom);
    }
    if (zoomOutButton) {
      zoomOutButton.disabled = zoom <= minZoom;
      zoomOutButton.classList.toggle("is-disabled", zoom <= minZoom);
    }
  }

  function requestTransform() {
    if (raf) return;
    raf = requestAnimationFrame(() => {
      raf = null;
      applyTransform();
      updateZoomButtons();
    });
  }

  function resetTransform() {
    zoom = 1;
    offsetX = 0;
    offsetY = 0;
    requestTransform();
  }

  let prevHtmlOverflow = "";
  let prevBodyOverflow = "";

  function disableScroll() {
    prevHtmlOverflow = document.documentElement.style.overflow || "";
    prevBodyOverflow = document.body.style.overflow || "";
    document.documentElement.style.overflow = "hidden";
    document.body.style.overflow = "hidden";
  }

  function enableScroll() {
    document.documentElement.style.overflow = prevHtmlOverflow;
    document.body.style.overflow = prevBodyOverflow;
  }

  function setInteractiveQuality(on) {
    // Fast mode while interacting
    imageEl.style.imageRendering = on ? "pixelated" : "auto";
  }

  function beginInteraction() {
    interacting = true;
    setInteractiveQuality(true);
    stage.classList.add("is-grabbing");
    clearTimeout(settleTimer);
  }

  function endInteraction() {
    interacting = false;
    stage.classList.remove("is-grabbing");

    // After user stops, restore quality and optionally swap to full-res
    clearTimeout(settleTimer);
    settleTimer = setTimeout(() => {
      setInteractiveQuality(false);
      maybeSwapFullRes();
    }, 120);
  }

  // Downscale large images for preview (super important)
  async function makePreview(src, maxDim = 2000) {
    try {
      const img = new Image();
      img.decoding = "async";
      img.src = src;
      await img.decode();

      const scale = Math.min(1, maxDim / Math.max(img.width, img.height));
      if (scale === 1) return src;

      const canvas = document.createElement("canvas");
      canvas.width = Math.round(img.width * scale);
      canvas.height = Math.round(img.height * scale);

      const ctx = canvas.getContext("2d", { alpha: false });
      ctx.drawImage(img, 0, 0, canvas.width, canvas.height);

      return canvas.toDataURL("image/jpeg", 0.9);
    } catch (e) {
      return src; // fallback
    }
  }

  function maybeSwapFullRes() {
    // If zoomed in enough and not already on full-res, swap it
    if (!fullSrc) return;
    if (zoom < 1.4) return;

    if (imageEl.src !== fullSrc) {
      const prev = imageEl.src;

      // keep current transform, swap source
      imageEl.src = fullSrc;

      // if full-res fails to load, fallback
      imageEl.onerror = () => {
        imageEl.src = prev;
        imageEl.onerror = null;
      };
    }
  }

  async function openLightbox(src, title) {
    if (!src) return;

    fullSrc = src;
    previewSrc = "";
    imageEl.onerror = null;

    imageEl.src = "";
    imageEl.alt = title || "";
    if (titleEl) titleEl.textContent = title || "Preview";

    resetTransform();
    backdrop.classList.remove("is-hidden");
    backdrop.setAttribute("aria-hidden", "false");
    disableScroll();

    // Load preview first to prevent freezing
    previewSrc = await makePreview(src, 2000);
    imageEl.src = previewSrc;
    imageEl.decoding = "async";
    updateZoomButtons();
  }

  function closeLightbox() {
    backdrop.classList.add("is-hidden");
    backdrop.setAttribute("aria-hidden", "true");
    imageEl.src = "";
    enableScroll();
    resetTransform();
    pointers.clear();
    endInteraction();
    updateZoomButtons();
  }

  function zoomTo(newZoom, anchorX, anchorY) {
    // Zoom around a stage point (anchor), adjusting offset so it stays under cursor
    const prevZoom = zoom;
    zoom = clamp(newZoom, minZoom, maxZoom);

    const rect = stage.getBoundingClientRect();
    const cx = anchorX - rect.left - rect.width / 2;
    const cy = anchorY - rect.top - rect.height / 2;

    const scale = zoom / prevZoom;
    offsetX = offsetX * scale + cx * (1 - scale);
    offsetY = offsetY * scale + cy * (1 - scale);

    requestTransform();
  }

  // Events: open/close
  triggerImages.forEach((img) => {
    img.addEventListener("click", () => {
      openLightbox(img.getAttribute("data-lightbox-src"), img.getAttribute("data-lightbox-title"));
    });
  });

  if (closeButton) closeButton.addEventListener("click", closeLightbox);

  backdrop.addEventListener("click", (event) => {
    if (event.target === backdrop) closeLightbox();
  });

  document.addEventListener("keydown", (event) => {
    if (event.key === "Escape") closeLightbox();
  });

  // Buttons
  actionButtons.forEach((button) => {
    button.addEventListener("click", () => {
      if (!isOpen()) return;
      const action = button.getAttribute("data-lightbox-action");
      if (action === "fit") resetTransform();
      else if (action === "zoom-in") zoomTo(zoom + 0.2, stage.clientWidth / 2, stage.clientHeight / 2);
      else if (action === "zoom-out") zoomTo(zoom - 0.2, stage.clientWidth / 2, stage.clientHeight / 2);
    });
  });

  imageEl.addEventListener("load", () => {
    requestTransform();
  });

  // Pointer Events (stable pan + pinch)
  function dist(a, b) {
    return Math.hypot(a.x - b.x, a.y - b.y);
  }

  function mid(a, b) {
    return { x: (a.x + b.x) / 2, y: (a.y + b.y) / 2 };
  }

  let panStart = { x: 0, y: 0 };
  let panOffsetStart = { x: 0, y: 0 };
  let pinchStartDist = 0;
  let pinchStartZoom = 1;
  let pinchStartOffset = { x: 0, y: 0 };
  let pinchStartMid = { x: 0, y: 0 };

  stage.addEventListener("pointerdown", (e) => {
    if (!isOpen() || !imageEl.src) return;
    e.preventDefault();
    if (e.pointerType === "mouse" && e.button !== 0) return;
    stage.setPointerCapture(e.pointerId);
    pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
    beginInteraction();

    if (pointers.size === 1) {
      const p = pointers.values().next().value;
      panStart = { x: p.x, y: p.y };
      panOffsetStart = { x: offsetX, y: offsetY };
    } else if (pointers.size === 2) {
      const pts = Array.from(pointers.values());
      pinchStartDist = dist(pts[0], pts[1]);
      pinchStartMid = mid(pts[0], pts[1]);
      pinchStartZoom = zoom;
      pinchStartOffset = { x: offsetX, y: offsetY };
    }
  }, { passive: false });

  stage.addEventListener("pointermove", (e) => {
    if (!pointers.has(e.pointerId) || !isOpen()) return;
    e.preventDefault();
    const p = pointers.get(e.pointerId);
    p.x = e.clientX;
    p.y = e.clientY;

    if (pointers.size === 1) {
      const p1 = pointers.values().next().value;
      offsetX = panOffsetStart.x + (p1.x - panStart.x);
      offsetY = panOffsetStart.y + (p1.y - panStart.y);
      requestTransform();
      return;
    }

    if (pointers.size === 2) {
      const pts = Array.from(pointers.values());
      const curDist = dist(pts[0], pts[1]);
      const curMid = mid(pts[0], pts[1]);
      const scale = curDist / pinchStartDist;
      const nextZoom = clamp(pinchStartZoom * scale, minZoom, maxZoom);
      const dx = curMid.x - pinchStartMid.x;
      const dy = curMid.y - pinchStartMid.y;
      const prevZoom = zoom;
      zoom = nextZoom;
      const rect = stage.getBoundingClientRect();
      const cx = curMid.x - rect.left - rect.width / 2;
      const cy = curMid.y - rect.top - rect.height / 2;
      const zScale = zoom / prevZoom;
      offsetX = (pinchStartOffset.x + dx) * zScale + cx * (1 - zScale);
      offsetY = (pinchStartOffset.y + dy) * zScale + cy * (1 - zScale);
      requestTransform();
    }
  }, { passive: false });

  function pointerEnd(e) {
    if (!pointers.has(e.pointerId)) return;
    e.preventDefault();
    pointers.delete(e.pointerId);

    if (pointers.size === 0) {
      endInteraction();
      return;
    }

    if (pointers.size === 1) {
      const p1 = pointers.values().next().value;
      panStart = { x: p1.x, y: p1.y };
      panOffsetStart = { x: offsetX, y: offsetY };
    }
  }

  stage.addEventListener("pointerup", pointerEnd, { passive: false });
  stage.addEventListener("pointercancel", pointerEnd, { passive: false });
  stage.addEventListener("lostpointercapture", pointerEnd, { passive: false });

  // Wheel zoom (throttled and anchored)
  stage.addEventListener(
    "wheel",
    (e) => {
      if (!isOpen() || !imageEl.src) return;
      e.preventDefault();
      beginInteraction();

      const delta = -e.deltaY * 0.001;
      zoomTo(zoom + delta, e.clientX, e.clientY);

      endInteraction();
    },
    { passive: false }
  );
})();
