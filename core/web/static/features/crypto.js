let cryptoReadyPromise = null;

export function getArkiveCrypto() {
  if (cryptoReadyPromise) {
    return cryptoReadyPromise;
  }

  cryptoReadyPromise = import("../vendor/arkive-crypto/arkive_crypto.js")
    .then(function(mod) {
      return mod.default("/static/vendor/arkive-crypto/arkive_crypto_bg.wasm")
        .then(function() {
          return mod;
        });
    });

  return cryptoReadyPromise;
}

export function initCrypto() {
  if (window.ArkiveCrypto && typeof window.ArkiveCrypto.ready === "function") {
    return;
  }

  window.ArkiveCrypto = {
    ready: getArkiveCrypto
  };
}
