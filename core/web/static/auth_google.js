(function () {
  const container = document.querySelector(".auth-oauth[data-google-client-id]");
  if (!container) {
    return;
  }

  const clientId = container.getAttribute("data-google-client-id");
  const loginEndpoint = container.getAttribute("data-google-login-endpoint") || "/auth/google";
  const button = container.querySelector("#google-signin-button");
  if (!clientId || !button) {
    window.Toast.error("Google sign-in is unavailable.", { title: "Sign-in failed" });
    return;
  }

  function handleCredentialResponse(response) {
    fetch(loginEndpoint, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ credential: response.credential }),
    })
      .then(async function (res) {
        if (!res.ok) {
          const data = await res.json();
          throw new Error((data && data.error) || "Google sign-in failed");
        }
        return res.json();
      })
      .then(function () {
        window.location.assign("/dashboard");
      })
      .catch(function (err) {
        console.error(err);
        window.Toast.error(err.message, { title: "Sign-in failed" });
      });
  }

  function renderButton() {
    if (!window.google || !window.google.accounts || !window.google.accounts.id) {
      window.Toast.error("Google sign-in is unavailable.", { title: "Sign-in failed" });
      return;
    }

    window.google.accounts.id.initialize({
      client_id: clientId,
      callback: handleCredentialResponse,
    });

    window.google.accounts.id.renderButton(button, {
      type: "standard",
      theme: "outline",
      size: "large",
      text: "continue_with",
      shape: "rectangular",
    });
  }

  if (window.google && window.google.accounts && window.google.accounts.id) {
    renderButton();
    return;
  }

  const script = document.createElement("script");
  script.src = "https://accounts.google.com/gsi/client";
  script.async = true;
  script.defer = true;
  script.onload = renderButton;
  script.onerror = function () {
    window.Toast.error("Google sign-in failed to load.", { title: "Sign-in failed" });
  };
  document.head.appendChild(script);
})();
