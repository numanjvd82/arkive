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
    if (metaValue) {
      const result = await window.ArkiveVault.decryptFolderMetadata(metaValue);
      const metadata = result && result.metadata ? result.metadata : {};
      name = typeof metadata === "string" ? metadata : String(metadata.name || "");
    }
    if (!name && nameValue && typeof window.ArkiveVault.decryptFolderName === "function") {
      const result = await window.ArkiveVault.decryptFolderName(nameValue);
      const metadata = result && result.metadata ? result.metadata : {};
      name = typeof metadata === "string" ? metadata : String(metadata.name || "");
    }
    if (!name) {
      return;
    }
    item.querySelectorAll("[data-folder-field='name']").forEach(function(node) {
      node.textContent = name;
    });
    if (item.hasAttribute("data-folder-breadcrumb")) {
      item.textContent = name;
    }
    item.setAttribute("data-folder-name", name);
  } catch (_) {}
}

function currentFolderId() {
  const current = document.querySelector("[data-current-folder-id]");
  if (!current) {
    return null;
  }
  const value = String(current.getAttribute("data-current-folder-id") || "").trim();
  return value || null;
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
      folderItems.forEach(function(item) {
        void decryptFolderItem(item);
      });
      folderBreadcrumbs.forEach(function(item) {
        void decryptFolderItem(item);
      });
    }).catch(function() {});
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
        const response = await fetch("/api/folders", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          credentials: "include",
          body: JSON.stringify({
            parentFolderId: currentFolderId(),
            encryptedName: encryptedName && encryptedName.encryptedMetadata ? encryptedName.encryptedMetadata : "",
            encryptedMetadata: encryptedMetadata && encryptedMetadata.encryptedMetadata ? encryptedMetadata.encryptedMetadata : ""
          })
        });
        if (!response.ok) {
          throw new Error("Create folder failed");
        }
        window.location.reload();
      } catch (error) {
        if (window.Toast) {
          window.Toast.error((error && error.message) || "Create folder failed.");
        }
      } finally {
        setButtonBusy(confirmButton, false);
      }
    });
  }
}
