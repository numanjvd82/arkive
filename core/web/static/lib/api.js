import { AppError, defaultCodeForStatus, firstValidationMessage, toAppError } from "./errors.js";

export const APIError = AppError;

export function parseAPIErrorPayload(data, fallback, status) {
  const baseFallback = String(fallback || "Request failed");
  if (data && typeof data === "object") {
    const payloadError = data.error;
    if (payloadError && typeof payloadError === "object" && !Array.isArray(payloadError)) {
      return new AppError(payloadError.message || baseFallback, {
        code: payloadError.code || defaultCodeForStatus(status),
        details: payloadError.details || null,
        status: status
      });
    }
    if (data.errors && typeof data.errors === "object") {
      return new AppError(firstValidationMessage(data.errors) || baseFallback, {
        code: "validation_failed",
        details: { fields: data.errors },
        status: status
      });
    }
    if (typeof payloadError === "string" && payloadError) {
      return new AppError(payloadError, {
        code: defaultCodeForStatus(status),
        status: status
      });
    }
  }
  return new AppError(baseFallback, {
    code: defaultCodeForStatus(status),
    status: status
  });
}

async function parseJSONBody(response) {
  const text = await response.text();
  let data = null;
  try {
    data = text ? JSON.parse(text) : null;
  } catch (_) {}
  return { text: text, data: data };
}

export async function apiRequest(url, options, fallback) {
  let response;
  try {
    response = await fetch(url, options || {});
  } catch (error) {
    if (error && error.name === "AbortError") {
      throw error;
    }
    throw toAppError(error, {
      code: "network_error",
      message: fallback && fallback.message ? fallback.message : "Network error",
    });
  }
  const parsed = await parseJSONBody(response);
  if (!response.ok) {
    throw parseAPIErrorPayload(parsed.data, parsed.text || (fallback && fallback.message) || "Request failed", response.status);
  }
  return parsed.data;
}

export async function readJSON(response, fallback) {
  const parsed = await parseJSONBody(response);
  if (!response.ok) {
    throw parseAPIErrorPayload(parsed.data, parsed.text || fallback, response.status);
  }
  return parsed.data;
}

export function installGlobalAPI() {
  window.ArkiveAPI = {
    APIError,
    AppError,
    apiRequest,
    firstValidationMessage,
    parseAPIErrorPayload,
    readJSON,
    toAppError,
  };
}
