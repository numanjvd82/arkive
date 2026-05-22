import { UploadRunner } from "../upload_runner.js";

let service = null;

export function getUploadService(options) {
  if (!service) {
    service = new UploadRunner(options || {});
    return service;
  }

  if (options && options.limits) {
    service.limits = options.limits;
  }

  return service;
}

export const uploadService = getUploadService();
