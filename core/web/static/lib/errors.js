export class AppError extends Error {
  constructor(message, options) {
    const opts = options || {};
    super(String(message || "Something went wrong"));
    this.name = String(opts.name || "AppError");
    this.code = String(opts.code || "unknown_error");
    this.details = opts.details || null;
    this.status = Number.isFinite(opts.status) ? opts.status : 0;
  }
}

export function defaultCodeForStatus(status) {
  switch (Number(status || 0)) {
    case 400:
      return "validation_failed";
    case 401:
      return "unauthorized";
    case 403:
      return "forbidden";
    case 404:
      return "not_found";
    case 409:
      return "conflict";
    case 413:
      return "storage_limit_exceeded";
    default:
      return "unknown_error";
  }
}

export function firstValidationMessage(fields) {
  if (!fields || typeof fields !== "object") {
    return "";
  }
  if (typeof fields.general === "string" && fields.general) {
    return fields.general;
  }
  if (typeof fields._general === "string" && fields._general) {
    return fields._general;
  }
  const keys = Object.keys(fields);
  for (let i = 0; i < keys.length; i++) {
    const value = fields[keys[i]];
    if (typeof value === "string" && value) {
      return value;
    }
  }
  return "";
}

export function toAppError(error, fallback) {
  const base = fallback || {};
  if (error && typeof error === "object") {
    if (error.name === "AbortError") {
      return error;
    }
    if (typeof error.code === "string" && typeof error.message === "string") {
      return new AppError(error.message, {
        name: error.name || "AppError",
        code: error.code,
        details: error.details || null,
        status: Number(error.status || 0),
      });
    }
    if (typeof error.message === "string" && error.message) {
      return new AppError(error.message, {
        name: error.name || "AppError",
        code: base.code || defaultCodeForStatus(error.status),
        details: error.details || null,
        status: Number(error.status || 0),
      });
    }
  }
  return new AppError(base.message || "Something went wrong", {
    code: base.code || "unknown_error",
    details: error ? { cause: error } : null,
    status: Number(error && error.status || 0),
  });
}
