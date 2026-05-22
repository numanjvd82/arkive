import { setButtonBusy } from "./button_state.js";

async function decryptFolderItem(item) {
  if (!window.ArkiveVault || typeof window.ArkiveVault.decryptFolderMetadata !== "function") {
    return;
  }
  const metaValue = String(item.getAttribute("data-folder-meta-b64") || "");
  const nameValue = String(item.getAttribute("data-folder-name-b64") || "");
  if (!metaValue && !nameValue) {
    return;
  }
  try {
    let name = "";
    let metadata = null;
    if (metaValue) {
      const result = await window.ArkiveVault.decryptFolderMetadata(metaValue);
      metadata = result && result.metadata ? result.metadata : null;
      name = typeof metadata === "string" ? metadata : String((metadata && metadata.name) || "");
    }
    if (!name && nameValue && typeof window.ArkiveVault.decryptFolderName === "function") {
      const result = await window.ArkiveVault.decryptFolderName(nameValue);
      const fallback = result && result.metadata ? result.metadata : null;
      if (!metadata && fallback && typeof fallback === "object") {
        metadata = fallback;
      }
      name = typeof fallback === "string" ? fallback : String((fallback && fallback.name) || "");
    }
    if (!name) {
      return;
    }
    item.querySelectorAll("[data-folder-field='name']").forEach(function(node) {
      node.textContent = name;
      node.removeAttribute("aria-hidden");
    });
    item.setAttribute("data-folder-name", name);
    try {
      item.setAttribute("data-folder-metadata-json", JSON.stringify(metadata && typeof metadata === "object" ? metadata : { name: name }));
    } catch (_) {}
    item.classList.add("is-hydrated");
    return name;
  } catch (_) {}
  return "";
}

function currentFolderId() {
  const current = document.querySelector("[data-current-folder-id]");
  if (!current) {
    return null;
  }
  const value = String(current.getAttribute("data-current-folder-id") || "").trim();
  return value || null;
}

function updateUploadLabel(name) {
  const label = document.querySelector("[data-upload-label='true']");
  if (!label) {
    return;
  }
  if (!currentFolderId()) {
    label.textContent = "Upload";
    return;
  }
  const value = String(name || "").trim();
  label.textContent = value ? "Upload to " + value : "Upload here";
}

export function initFolders() {
  const folderItems = Array.from(document.querySelectorAll("[data-folder-item]"));
  const folderBreadcrumbs = Array.from(document.querySelectorAll("[data-folder-breadcrumb]"));
  const newFolderButton = document.getElementById("new-folder-button");
  const confirmButton = document.getElementById("folder-create-confirm");
  const cancelButton = document.getElementById("folder-create-cancel");
  const nameInput = document.getElementById("folder-name-input");

  folderItems.forEach(function(item) {
    if (!item.hasAttribute("data-folder-open-bound")) {
      item.setAttribute("data-folder-open-bound", "true");
      item.addEventListener("dblclick", function(event) {
        if (event.target.closest("a, button, input, label")) {
          return;
        }
        const href = String(item.getAttribute("data-folder-open") || "");
        if (href) {
          window.location.href = href;
        }
      });
      item.addEventListener("keydown", function(event) {
        if (event.key !== "Enter") {
          return;
        }
        const href = String(item.getAttribute("data-folder-open") || "");
        if (href) {
          window.location.href = href;
        }
      });
    }
  });

  if (window.ArkiveVault && typeof window.ArkiveVault.waitUntilReady === "function") {
    window.ArkiveVault.waitUntilReady().then(function() {
      const itemJobs = folderItems.map(function(item) {
        return decryptFolderItem(item);
      });
      const breadcrumbJobs = folderBreadcrumbs.map(function(item) {
        return decryptFolderItem(item).then(function(name) {
          if (item.classList.contains("files-location-current")) {
            updateUploadLabel(name);
          }
          return name;
        });
      });
      return Promise.all(itemJobs.concat(breadcrumbJobs)).then(function() {
        if (currentFolderId() && !document.querySelector(".files-location-current[data-folder-name]")) {
          updateUploadLabel("");
        }
      });
    }).catch(function() {});
  } else {
    updateUploadLabel("");
  }

  if (newFolderButton && !newFolderButton.hasAttribute("data-folder-create-bound")) {
    newFolderButton.setAttribute("data-folder-create-bound", "true");
    newFolderButton.addEventListener("click", function() {
      if (window.Dialog && window.Dialog.open) {
        window.Dialog.open("folder-create-backdrop");
      }
      if (nameInput) {
        nameInput.value = "";
        nameInput.focus();
      }
    });
  }

  if (cancelButton && !cancelButton.hasAttribute("data-folder-cancel-bound")) {
    cancelButton.setAttribute("data-folder-cancel-bound", "true");
    cancelButton.addEventListener("click", function() {
      if (window.Dialog && window.Dialog.close) {
        window.Dialog.close("folder-create-backdrop");
      }
    });
  }

  if (confirmButton && !confirmButton.hasAttribute("data-folder-confirm-bound")) {
    confirmButton.setAttribute("data-folder-confirm-bound", "true");
    confirmButton.addEventListener("click", async function() {
      const name = String(nameInput && nameInput.value || "").trim();
      if (!name || !window.ArkiveVault || typeof window.ArkiveVault.encryptFolderName !== "function" || typeof window.ArkiveVault.encryptFolderMetadata !== "function") {
        return;
      }
      try {
        setButtonBusy(confirmButton, true, { busyText: "Creating..." });
        const encryptedName = await window.ArkiveVault.encryptFolderName({ name: name });
        const encryptedMetadata = await window.ArkiveVault.encryptFolderMetadata({ name: name });
        await window.ArkiveAPI.apiRequest("/api/folders", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
          body: JSON.stringify({
            parentFolderId: currentFolderId(),
            encryptedName: encryptedName && encryptedName.encryptedMetadata ? encryptedName.encryptedMetadata : "",
            encryptedMetadata: encryptedMetadata && encryptedMetadata.encryptedMetadata ? encryptedMetadata.encryptedMetadata : ""
          })
        }, {
          code: "validation_failed",
          message: "Create folder failed"
        });
        window.location.reload();
      } catch (error) {
        if (window.ArkiveUI && typeof window.ArkiveUI.showAppError === "function") {
          window.ArkiveUI.showAppError(error, {
            code: "validation_failed",
            message: "Create folder failed"
          });
        }
      } finally {
        setButtonBusy(confirmButton, false);
      }
    });
  }
}
